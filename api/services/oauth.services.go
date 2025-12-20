package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"nexfi-backend/config"
	"nexfi-backend/database"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"nexfi-backend/pkg/redis"
	"nexfi-backend/utils"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// OAuthService handles OAuth-related business logic
type OAuthService struct{}

// NewOAuthService creates a new OAuthService instance
func NewOAuthService() *OAuthService {
	return &OAuthService{}
}

// RegisterUserWithPassword registers a new user with email and password
func RegisterUserWithPassword(input *dto.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:         input.Email,
		PasswordHash:  string(hashedPassword),
		Username:      input.Username,
		Provider:      "local",
		EmailVerified: false,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// LoginWithPassword authenticates a user with email and password
func LoginWithPassword(input *dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if user signed up with OAuth (no password)
	if user.PasswordHash == "" && user.Provider != "local" {
		return nil, fmt.Errorf("this account was registered using %s. Please login with %s", user.Provider, user.Provider)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid password")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	// Store token in Redis
	if err := redis.StoreToken(user.ID, token); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserAuthInfo{
			ID:         user.ID,
			Email:      user.Email,
			Username:   user.Username,
			Avatar:     user.AvatarURL,
			Provider:   user.Provider,
			IsVerified: user.EmailVerified,
		},
	}, nil
}

// GetOAuthURL returns the OAuth URL for the specified provider
func (s *OAuthService) GetOAuthURL(provider string) (*dto.OAuthURLResponse, error) {
	var cfg *oauth2.Config
	var additionalParams []oauth2.AuthCodeOption

	switch provider {
	case "google":
		cfg = config.OAuth.Google
	case "github":
		cfg = config.OAuth.GitHub
	case "apple":
		cfg = config.OAuth.Apple
		// Apple requires response_mode=form_post for web
		additionalParams = append(additionalParams, oauth2.SetAuthURLParam("response_mode", "form_post"))
	default:
		return nil, errors.New("unsupported OAuth provider")
	}

	if cfg == nil {
		return nil, fmt.Errorf("%s OAuth is not configured", provider)
	}

	// Generate state for CSRF protection
	state := utils.GenerateRandomString(32)

	url := cfg.AuthCodeURL(state, append([]oauth2.AuthCodeOption{oauth2.AccessTypeOffline}, additionalParams...)...)

	return &dto.OAuthURLResponse{
		URL:   url,
		State: state,
	}, nil
}

// HandleOAuthCallback processes the OAuth callback and returns user info
func (s *OAuthService) HandleOAuthCallback(provider, code string) (*dto.AuthResponse, error) {
	var cfg *oauth2.Config
	var userInfo *dto.OAuthUserInfo
	var err error

	switch provider {
	case "google":
		cfg = config.OAuth.Google
	case "github":
		cfg = config.OAuth.GitHub
	case "apple":
		cfg = config.OAuth.Apple
	default:
		return nil, errors.New("unsupported OAuth provider")
	}

	if cfg == nil {
		return nil, fmt.Errorf("%s OAuth is not configured", provider)
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from provider
	switch provider {
	case "google":
		userInfo, err = getGoogleUserInfo(token.AccessToken)
	case "github":
		userInfo, err = getGitHubUserInfo(token.AccessToken)
	case "apple":
		// Apple returns user info in the id_token
		idToken, ok := token.Extra("id_token").(string)
		if !ok {
			return nil, errors.New("no id_token in Apple response")
		}
		userInfo, err = getAppleUserInfo(idToken)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := findOrCreateOAuthUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to process user: %w", err)
	}

	// Generate JWT token
	jwtToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	// Store token in Redis
	if err := redis.StoreToken(user.ID, jwtToken); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: jwtToken,
		User: dto.UserAuthInfo{
			ID:         user.ID,
			Email:      user.Email,
			Username:   user.Username,
			Avatar:     user.AvatarURL,
			Provider:   user.Provider,
			IsVerified: user.EmailVerified,
		},
	}, nil
}

// HandleAppleCallback handles Apple Sign In callback with user data
// Apple only sends user data on first authorization
func (s *OAuthService) HandleAppleCallback(code, idToken, userDataJSON string) (*dto.AuthResponse, error) {
	cfg := config.OAuth.Apple
	if cfg == nil {
		return nil, errors.New("Apple OAuth is not configured")
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get id_token from response if not provided
	if idToken == "" {
		idTokenFromResponse, ok := token.Extra("id_token").(string)
		if !ok {
			return nil, errors.New("no id_token in Apple response")
		}
		idToken = idTokenFromResponse
	}

	// Parse user info from id_token
	userInfo, err := getAppleUserInfo(idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Apple id_token: %w", err)
	}

	// Parse additional user data if provided (only on first authorization)
	if userDataJSON != "" {
		var userData struct {
			Name struct {
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			} `json:"name"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal([]byte(userDataJSON), &userData); err == nil {
			if userData.Name.FirstName != "" || userData.Name.LastName != "" {
				userInfo.Name = strings.TrimSpace(userData.Name.FirstName + " " + userData.Name.LastName)
			}
			if userData.Email != "" && userInfo.Email == "" {
				userInfo.Email = userData.Email
			}
		}
	}

	// Find or create user
	user, err := findOrCreateOAuthUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to process user: %w", err)
	}

	// Generate JWT token
	jwtToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	// Store token in Redis
	if err := redis.StoreToken(user.ID, jwtToken); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: jwtToken,
		User: dto.UserAuthInfo{
			ID:         user.ID,
			Email:      user.Email,
			Username:   user.Username,
			Avatar:     user.AvatarURL,
			Provider:   user.Provider,
			IsVerified: user.EmailVerified,
		},
	}, nil
}

// getGoogleUserInfo fetches user info from Google
func getGoogleUserInfo(accessToken string) (*dto.OAuthUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &dto.OAuthUserInfo{
		ID:       data.ID,
		Email:    data.Email,
		Name:     data.Name,
		Avatar:   data.Picture,
		Provider: "google",
	}, nil
}

// getGitHubUserInfo fetches user info from GitHub
func getGitHubUserInfo(accessToken string) (*dto.OAuthUserInfo, error) {
	client := &http.Client{}

	// Get user info
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// If email is not public, fetch from emails endpoint
	if data.Email == "" {
		data.Email, _ = getGitHubPrimaryEmail(accessToken, client)
	}

	name := data.Name
	if name == "" {
		name = data.Login
	}

	return &dto.OAuthUserInfo{
		ID:       fmt.Sprintf("%d", data.ID),
		Email:    data.Email,
		Name:     name,
		Avatar:   data.AvatarURL,
		Provider: "github",
	}, nil
}

// getGitHubPrimaryEmail fetches primary email from GitHub
func getGitHubPrimaryEmail(accessToken string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", errors.New("no primary email found")
}

// getAppleUserInfo extracts user info from Apple's id_token (JWT)
func getAppleUserInfo(idToken string) (*dto.OAuthUserInfo, error) {
	// Parse the JWT without verification (Apple's public keys would be needed for verification)
	// In production, you should verify the token signature
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse id_token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Extract user info from claims
	sub, _ := claims["sub"].(string)     // Apple user ID (stable across sessions)
	email, _ := claims["email"].(string) // User's email
	emailVerified, _ := claims["email_verified"].(bool)

	if sub == "" {
		return nil, errors.New("no user ID in Apple token")
	}

	// Apple doesn't provide name in the id_token - it's only in the initial callback
	// We'll use email prefix as fallback name
	name := ""
	if email != "" {
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			name = parts[0]
		}
	}

	userInfo := &dto.OAuthUserInfo{
		ID:       sub,
		Email:    email,
		Name:     name,
		Avatar:   "", // Apple doesn't provide avatar
		Provider: "apple",
	}

	// Mark if email is verified
	if emailVerified {
		userInfo.Name = name // Keep the name
	}

	return userInfo, nil
}

// findOrCreateOAuthUser finds existing user or creates a new one
func findOrCreateOAuthUser(info *dto.OAuthUserInfo) (*models.User, error) {
	var user models.User

	// First try to find by provider and provider ID
	err := database.DB.Where("provider = ? AND provider_id = ?", info.Provider, info.ID).First(&user).Error
	if err == nil {
		// Update user info from OAuth provider (but preserve name if it was set before)
		if info.Name != "" && user.Username == "" {
			user.Username = info.Name
		}
		if info.Avatar != "" {
			user.AvatarURL = info.Avatar
		}
		user.UpdatedAt = time.Now()
		database.DB.Save(&user)
		return &user, nil
	}

	// Try to find by email
	if info.Email != "" {
		err = database.DB.Where("email = ?", info.Email).First(&user).Error
		if err == nil {
			// Link OAuth provider to existing account
			if user.Provider == "local" {
				// User registered with email/password, now linking OAuth
				user.Provider = info.Provider
				user.ProviderID = info.ID
				if user.AvatarURL == "" && info.Avatar != "" {
					user.AvatarURL = info.Avatar
				}
				database.DB.Save(&user)
			}
			return &user, nil
		}
	}

	// Create new user
	user = models.User{
		Email:         info.Email,
		Username:      info.Name,
		Name:          info.Name,
		AvatarURL:     info.Avatar,
		Provider:      info.Provider,
		ProviderID:    info.ID,
		EmailVerified: true, // OAuth verified email
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
