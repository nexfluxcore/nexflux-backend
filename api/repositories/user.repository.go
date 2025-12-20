package repositories

import (
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// UserRepository handles user database operations
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// FindByID finds user by ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.DB.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.DB.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByProvider finds user by OAuth provider
func (r *UserRepository) FindByProvider(provider, providerID string) (*models.User, error) {
	var user models.User
	err := r.DB.First(&user, "provider = ? AND provider_id = ?", provider, providerID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.DB.Save(user).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(id string) error {
	return r.DB.Delete(&models.User{}, "id = ?", id).Error
}

// GetSettings gets user settings
func (r *UserRepository) GetSettings(userID string) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := r.DB.FirstOrCreate(&settings, models.UserSettings{UserID: userID}).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateSettings updates user settings
func (r *UserRepository) UpdateSettings(settings *models.UserSettings) error {
	return r.DB.Save(settings).Error
}

// GetStreak gets user streak
func (r *UserRepository) GetStreak(userID string) (*models.UserStreak, error) {
	var streak models.UserStreak
	err := r.DB.FirstOrCreate(&streak, models.UserStreak{UserID: userID}).Error
	if err != nil {
		return nil, err
	}
	return &streak, nil
}

// UpdateStreak updates user streak
func (r *UserRepository) UpdateStreak(streak *models.UserStreak) error {
	return r.DB.Save(streak).Error
}

// AddXP adds XP to user and handles level up
func (r *UserRepository) AddXP(userID string, xp int) (*models.User, bool, error) {
	var user models.User
	err := r.DB.First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, false, err
	}

	user.TotalXP += xp
	user.CurrentXP += xp

	levelUp := false
	for user.CurrentXP >= user.TargetXP {
		user.CurrentXP -= user.TargetXP
		user.Level++
		user.TargetXP = calculateNextLevelXP(user.Level)
		levelUp = true
	}

	err = r.DB.Save(&user).Error
	return &user, levelUp, err
}

// calculateNextLevelXP calculates XP needed for next level
func calculateNextLevelXP(level int) int {
	// Formula: 1000 * level * 1.1
	return int(float64(1000*level) * 1.1)
}

// GetUserStats gets user statistics
func (r *UserRepository) GetUserStats(userID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get user
	var user models.User
	if err := r.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	// Count completed projects
	var projectsCompleted int64
	r.DB.Model(&models.Project{}).Where("user_id = ? AND progress = 100", userID).Count(&projectsCompleted)

	// Count projects in progress
	var projectsInProgress int64
	r.DB.Model(&models.Project{}).Where("user_id = ? AND progress < 100", userID).Count(&projectsInProgress)

	// Count completed challenges
	var challengesCompleted int64
	r.DB.Model(&models.ChallengeProgress{}).Where("user_id = ? AND status = 'completed'", userID).Count(&challengesCompleted)

	// Count challenges in progress
	var challengesInProgress int64
	r.DB.Model(&models.ChallengeProgress{}).Where("user_id = ? AND status = 'in_progress'", userID).Count(&challengesInProgress)

	// Count achievements
	var achievementsUnlocked int64
	r.DB.Model(&models.UserAchievement{}).Where("user_id = ?", userID).Count(&achievementsUnlocked)

	var totalAchievements int64
	r.DB.Model(&models.Achievement{}).Where("is_active = true").Count(&totalAchievements)

	// Get streak
	streak, _ := r.GetStreak(userID)

	stats["level"] = user.Level
	stats["total_xp"] = user.TotalXP
	stats["current_xp"] = user.CurrentXP
	stats["target_xp"] = user.TargetXP
	stats["xp_to_next_level"] = user.TargetXP - user.CurrentXP
	stats["projects_total"] = projectsCompleted + projectsInProgress
	stats["projects_completed"] = projectsCompleted
	stats["projects_in_progress"] = projectsInProgress
	stats["challenges_completed"] = challengesCompleted
	stats["challenges_in_progress"] = challengesInProgress
	stats["achievements_unlocked"] = achievementsUnlocked
	stats["achievements_total"] = totalAchievements

	if streak != nil {
		stats["streak_current"] = streak.CurrentStreak
		stats["streak_longest"] = streak.LongestStreak
		stats["streak_last_active"] = streak.LastActivityDate
	}

	return stats, nil
}
