package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTClaims contains JWT token claims with user info
// SECURITY NOTE: Only include non-sensitive data in JWT claims
// - Avoid: passwords, tokens, internal IDs that shouldn't be exposed
// - Safe to include: user ID, email, role, name (for display purposes)
type JWTClaims struct {
	UserID   string `json:"sub"`      // Subject = User ID
	Email    string `json:"email"`    // Safe: used for identification
	Role     string `json:"role"`     // Safe: used for authorization
	Name     string `json:"name"`     // Safe: for display purposes only
	Username string `json:"username"` // Safe: public identifier
	jwt.StandardClaims
}

// JWTUserInfo contains user info for generating JWT
type JWTUserInfo struct {
	ID       string
	Email    string
	Role     string
	Name     string
	Username string
}

// GenerateJWT creates a JWT token with user claims
func GenerateJWT(userID string) (string, error) {
	return GenerateJWTWithInfo(JWTUserInfo{ID: userID})
}

// GenerateJWTWithInfo creates a JWT token with full user info in claims
func GenerateJWTWithInfo(user JWTUserInfo) (string, error) {
	// Get expiration from env, default 24 hours
	expireHours := 24
	if h := os.Getenv("JWT_EXPIRE_HOURS"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil {
			expireHours = parsed
		}
	}

	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)

	claims := &JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		Name:     user.Name,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "nexflux",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// GenerateRefreshToken creates a refresh token with longer expiry
func GenerateRefreshToken(userID string) (string, error) {
	// Refresh token expires in 7 days
	expireDays := 7
	if d := os.Getenv("JWT_REFRESH_EXPIRE_DAYS"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			expireDays = parsed
		}
	}

	expirationTime := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)

	claims := &JWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "nexflux-refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// VerifyJWT validates a JWT token and returns claims
func VerifyJWT(tokenString string) (*JWTClaims, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("secret key is not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token has expired
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// Legacy alias for backward compatibility
type Claims = JWTClaims
