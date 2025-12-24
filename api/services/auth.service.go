package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"nexfi-backend/database"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"nexfi-backend/pkg/redis"
	"nexfi-backend/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic
type AuthService struct {
	db *gorm.DB
}

// NewAuthService creates a new AuthService
func NewAuthService() *AuthService {
	return &AuthService{
		db: database.DB,
	}
}

// ==================================================
// REGISTRATION & LOGIN
// ==================================================

// Register creates a new user account
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if email exists
	var existing models.User
	if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("email already registered")
	}

	// Check if username exists
	// if err := s.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
	// 	return nil, errors.New("username already taken")
	// }

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := models.User{
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		Name:          req.Name,
		Username:      req.Username,
		Provider:      "local",
		EmailVerified: false,
		Role:          models.RoleStudent,
		Level:         1,
		TotalXP:       0,
		CurrentXP:     0,
		TargetXP:      1000,
	}

	if err := s.db.Create(&user).Error; err != nil {
		// Handle duplicate key errors with user-friendly messages
		errStr := err.Error()
		if strings.Contains(errStr, "idx_users_email") || strings.Contains(errStr, "email") {
			return nil, errors.New("email already registered")
		}
		if strings.Contains(errStr, "idx_users_username") || strings.Contains(errStr, "username") {
			return nil, errors.New("username already taken")
		}
		return nil, errors.New("failed to create account, please try again")
	}

	// Generate tokens
	token, refreshToken, err := s.generateTokens(&user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    s.getExpiryHours() * 3600,
		User:         s.toUserAuthInfo(&user),
	}, nil
}

// Login authenticates user and returns tokens
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user has password (not OAuth-only)
	if user.PasswordHash == "" && user.Provider != "local" {
		return nil, errors.New("please login with " + user.Provider)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Update last active
	now := time.Now()
	s.db.Model(&user).Updates(map[string]interface{}{
		"last_active_at": now,
	})

	// Generate tokens
	token, refreshToken, err := s.generateTokens(&user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    s.getExpiryHours() * 3600,
		User:         s.toUserAuthInfo(&user),
	}, nil
}

// ==================================================
// PASSWORD RESET
// ==================================================

// ForgotPassword initiates password reset flow
func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) error {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists - security best practice
		return nil
	}

	// Invalidate any existing tokens
	s.db.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", user.ID).
		Updates(map[string]interface{}{
			"used_at": time.Now(),
		})

	// Generate new token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)

	// Create reset token with 1 hour expiry
	resetToken := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.db.Create(&resetToken).Error; err != nil {
		return err
	}

	// TODO: Send email with reset link
	// frontendURL := os.Getenv("FRONTEND_URL")
	// resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)
	// emailService.SendPasswordResetEmail(user.Email, user.Name, resetLink)

	return nil
}

// VerifyResetToken validates a password reset token
func (s *AuthService) VerifyResetToken(token string) error {
	var resetToken models.PasswordResetToken
	err := s.db.Where("token = ?", token).First(&resetToken).Error
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if !resetToken.IsValid() {
		return errors.New("invalid or expired token")
	}

	return nil
}

// ResetPassword resets user password with valid token
func (s *AuthService) ResetPassword(req dto.ResetPasswordRequest) error {
	var resetToken models.PasswordResetToken
	err := s.db.Where("token = ?", req.Token).Preload("User").First(&resetToken).Error
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if !resetToken.IsValid() {
		return errors.New("invalid or expired token")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	if err := s.db.Model(&models.User{}).
		Where("id = ?", resetToken.UserID).
		Update("password_hash", string(hashedPassword)).Error; err != nil {
		return err
	}

	// Mark token as used
	now := time.Now()
	s.db.Model(&resetToken).Update("used_at", now)

	// Invalidate all sessions (optional - for extra security)
	redis.InvalidateUserTokens(resetToken.UserID)

	return nil
}

// ChangePassword changes password for authenticated user
func (s *AuthService) ChangePassword(userID string, req dto.ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	return s.db.Model(&user).Update("password_hash", string(hashedPassword)).Error
}

// ==================================================
// TOKEN REFRESH & LOGOUT
// ==================================================

// RefreshToken refreshes access token
func (s *AuthService) RefreshToken(refreshToken string) (*dto.RefreshTokenResponse, error) {
	claims, err := utils.VerifyJWT(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Verify it's a refresh token
	if claims.Issuer != "nexflux-refresh" {
		return nil, errors.New("invalid refresh token")
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new tokens
	newToken, newRefreshToken, err := s.generateTokens(&user)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.getExpiryHours() * 3600,
	}, nil
}

// Logout invalidates user's tokens
func (s *AuthService) Logout(userID string) error {
	return redis.InvalidateUserTokens(userID)
}

// GetCurrentUser returns current user from token
func (s *AuthService) GetCurrentUser(userID string) (*dto.UserAuthInfo, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	result := s.toUserAuthInfo(&user)
	return &result, nil
}

// ==================================================
// HELPER FUNCTIONS
// ==================================================

func (s *AuthService) generateTokens(user *models.User) (string, string, error) {
	// Generate access token with user info
	token, err := utils.GenerateJWTWithInfo(utils.JWTUserInfo{
		ID:       user.ID,
		Email:    user.Email,
		Role:     string(user.Role),
		Name:     user.Name,
		Username: user.Username,
	})
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	// Store token in Redis
	if err := redis.StoreToken(user.ID, token); err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func (s *AuthService) toUserAuthInfo(user *models.User) dto.UserAuthInfo {
	return dto.UserAuthInfo{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		Username:   user.Username,
		Avatar:     user.AvatarURL,
		Role:       string(user.Role),
		Level:      user.Level,
		TotalXP:    user.TotalXP,
		StreakDays: user.StreakDays,
		IsPro:      user.IsPro,
		IsVerified: user.EmailVerified,
		Provider:   user.Provider,
	}
}

func (s *AuthService) getExpiryHours() int {
	expireHours := 24
	if h := os.Getenv("JWT_EXPIRE_HOURS"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil {
			expireHours = parsed
		}
	}
	return expireHours
}
