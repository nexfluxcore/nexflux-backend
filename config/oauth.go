package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Apple OAuth2 endpoint
var AppleEndpoint = oauth2.Endpoint{
	AuthURL:  "https://appleid.apple.com/auth/authorize",
	TokenURL: "https://appleid.apple.com/auth/token",
}

// OAuthConfig holds configuration for all OAuth providers
type OAuthConfig struct {
	Google *oauth2.Config
	GitHub *oauth2.Config
	Apple  *oauth2.Config
}

// AppleConfig holds Apple-specific configuration
type AppleConfig struct {
	TeamID     string
	ClientID   string
	KeyID      string
	PrivateKey string
}

var OAuth *OAuthConfig
var AppleOAuthConfig *AppleConfig

// InitOAuthConfig initializes all OAuth provider configurations
func InitOAuthConfig() {
	baseURL := getOAuthEnv("OAUTH_BASE_URL", "http://localhost:8080")

	OAuth = &OAuthConfig{
		Google: initGoogleOAuth(baseURL),
		GitHub: initGitHubOAuth(baseURL),
		Apple:  initAppleOAuth(baseURL),
	}

	// Initialize Apple-specific config
	AppleOAuthConfig = &AppleConfig{
		TeamID:     os.Getenv("APPLE_TEAM_ID"),
		ClientID:   os.Getenv("APPLE_CLIENT_ID"),
		KeyID:      os.Getenv("APPLE_KEY_ID"),
		PrivateKey: os.Getenv("APPLE_PRIVATE_KEY"),
	}
}

func initGoogleOAuth(baseURL string) *oauth2.Config {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/api/v1/auth/google/callback",
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}

func initGitHubOAuth(baseURL string) *oauth2.Config {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/api/v1/auth/github/callback",
		Scopes: []string{
			"user:email",
			"read:user",
		},
		Endpoint: github.Endpoint,
	}
}

func initAppleOAuth(baseURL string) *oauth2.Config {
	clientID := os.Getenv("APPLE_CLIENT_ID")
	teamID := os.Getenv("APPLE_TEAM_ID")
	keyID := os.Getenv("APPLE_KEY_ID")
	privateKey := os.Getenv("APPLE_PRIVATE_KEY")

	if clientID == "" || teamID == "" || keyID == "" || privateKey == "" {
		return nil
	}

	// Generate client secret for Apple
	clientSecret, err := GenerateAppleClientSecret(teamID, clientID, keyID, privateKey)
	if err != nil {
		return nil
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/api/v1/auth/apple/callback",
		Scopes: []string{
			"name",
			"email",
		},
		Endpoint: AppleEndpoint,
	}
}

// GenerateAppleClientSecret generates a JWT client secret for Apple Sign In
// Apple requires a JWT signed with your private key as the client_secret
func GenerateAppleClientSecret(teamID, clientID, keyID, privateKeyPEM string) (string, error) {
	// Parse the private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", nil
	}

	// Create the JWT claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24 * 180).Unix(), // Max 6 months
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	}

	// Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	return token.SignedString(ecdsaKey)
}

func getOAuthEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
