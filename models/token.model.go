package models

import (
	"time"
)

// PasswordResetToken stores password reset tokens
type PasswordResetToken struct {
	ID        string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID    string     `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string     `gorm:"uniqueIndex;size:255;not null" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsValid checks if the token is still valid
func (t *PasswordResetToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}

// EmailVerificationToken stores email verification tokens
type EmailVerificationToken struct {
	ID         string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID     string     `gorm:"type:uuid;not null;index" json:"user_id"`
	Token      string     `gorm:"uniqueIndex;size:255;not null" json:"token"`
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

// IsValid checks if the token is still valid
func (t *EmailVerificationToken) IsValid() bool {
	return t.VerifiedAt == nil && time.Now().Before(t.ExpiresAt)
}
