package repositories

import (
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// GamificationRepository handles gamification database operations
type GamificationRepository struct {
	*BaseRepository
}

// NewGamificationRepository creates a new GamificationRepository
func NewGamificationRepository(db *gorm.DB) *GamificationRepository {
	return &GamificationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// ============================================
// Achievement Methods
// ============================================

// GetAllAchievements gets all achievements
func (r *GamificationRepository) GetAllAchievements() ([]models.Achievement, error) {
	var achievements []models.Achievement
	err := r.DB.Where("is_active = ?", true).Order("rarity, name").Find(&achievements).Error
	return achievements, err
}

// GetUserAchievements gets user's unlocked achievements
func (r *GamificationRepository) GetUserAchievements(userID string) ([]models.UserAchievement, error) {
	var achievements []models.UserAchievement
	err := r.DB.Where("user_id = ?", userID).
		Preload("Achievement").
		Find(&achievements).Error
	return achievements, err
}

// UnlockAchievement unlocks an achievement for user
func (r *GamificationRepository) UnlockAchievement(userID, achievementID string) (*models.UserAchievement, error) {
	// Check if already unlocked
	var existing models.UserAchievement
	err := r.DB.Where("user_id = ? AND achievement_id = ?", userID, achievementID).First(&existing).Error
	if err == nil {
		return &existing, nil // Already unlocked
	}

	ua := &models.UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
	}

	err = r.DB.Create(ua).Error
	if err != nil {
		return nil, err
	}

	// Load achievement data
	r.DB.Preload("Achievement").First(ua, "id = ?", ua.ID)
	return ua, nil
}

// HasAchievement checks if user has an achievement
func (r *GamificationRepository) HasAchievement(userID, achievementID string) bool {
	var count int64
	r.DB.Model(&models.UserAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count)
	return count > 0
}

// GetAchievementProgress gets user's progress towards achievements
func (r *GamificationRepository) GetAchievementProgress(userID string) (map[string]int, error) {
	progress := make(map[string]int)

	// Challenges completed
	var challengesCompleted int64
	r.DB.Model(&models.ChallengeProgress{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&challengesCompleted)
	progress["challenges_completed"] = int(challengesCompleted)

	// Projects created
	var projectsCreated int64
	r.DB.Model(&models.Project{}).Where("user_id = ?", userID).Count(&projectsCreated)
	progress["projects_created"] = int(projectsCreated)

	// Streak days
	var streak models.UserStreak
	if err := r.DB.Where("user_id = ?", userID).First(&streak).Error; err == nil {
		progress["streak_days"] = streak.CurrentStreak
		progress["longest_streak"] = streak.LongestStreak
	}

	return progress, nil
}

// ============================================
// Leaderboard Methods
// ============================================

// GetOrCreateLeaderboard gets or creates a leaderboard
func (r *GamificationRepository) GetOrCreateLeaderboard(lbType string) (*models.Leaderboard, error) {
	var lb models.Leaderboard

	now := time.Now()
	var periodStart, periodEnd string

	switch lbType {
	case "weekly":
		// Get start of week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -(weekday - 1))
		end := start.AddDate(0, 0, 6)
		periodStart = start.Format("2006-01-02")
		periodEnd = end.Format("2006-01-02")
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, -1)
		periodStart = start.Format("2006-01-02")
		periodEnd = end.Format("2006-01-02")
	case "all_time":
		// No period for all-time
	}

	query := r.DB.Where("type = ?", lbType)
	if periodStart != "" {
		query = query.Where("period_start = ?", periodStart)
	}

	err := query.First(&lb).Error
	if err == gorm.ErrRecordNotFound {
		lb = models.Leaderboard{
			Type:        models.LeaderboardType(lbType),
			PeriodStart: &periodStart,
			PeriodEnd:   &periodEnd,
		}
		if lbType == "all_time" {
			lb.PeriodStart = nil
			lb.PeriodEnd = nil
		}
		r.DB.Create(&lb)
	}

	return &lb, nil
}

// GetLeaderboardEntries gets leaderboard entries
func (r *GamificationRepository) GetLeaderboardEntries(leaderboardID string, limit int) ([]models.LeaderboardEntry, error) {
	var entries []models.LeaderboardEntry
	err := r.DB.Where("leaderboard_id = ?", leaderboardID).
		Order("rank ASC").
		Limit(limit).
		Preload("User").
		Find(&entries).Error
	return entries, err
}

// GetUserRank gets user's rank in a leaderboard
func (r *GamificationRepository) GetUserRank(leaderboardID, userID string) (int, error) {
	var entry models.LeaderboardEntry
	err := r.DB.Where("leaderboard_id = ? AND user_id = ?", leaderboardID, userID).First(&entry).Error
	if err != nil {
		return 0, err
	}
	return entry.Rank, nil
}

// UpdateLeaderboardRankings recalculates leaderboard rankings
func (r *GamificationRepository) UpdateLeaderboardRankings(leaderboardID string, lbType string) error {
	// This would typically be run by a cron job
	// For now, we calculate based on total_xp or period XP

	var users []models.User
	if lbType == "all_time" {
		r.DB.Order("total_xp DESC").Limit(100).Find(&users)
	} else {
		// For weekly/monthly, we'd need to track XP gained in the period
		// Simplified: use total_xp for now
		r.DB.Order("total_xp DESC").Limit(100).Find(&users)
	}

	// Delete existing entries
	r.DB.Delete(&models.LeaderboardEntry{}, "leaderboard_id = ?", leaderboardID)

	// Create new entries
	for i, user := range users {
		entry := models.LeaderboardEntry{
			LeaderboardID: leaderboardID,
			UserID:        user.ID,
			Rank:          i + 1,
			XP:            user.TotalXP,
		}
		r.DB.Create(&entry)
	}

	return nil
}

// ============================================
// Streak Methods
// ============================================

// UpdateStreak updates user's streak
func (r *GamificationRepository) UpdateStreak(userID string) (*models.UserStreak, error) {
	var streak models.UserStreak
	err := r.DB.FirstOrCreate(&streak, models.UserStreak{UserID: userID}).Error
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")

	if streak.LastActivityDate == nil {
		// First activity
		streak.CurrentStreak = 1
		streak.LongestStreak = 1
		streak.LastActivityDate = &today
	} else if *streak.LastActivityDate == today {
		// Already active today
		return &streak, nil
	} else {
		// Check if yesterday
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		if *streak.LastActivityDate == yesterday {
			streak.CurrentStreak++
			if streak.CurrentStreak > streak.LongestStreak {
				streak.LongestStreak = streak.CurrentStreak
			}
		} else {
			// Streak broken
			streak.CurrentStreak = 1
		}
		streak.LastActivityDate = &today
	}

	err = r.DB.Save(&streak).Error
	return &streak, err
}

// GetStreak gets user's streak info
func (r *GamificationRepository) GetStreak(userID string) (*models.UserStreak, error) {
	var streak models.UserStreak
	err := r.DB.FirstOrCreate(&streak, models.UserStreak{UserID: userID}).Error
	return &streak, err
}
