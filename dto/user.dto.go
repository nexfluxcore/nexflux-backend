package dto

import "time"

// ============================================
// User DTOs
// ============================================

// UserResponse represents user data in responses
type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Username      string    `json:"username"`
	Role          string    `json:"role"`
	AvatarURL     string    `json:"avatar_url"`
	Bio           string    `json:"bio"`
	Level         int       `json:"level"`
	TotalXP       int       `json:"total_xp"`
	CurrentXP     int       `json:"current_xp"`
	TargetXP      int       `json:"target_xp"`
	StreakDays    int       `json:"streak_days"`
	Language      string    `json:"language"`
	Theme         string    `json:"theme"`
	IsPro         bool      `json:"is_pro"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserUpdateRequest for updating user profile
type UserUpdateRequest struct {
	Name      string `json:"name" binding:"omitempty,max=100"`
	Username  string `json:"username" binding:"omitempty,max=50"`
	Bio       string `json:"bio" binding:"omitempty,max=500"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,url"` // URL from upload endpoint
	Language  string `json:"language" binding:"omitempty,oneof=en id jp"`
	Theme     string `json:"theme" binding:"omitempty,oneof=light dark system"`
}

// UserStatsResponse for gamification stats
type UserStatsResponse struct {
	Level         int              `json:"level"`
	TotalXP       int              `json:"total_xp"`
	CurrentXP     int              `json:"current_xp"`
	TargetXP      int              `json:"target_xp"`
	XPToNextLevel int              `json:"xp_to_next_level"`
	Streak        StreakInfo       `json:"streak"`
	Challenges    ChallengeStats   `json:"challenges"`
	Projects      ProjectStats     `json:"projects"`
	Achievements  AchievementStats `json:"achievements"`
	XPThisWeek    int              `json:"xp_this_week"`
}

type StreakInfo struct {
	Current    int    `json:"current"`
	Longest    int    `json:"longest"`
	LastActive string `json:"last_active"`
}

type ChallengeStats struct {
	Completed      int `json:"completed"`
	InProgress     int `json:"in_progress"`
	TotalAvailable int `json:"total_available"`
}

type ProjectStats struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
}

type AchievementStats struct {
	Unlocked int `json:"unlocked"`
	Total    int `json:"total"`
}

// ============================================
// Settings DTOs
// ============================================

// UserSettingsResponse for user settings
type UserSettingsResponse struct {
	NotificationEmail     bool `json:"notification_email"`
	NotificationPush      bool `json:"notification_push"`
	NotificationMarketing bool `json:"notification_marketing"`
	NotificationUpdates   bool `json:"notification_updates"`
}

// UserSettingsUpdateRequest for updating settings
type UserSettingsUpdateRequest struct {
	NotificationEmail     *bool `json:"notification_email"`
	NotificationPush      *bool `json:"notification_push"`
	NotificationMarketing *bool `json:"notification_marketing"`
	NotificationUpdates   *bool `json:"notification_updates"`
}

// NOTE: Password-related DTOs (ChangePasswordRequest, ForgotPasswordRequest, ResetPasswordRequest)
// are defined in auth.dto.go
