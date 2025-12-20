package models

import (
	"time"

	"github.com/lib/pq"
)

// UserRole represents user role types
type UserRole string

const (
	RoleStudent     UserRole = "student"
	RoleEducator    UserRole = "educator"
	RoleMaker       UserRole = "maker"
	RoleInstitution UserRole = "institution"
	RoleAdmin       UserRole = "admin"
)

// Theme represents theme preference
type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

// User represents the users table
type User struct {
	ID            string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Email         string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash  string     `gorm:"size:255" json:"-"`
	Name          string     `gorm:"size:100;not null" json:"name"`
	Username      string     `gorm:"uniqueIndex;size:50" json:"username"`
	Role          UserRole   `gorm:"type:varchar(20);default:'student'" json:"role"`
	AvatarURL     string     `gorm:"size:500" json:"avatar_url"`
	Bio           string     `gorm:"type:text" json:"bio"`
	Level         int        `gorm:"default:1" json:"level"`
	TotalXP       int        `gorm:"default:0" json:"total_xp"`
	CurrentXP     int        `gorm:"default:0" json:"current_xp"`
	TargetXP      int        `gorm:"default:1000" json:"target_xp"`
	StreakDays    int        `gorm:"default:0" json:"streak_days"`
	LastActiveAt  *time.Time `json:"last_active_at"`
	Language      string     `gorm:"size:5;default:'en'" json:"language"`
	Theme         Theme      `gorm:"type:varchar(10);default:'system'" json:"theme"`
	IsPro         bool       `gorm:"default:false" json:"is_pro"`
	EmailVerified bool       `gorm:"default:false" json:"email_verified"`
	Provider      string     `gorm:"size:50" json:"provider"`
	ProviderID    string     `gorm:"size:255" json:"provider_id"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Settings     *UserSettings     `gorm:"foreignKey:UserID" json:"settings,omitempty"`
	Sessions     []UserSession     `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
	Projects     []Project         `gorm:"foreignKey:UserID" json:"projects,omitempty"`
	Streak       *UserStreak       `gorm:"foreignKey:UserID" json:"streak,omitempty"`
	Achievements []UserAchievement `gorm:"foreignKey:UserID" json:"achievements,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// UserSettings represents user notification settings
type UserSettings struct {
	ID                    string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID                string    `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	NotificationEmail     bool      `gorm:"default:true" json:"notification_email"`
	NotificationPush      bool      `gorm:"default:true" json:"notification_push"`
	NotificationMarketing bool      `gorm:"default:false" json:"notification_marketing"`
	NotificationUpdates   bool      `gorm:"default:true" json:"notification_updates"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// UserSession represents active user sessions
type UserSession struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID     string    `gorm:"type:uuid;index;not null" json:"user_id"`
	Token      string    `gorm:"size:500;not null" json:"-"`
	DeviceInfo string    `gorm:"size:255" json:"device_info"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

// UserStreak represents user streak information
type UserStreak struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID           string    `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	CurrentStreak    int       `gorm:"default:0" json:"current_streak"`
	LongestStreak    int       `gorm:"default:0" json:"longest_streak"`
	LastActivityDate *string   `gorm:"type:date" json:"last_activity_date"`
	StreakProtects   int       `gorm:"default:0" json:"streak_protects"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (UserStreak) TableName() string {
	return "user_streaks"
}

// UserFavoriteComponent represents user's favorite components
type UserFavoriteComponent struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID      string    `gorm:"type:uuid;index;not null" json:"user_id"`
	ComponentID string    `gorm:"type:uuid;index;not null" json:"component_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	User      *User      `gorm:"foreignKey:UserID" json:"-"`
	Component *Component `gorm:"foreignKey:ComponentID" json:"component,omitempty"`
}

func (UserFavoriteComponent) TableName() string {
	return "user_favorite_components"
}

// Helper to check if password is set (for OAuth users)
func (u *User) HasPassword() bool {
	return u.PasswordHash != ""
}

// XPToNextLevel calculates XP needed to reach next level
func (u *User) XPToNextLevel() int {
	return u.TargetXP - u.CurrentXP
}

// ProgressToNextLevel calculates progress percentage to next level
func (u *User) ProgressToNextLevel() float64 {
	if u.TargetXP == 0 {
		return 0
	}
	return float64(u.CurrentXP) / float64(u.TargetXP) * 100
}

// Unused but needed for pq package
var _ = pq.Array
