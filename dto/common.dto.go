package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Notification DTOs
// ============================================

// NotificationListRequest for listing notifications
type NotificationListRequest struct {
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
	Type       string `form:"type"`
	UnreadOnly bool   `form:"unread_only"`
}

// NotificationResponse for single notification
type NotificationResponse struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Data      datatypes.JSON `json:"data"`
	IsRead    bool           `json:"is_read"`
	CreatedAt time.Time      `json:"created_at"`
}

// NotificationListResponse with unread count
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int                    `json:"unread_count"`
	Pagination    PaginationResponse     `json:"pagination"`
}

// ============================================
// Achievement DTOs
// ============================================

// AchievementResponse for single achievement
type AchievementResponse struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Icon             string     `json:"icon"`
	Rarity           string     `json:"rarity"`
	RequirementType  string     `json:"requirement_type"`
	RequirementValue int        `json:"requirement_value"`
	XPReward         int        `json:"xp_reward"`
	IsUnlocked       bool       `json:"is_unlocked,omitempty"`
	UnlockedAt       *time.Time `json:"unlocked_at,omitempty"`
	Progress         int        `json:"progress,omitempty"`
}

// AchievementListResponse for achievements list
type AchievementListResponse struct {
	Achievements []AchievementResponse `json:"achievements"`
	Stats        struct {
		Unlocked int `json:"unlocked"`
		Total    int `json:"total"`
	} `json:"stats"`
}

// ============================================
// Leaderboard DTOs
// ============================================

// LeaderboardRequest for getting leaderboard
type LeaderboardRequest struct {
	Type  string `form:"type,default=weekly"` // weekly, monthly, all_time
	Limit int    `form:"limit,default=10"`
}

// LeaderboardEntryResponse for single leaderboard entry
type LeaderboardEntryResponse struct {
	Rank          int               `json:"rank"`
	User          UserBriefResponse `json:"user"`
	XP            int               `json:"xp"`
	IsCurrentUser bool              `json:"is_current_user"`
}

// LeaderboardResponse for full leaderboard
type LeaderboardResponse struct {
	Type            string                     `json:"type"`
	Period          *PeriodResponse            `json:"period,omitempty"`
	Entries         []LeaderboardEntryResponse `json:"entries"`
	CurrentUserRank int                        `json:"current_user_rank"`
}

// PeriodResponse for leaderboard period
type PeriodResponse struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// UserBriefResponse for minimal user info
type UserBriefResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Level     int    `json:"level"`
}

// ============================================
// Streak DTOs
// ============================================

// StreakResponse for user streak info
type StreakResponse struct {
	CurrentStreak    int    `json:"current_streak"`
	LongestStreak    int    `json:"longest_streak"`
	LastActivityDate string `json:"last_activity_date"`
	StreakProtects   int    `json:"streak_protects"`
	IsActiveToday    bool   `json:"is_active_today"`
}

// ============================================
// Common DTOs
// ============================================

// PaginationResponse for paginated responses
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// APIResponse is the standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError for error responses
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// IDResponse for responses that return just an ID
type IDResponse struct {
	ID string `json:"id"`
}

// MessageResponse for simple message responses
type MessageResponse struct {
	Message string `json:"message"`
}
