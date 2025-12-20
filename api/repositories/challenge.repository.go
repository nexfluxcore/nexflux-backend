package repositories

import (
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// ChallengeRepository handles challenge database operations
type ChallengeRepository struct {
	*BaseRepository
}

// NewChallengeRepository creates a new ChallengeRepository
func NewChallengeRepository(db *gorm.DB) *ChallengeRepository {
	return &ChallengeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// ChallengeFilter defines filter options
type ChallengeFilter struct {
	Type       string
	Difficulty string
	Category   string
	IsActive   bool
}

// FindAll finds all challenges with filters
func (r *ChallengeRepository) FindAll(filter ChallengeFilter, page, limit int) ([]models.Challenge, int64, error) {
	var challenges []models.Challenge
	var total int64

	query := r.DB.Model(&models.Challenge{})

	if filter.IsActive {
		query = query.Where("is_active = ?", true)
	}

	if filter.Type != "" && filter.Type != "all" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.Difficulty != "" {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Order("created_at DESC").
		Find(&challenges).Error

	return challenges, total, err
}

// FindByID finds challenge by ID
func (r *ChallengeRepository) FindByID(id string) (*models.Challenge, error) {
	var challenge models.Challenge
	err := r.DB.First(&challenge, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

// GetDailyChallenge gets today's daily challenge
func (r *ChallengeRepository) GetDailyChallenge() (*models.DailyChallenge, error) {
	today := time.Now().Format("2006-01-02")

	var daily models.DailyChallenge
	err := r.DB.Where("date = ?", today).
		Preload("Challenge").
		First(&daily).Error

	if err != nil {
		return nil, err
	}
	return &daily, nil
}

// GetUserProgress gets user's progress on a challenge
func (r *ChallengeRepository) GetUserProgress(userID, challengeID string) (*models.ChallengeProgress, error) {
	var progress models.ChallengeProgress
	err := r.DB.Where("user_id = ? AND challenge_id = ?", userID, challengeID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetUserProgressList gets user's all challenge progress
func (r *ChallengeRepository) GetUserProgressList(userID string, status string, page, limit int) ([]models.ChallengeProgress, int64, error) {
	var progresses []models.ChallengeProgress
	var total int64

	query := r.DB.Model(&models.ChallengeProgress{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Preload("Challenge").
		Order("updated_at DESC").
		Find(&progresses).Error

	return progresses, total, err
}

// StartChallenge starts a challenge for user
func (r *ChallengeRepository) StartChallenge(userID, challengeID string) (*models.ChallengeProgress, error) {
	now := time.Now()
	progress := &models.ChallengeProgress{
		UserID:      userID,
		ChallengeID: challengeID,
		Status:      models.ProgressInProgress,
		Progress:    0,
		CurrentStep: 0,
		StartedAt:   &now,
	}

	err := r.DB.Create(progress).Error
	if err != nil {
		return nil, err
	}

	// Increment daily challenge participant count if applicable
	today := time.Now().Format("2006-01-02")
	r.DB.Model(&models.DailyChallenge{}).
		Where("challenge_id = ? AND date = ?", challengeID, today).
		UpdateColumn("participant_count", gorm.Expr("participant_count + 1"))

	return progress, nil
}

// UpdateProgress updates challenge progress
func (r *ChallengeRepository) UpdateProgress(progress *models.ChallengeProgress) error {
	return r.DB.Save(progress).Error
}

// CompleteChallenge marks challenge as completed
func (r *ChallengeRepository) CompleteChallenge(userID, challengeID string, xpEarned int) (*models.ChallengeProgress, error) {
	var progress models.ChallengeProgress
	err := r.DB.Where("user_id = ? AND challenge_id = ?", userID, challengeID).First(&progress).Error
	if err != nil {
		return nil, err
	}

	now := time.Now()
	progress.Status = models.ProgressCompleted
	progress.Progress = 100
	progress.XPEarned = &xpEarned
	progress.CompletedAt = &now

	err = r.DB.Save(&progress).Error
	return &progress, err
}

// GetChallengeStats gets user's challenge statistics
func (r *ChallengeRepository) GetChallengeStats(userID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	var completed, inProgress, total int64

	r.DB.Model(&models.ChallengeProgress{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&completed)
	stats["completed"] = completed

	r.DB.Model(&models.ChallengeProgress{}).
		Where("user_id = ? AND status = 'in_progress'", userID).
		Count(&inProgress)
	stats["in_progress"] = inProgress

	r.DB.Model(&models.Challenge{}).
		Where("is_active = true").
		Count(&total)
	stats["total"] = total

	return stats, nil
}

// GetParticipantCount gets challenge participant count
func (r *ChallengeRepository) GetParticipantCount(challengeID string) int64 {
	var count int64
	r.DB.Model(&models.ChallengeProgress{}).
		Where("challenge_id = ?", challengeID).
		Count(&count)
	return count
}

// HasStarted checks if user has started a challenge
func (r *ChallengeRepository) HasStarted(userID, challengeID string) bool {
	var count int64
	r.DB.Model(&models.ChallengeProgress{}).
		Where("user_id = ? AND challenge_id = ?", userID, challengeID).
		Count(&count)
	return count > 0
}
