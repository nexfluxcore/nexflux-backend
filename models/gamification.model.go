package models

import (
	"time"
)

// AchievementRarity represents achievement rarity levels
type AchievementRarity string

const (
	RarityCommon    AchievementRarity = "common"
	RarityRare      AchievementRarity = "rare"
	RarityEpic      AchievementRarity = "epic"
	RarityLegendary AchievementRarity = "legendary"
)

// Achievement represents achievements
type Achievement struct {
	ID               string            `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name             string            `gorm:"size:100;not null" json:"name"`
	Description      string            `gorm:"type:text" json:"description"`
	Icon             string            `gorm:"size:50" json:"icon"`
	Rarity           AchievementRarity `gorm:"type:varchar(20);not null" json:"rarity"`
	RequirementType  string            `gorm:"size:50" json:"requirement_type"`
	RequirementValue int               `json:"requirement_value"`
	XPReward         int               `gorm:"default:0" json:"xp_reward"`
	IsActive         bool              `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time         `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Users []UserAchievement `gorm:"foreignKey:AchievementID" json:"users,omitempty"`
}

func (Achievement) TableName() string {
	return "achievements"
}

// UserAchievement represents unlocked achievements by users
type UserAchievement struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID        string    `gorm:"type:uuid;index;not null" json:"user_id"`
	AchievementID string    `gorm:"type:uuid;index;not null" json:"achievement_id"`
	UnlockedAt    time.Time `gorm:"autoCreateTime" json:"unlocked_at"`

	// Relations
	User        *User        `gorm:"foreignKey:UserID" json:"-"`
	Achievement *Achievement `gorm:"foreignKey:AchievementID" json:"achievement,omitempty"`
}

func (UserAchievement) TableName() string {
	return "user_achievements"
}

// LeaderboardType represents leaderboard types
type LeaderboardType string

const (
	LeaderboardWeekly  LeaderboardType = "weekly"
	LeaderboardMonthly LeaderboardType = "monthly"
	LeaderboardAllTime LeaderboardType = "all_time"
)

// Leaderboard represents leaderboards
type Leaderboard struct {
	ID          string          `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Type        LeaderboardType `gorm:"type:varchar(20);not null" json:"type"`
	PeriodStart *string         `gorm:"type:date" json:"period_start"`
	PeriodEnd   *string         `gorm:"type:date" json:"period_end"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Entries []LeaderboardEntry `gorm:"foreignKey:LeaderboardID" json:"entries,omitempty"`
}

func (Leaderboard) TableName() string {
	return "leaderboards"
}

// LeaderboardEntry represents entries in a leaderboard
type LeaderboardEntry struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	LeaderboardID string    `gorm:"type:uuid;index;not null" json:"leaderboard_id"`
	UserID        string    `gorm:"type:uuid;index;not null" json:"user_id"`
	Rank          int       `gorm:"not null" json:"rank"`
	XP            int       `gorm:"not null" json:"xp"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Leaderboard *Leaderboard `gorm:"foreignKey:LeaderboardID" json:"-"`
	User        *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (LeaderboardEntry) TableName() string {
	return "leaderboard_entries"
}

// GetRarityColor returns CSS color class for achievement rarity
func (a *Achievement) GetRarityColor() string {
	switch a.Rarity {
	case RarityCommon:
		return "text-gray-400"
	case RarityRare:
		return "text-blue-500"
	case RarityEpic:
		return "text-purple-500"
	case RarityLegendary:
		return "text-yellow-500"
	default:
		return "text-gray-400"
	}
}
