package dto

// ==================================================
// REGISTRATION & LOGIN
// ==================================================

// RegisterRequest - request body untuk registrasi user baru
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
	Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
}

// LoginRequest - request body untuk login user
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse - response berisi token dan info user
type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresIn    int          `json:"expires_in,omitempty"` // seconds
	User         UserAuthInfo `json:"user"`
}

// UserAuthInfo - informasi user yang dikembalikan saat auth
type UserAuthInfo struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar,omitempty"`
	Role       string `json:"role"`
	Level      int    `json:"level"`
	TotalXP    int    `json:"total_xp"`
	StreakDays int    `json:"streak_days"`
	IsPro      bool   `json:"is_pro"`
	IsVerified bool   `json:"is_verified"`
	Provider   string `json:"provider,omitempty"`
}

// ==================================================
// PASSWORD RESET
// ==================================================

// ForgotPasswordRequest - request for forgot password
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyResetTokenRequest - request to verify reset token
type VerifyResetTokenRequest struct {
	Token string `form:"token" binding:"required"`
}

// ResetPasswordRequest - request to reset password with token
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ChangePasswordRequest - request to change password (authenticated)
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ==================================================
// EMAIL VERIFICATION
// ==================================================

// ResendVerificationRequest - request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ==================================================
// TOKEN REFRESH
// ==================================================

// RefreshTokenRequest - request to refresh access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse - response with new tokens
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// ==================================================
// OAUTH
// ==================================================

// OAuthCallbackRequest - request dari OAuth callback
type OAuthCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// OAuthUserInfo - informasi user dari OAuth provider
type OAuthUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
}

// OAuthURLResponse - response berisi URL untuk OAuth
type OAuthURLResponse struct {
	URL   string `json:"url"`
	State string `json:"state"`
}
