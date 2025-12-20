package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// ChallengeDifficulty represents challenge difficulty levels
type ChallengeDifficulty string

const (
	ChallengeEasy   ChallengeDifficulty = "Easy"
	ChallengeMedium ChallengeDifficulty = "Medium"
	ChallengeHard   ChallengeDifficulty = "Hard"
	ChallengeExpert ChallengeDifficulty = "Expert"
)

// ChallengeType represents types of challenges
type ChallengeType string

const (
	TypeDaily   ChallengeType = "daily"
	TypeWeekly  ChallengeType = "weekly"
	TypeSpecial ChallengeType = "special"
	TypeRegular ChallengeType = "regular"
)

// ProgressStatus represents challenge progress status
type ProgressStatus string

const (
	ProgressNotStarted ProgressStatus = "not_started"
	ProgressInProgress ProgressStatus = "in_progress"
	ProgressCompleted  ProgressStatus = "completed"
	ProgressFailed     ProgressStatus = "failed"
)

// Challenge represents challenges
type Challenge struct {
	ID                 string              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Title              string              `gorm:"size:200;not null" json:"title"`
	Description        string              `gorm:"type:text" json:"description"`
	Difficulty         ChallengeDifficulty `gorm:"type:varchar(20);not null" json:"difficulty"`
	XPReward           int                 `gorm:"not null" json:"xp_reward"`
	Category           string              `gorm:"size:50" json:"category"`
	Type               ChallengeType       `gorm:"type:varchar(20);default:'regular'" json:"type"`
	TimeLimitHours     *int                `json:"time_limit_hours"`
	MaxParticipants    *int                `json:"max_participants"`
	Prerequisites      pq.StringArray      `gorm:"type:uuid[]" json:"prerequisites"`
	Instructions       datatypes.JSON      `gorm:"type:jsonb" json:"instructions"`
	ValidationCriteria datatypes.JSON      `gorm:"type:jsonb" json:"validation_criteria"`
	IsActive           bool                `gorm:"default:true" json:"is_active"`
	StartsAt           *time.Time          `json:"starts_at"`
	EndsAt             *time.Time          `json:"ends_at"`
	CreatedAt          time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time           `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Participants []ChallengeProgress `gorm:"foreignKey:ChallengeID" json:"participants,omitempty"`
}

func (Challenge) TableName() string {
	return "challenges"
}

// ChallengeProgress represents user's progress on a challenge
type ChallengeProgress struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID           string         `gorm:"type:uuid;index;not null" json:"user_id"`
	ChallengeID      string         `gorm:"type:uuid;index;not null" json:"challenge_id"`
	Progress         int            `gorm:"default:0" json:"progress"`
	Status           ProgressStatus `gorm:"type:varchar(20);default:'not_started'" json:"status"`
	CurrentStep      int            `gorm:"default:0" json:"current_step"`
	SubmissionData   datatypes.JSON `gorm:"type:jsonb" json:"submission_data"`
	Feedback         string         `gorm:"type:text" json:"feedback"`
	XPEarned         *int           `json:"xp_earned"`
	TimeSpentMinutes int            `gorm:"default:0" json:"time_spent_minutes"`
	StartedAt        *time.Time     `json:"started_at"`
	CompletedAt      *time.Time     `json:"completed_at"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
}

func (ChallengeProgress) TableName() string {
	return "challenge_progress"
}

// DailyChallenge represents daily challenges
type DailyChallenge struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ChallengeID      string    `gorm:"type:uuid;index;not null" json:"challenge_id"`
	Date             string    `gorm:"type:date;uniqueIndex;not null" json:"date"`
	XPMultiplier     float64   `gorm:"type:decimal(3,2);default:1.0" json:"xp_multiplier"`
	ParticipantCount int       `gorm:"default:0" json:"participant_count"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
}

func (DailyChallenge) TableName() string {
	return "daily_challenges"
}

// IsAvailable checks if challenge is currently available
func (c *Challenge) IsAvailable() bool {
	now := time.Now()

	if !c.IsActive {
		return false
	}

	if c.StartsAt != nil && now.Before(*c.StartsAt) {
		return false
	}

	if c.EndsAt != nil && now.After(*c.EndsAt) {
		return false
	}

	return true
}

// TimeRemaining returns time remaining for timed challenges
func (c *Challenge) TimeRemaining() *time.Duration {
	if c.EndsAt == nil {
		return nil
	}

	remaining := time.Until(*c.EndsAt)
	if remaining < 0 {
		remaining = 0
	}

	return &remaining
}
