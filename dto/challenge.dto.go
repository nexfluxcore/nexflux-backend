package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Challenge DTOs
// ============================================

// ChallengeListRequest for listing challenges
type ChallengeListRequest struct {
	Type       string `form:"type"`       // all, daily, weekly, special, regular
	Difficulty string `form:"difficulty"` // Easy, Medium, Hard, Expert
	Status     string `form:"status"`     // available, in_progress, completed, locked
	Category   string `form:"category"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
}

// ChallengeResponse for single challenge
type ChallengeResponse struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Difficulty       string     `json:"difficulty"`
	XPReward         int        `json:"xp_reward"`
	Category         string     `json:"category"`
	Type             string     `json:"type"`
	TimeLimitHours   *int       `json:"time_limit_hours"`
	MaxParticipants  *int       `json:"max_participants"`
	ParticipantCount int        `json:"participant_count,omitempty"`
	IsActive         bool       `json:"is_active"`
	StartsAt         *time.Time `json:"starts_at"`
	EndsAt           *time.Time `json:"ends_at"`
	TimeLeftHours    *float64   `json:"time_left_hours,omitempty"`
	UserStatus       string     `json:"user_status,omitempty"` // not_started, in_progress, completed, failed
	UserProgress     int        `json:"user_progress,omitempty"`
}

// ChallengeDetailResponse includes full challenge data
type ChallengeDetailResponse struct {
	ChallengeResponse
	Prerequisites      []string       `json:"prerequisites"`
	Instructions       datatypes.JSON `json:"instructions"`
	ValidationCriteria datatypes.JSON `json:"validation_criteria"`
}

// ChallengeListResponse with stats
type ChallengeListResponse struct {
	Challenges []ChallengeResponse `json:"challenges"`
	Stats      ChallengeStats      `json:"stats"`
}

// ============================================
// Daily Challenge DTOs
// ============================================

// DailyChallengeResponse for daily challenge
type DailyChallengeResponse struct {
	ID               string            `json:"id"`
	Challenge        ChallengeResponse `json:"challenge"`
	Date             string            `json:"date"`
	XPMultiplier     float64           `json:"xp_multiplier"`
	TimeLeftHours    float64           `json:"time_left_hours"`
	ParticipantCount int               `json:"participant_count"`
	UserStatus       string            `json:"user_status"`
}

// ============================================
// Challenge Progress DTOs
// ============================================

// ChallengeProgressResponse for user's progress
type ChallengeProgressResponse struct {
	ID               string             `json:"id"`
	ChallengeID      string             `json:"challenge_id"`
	Progress         int                `json:"progress"`
	Status           string             `json:"status"`
	CurrentStep      int                `json:"current_step"`
	XPEarned         *int               `json:"xp_earned"`
	TimeSpentMinutes int                `json:"time_spent_minutes"`
	StartedAt        *time.Time         `json:"started_at"`
	CompletedAt      *time.Time         `json:"completed_at"`
	Challenge        *ChallengeResponse `json:"challenge,omitempty"`
}

// UpdateProgressRequest for updating challenge progress
type UpdateProgressRequest struct {
	Progress    int `json:"progress" binding:"min=0,max=100"`
	CurrentStep int `json:"current_step" binding:"min=0"`
	TimeSpent   int `json:"time_spent_minutes" binding:"min=0"`
}

// SubmitChallengeRequest for submitting challenge completion
type SubmitChallengeRequest struct {
	SubmissionData datatypes.JSON `json:"submission_data" binding:"required"`
	Notes          string         `json:"notes"`
}

// SubmitChallengeResponse for challenge submission result
type SubmitChallengeResponse struct {
	XPEarned             int                   `json:"xp_earned"`
	BonusXP              int                   `json:"bonus_xp"`
	TotalXP              int                   `json:"total_xp"`
	NewLevel             int                   `json:"new_level"`
	LevelUp              bool                  `json:"level_up"`
	AchievementsUnlocked []AchievementResponse `json:"achievements_unlocked"`
}
