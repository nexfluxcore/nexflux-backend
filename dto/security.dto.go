package dto

import "time"

// ============================================
// Password Management DTOs
// ============================================

// SecurityChangePasswordRequest for changing password (security settings)
type SecurityChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// SecurityChangePasswordResponse for password change response
type SecurityChangePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================
// Two-Factor Authentication DTOs
// ============================================

// Enable2FAResponse for enabling 2FA
type Enable2FAResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// Verify2FARequest for verifying 2FA code
type Verify2FARequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// Disable2FARequest for disabling 2FA
type Disable2FARequest struct {
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TwoFAStatusResponse for 2FA status
type TwoFAStatusResponse struct {
	Enabled              bool       `json:"enabled"`
	EnabledAt            *time.Time `json:"enabled_at"`
	BackupCodesRemaining int        `json:"backup_codes_remaining"`
}

// ============================================
// Session Management DTOs
// ============================================

// SecuritySessionResponse for session list (security settings)
type SecuritySessionResponse struct {
	ID         string    `json:"id"`
	Device     string    `json:"device"`
	Browser    string    `json:"browser"`
	IPAddress  string    `json:"ip_address"`
	Location   string    `json:"location"`
	LastActive time.Time `json:"last_active"`
	IsCurrent  bool      `json:"is_current"`
	CreatedAt  time.Time `json:"created_at"`
}

// RevokeAllSessionsResponse for revoking all sessions
type RevokeAllSessionsResponse struct {
	RevokedCount int `json:"revoked_count"`
}

// ============================================
// Login History DTOs
// ============================================

// LoginHistoryRequest for listing login history
type LoginHistoryRequest struct {
	Page  int `form:"page,default=1"`
	Limit int `form:"limit,default=20"`
}

// LoginHistoryResponse for login history item
type LoginHistoryResponse struct {
	ID            string    `json:"id"`
	Device        string    `json:"device"`
	Browser       string    `json:"browser"`
	IPAddress     string    `json:"ip_address"`
	Location      string    `json:"location"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	FailureReason *string   `json:"failure_reason"`
}

// ============================================
// Security Settings DTOs
// ============================================

// SecuritySettingsResponse for overall security settings
type SecuritySettingsResponse struct {
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	PasswordChangedAt   *time.Time `json:"password_changed_at"`
	ActiveSessionsCount int        `json:"active_sessions_count"`
	RecentLoginAttempts int        `json:"recent_login_attempts"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	LastLoginIP         string     `json:"last_login_ip"`
	LastLoginLocation   string     `json:"last_login_location"`
	AccountLocked       bool       `json:"account_locked"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
}

// Login2FARequest for 2FA during login
type Login2FARequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"` // Optional - only required if 2FA is enabled
}
