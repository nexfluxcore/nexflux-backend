package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// ============================================
// Security Models
// ============================================

// NOTE: UserSession is defined in user.model.go

// LoginHistory represents a login attempt (success or failure)
type LoginHistory struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID        *string   `gorm:"type:uuid;index" json:"user_id"`
	Email         string    `gorm:"size:255" json:"email"` // Store email even for failed attempts
	Device        string    `gorm:"size:100" json:"device"`
	Browser       string    `gorm:"size:100" json:"browser"`
	IPAddress     string    `gorm:"size:45" json:"ip_address"`
	Location      string    `gorm:"size:200" json:"location"`
	UserAgent     string    `gorm:"type:text" json:"-"`
	Status        string    `gorm:"size:20;not null" json:"status"` // success, failed
	FailureReason *string   `gorm:"size:50" json:"failure_reason"`  // invalid_password, invalid_email, 2fa_failed, etc.
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (LoginHistory) TableName() string {
	return "login_history"
}

// LoginStatus constants
const (
	LoginStatusSuccess = "success"
	LoginStatusFailed  = "failed"
)

// LoginFailureReason constants
const (
	LoginFailureInvalidPassword = "invalid_password"
	LoginFailureInvalidEmail    = "invalid_email"
	LoginFailure2FAFailed       = "2fa_failed"
	LoginFailureAccountLocked   = "account_locked"
	LoginFailureIPBlocked       = "ip_blocked"
)

// User2FA represents 2FA configuration for a user
type User2FA struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID      string         `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	SecretKey   string         `gorm:"size:100;not null" json:"-"` // Encrypted TOTP secret
	IsEnabled   bool           `gorm:"default:false" json:"is_enabled"`
	BackupCodes datatypes.JSON `gorm:"type:jsonb" json:"-"` // Encrypted array of backup codes
	EnabledAt   *time.Time     `json:"enabled_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (User2FA) TableName() string {
	return "user_2fa"
}

// PasswordHistory tracks password changes to prevent reuse
type PasswordHistory struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID       string    `gorm:"type:uuid;index;not null" json:"user_id"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (PasswordHistory) TableName() string {
	return "password_history"
}

// SecurityLog for tracking security-related events
type SecurityLog struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID    string         `gorm:"type:uuid;index;not null" json:"user_id"`
	EventType string         `gorm:"size:50;not null" json:"event_type"` // password_change, 2fa_enable, session_revoke, etc.
	IPAddress string         `gorm:"size:45" json:"ip_address"`
	UserAgent string         `gorm:"type:text" json:"-"`
	Details   datatypes.JSON `gorm:"type:jsonb" json:"details"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (SecurityLog) TableName() string {
	return "security_logs"
}

// SecurityEventType constants
const (
	SecurityEventPasswordChange   = "password_change"
	SecurityEvent2FAEnable        = "2fa_enable"
	SecurityEvent2FADisable       = "2fa_disable"
	SecurityEventSessionRevoke    = "session_revoke"
	SecurityEventSessionRevokeAll = "session_revoke_all"
	SecurityEventAccountLocked    = "account_locked"
	SecurityEventAccountUnlocked  = "account_unlocked"
	SecurityEventSuspiciousLogin  = "suspicious_login"
)

// User extensions for security
// These fields should be added to the User model:
// PasswordChangedAt    *time.Time `json:"password_changed_at"`
// FailedLoginAttempts  int        `gorm:"default:0" json:"-"`
// LockedUntil         *time.Time `json:"locked_until"`
// TwoFactorEnabled     bool       `gorm:"default:false" json:"two_factor_enabled"`

// Placeholder for pq usage
var _ = pq.Array
