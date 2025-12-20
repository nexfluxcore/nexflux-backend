package services

import (
	"errors"
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// ChallengeService handles challenge business logic
type ChallengeService struct {
	repo      *repositories.ChallengeRepository
	userRepo  *repositories.UserRepository
	gamifRepo *repositories.GamificationRepository
}

// NewChallengeService creates a new ChallengeService
func NewChallengeService(db *gorm.DB) *ChallengeService {
	return &ChallengeService{
		repo:      repositories.NewChallengeRepository(db),
		userRepo:  repositories.NewUserRepository(db),
		gamifRepo: repositories.NewGamificationRepository(db),
	}
}

// ListChallenges lists challenges with filters
func (s *ChallengeService) ListChallenges(userID string, req dto.ChallengeListRequest) (*dto.ChallengeListResponse, error) {
	filter := repositories.ChallengeFilter{
		Type:       req.Type,
		Difficulty: req.Difficulty,
		Category:   req.Category,
		IsActive:   true,
	}

	challenges, _, err := s.repo.FindAll(filter, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ChallengeResponse, len(challenges))
	for i, c := range challenges {
		responses[i] = s.toChallengeResponse(&c, userID)
	}

	stats, _ := s.repo.GetChallengeStats(userID)

	return &dto.ChallengeListResponse{
		Challenges: responses,
		Stats: dto.ChallengeStats{
			Completed:      int(stats["completed"]),
			InProgress:     int(stats["in_progress"]),
			TotalAvailable: int(stats["total"]),
		},
	}, nil
}

// GetChallenge gets challenge by ID
func (s *ChallengeService) GetChallenge(challengeID, userID string) (*dto.ChallengeDetailResponse, error) {
	challenge, err := s.repo.FindByID(challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("challenge not found")
		}
		return nil, err
	}

	base := s.toChallengeResponse(challenge, userID)

	return &dto.ChallengeDetailResponse{
		ChallengeResponse:  base,
		Prerequisites:      challenge.Prerequisites,
		Instructions:       challenge.Instructions,
		ValidationCriteria: challenge.ValidationCriteria,
	}, nil
}

// GetDailyChallenge gets today's daily challenge
func (s *ChallengeService) GetDailyChallenge(userID string) (*dto.DailyChallengeResponse, error) {
	daily, err := s.repo.GetDailyChallenge()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no daily challenge today")
		}
		return nil, err
	}

	challengeResp := s.toChallengeResponse(daily.Challenge, userID)

	// Calculate time left (end of day)
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	timeLeft := endOfDay.Sub(now).Hours()

	return &dto.DailyChallengeResponse{
		ID:               daily.ID,
		Challenge:        challengeResp,
		Date:             daily.Date,
		XPMultiplier:     daily.XPMultiplier,
		TimeLeftHours:    timeLeft,
		ParticipantCount: daily.ParticipantCount,
		UserStatus:       challengeResp.UserStatus,
	}, nil
}

// StartChallenge starts a challenge for user
func (s *ChallengeService) StartChallenge(challengeID, userID string) (*dto.ChallengeProgressResponse, error) {
	// Check if challenge exists
	challenge, err := s.repo.FindByID(challengeID)
	if err != nil {
		return nil, errors.New("challenge not found")
	}

	// Check if challenge is available
	if !challenge.IsAvailable() {
		return nil, errors.New("challenge is not available")
	}

	// Check if already started
	if s.repo.HasStarted(userID, challengeID) {
		progress, _ := s.repo.GetUserProgress(userID, challengeID)
		return s.toProgressResponse(progress), nil
	}

	// Start the challenge
	progress, err := s.repo.StartChallenge(userID, challengeID)
	if err != nil {
		return nil, err
	}

	// Update user streak
	s.gamifRepo.UpdateStreak(userID)

	return s.toProgressResponse(progress), nil
}

// UpdateProgress updates challenge progress
func (s *ChallengeService) UpdateProgress(challengeID, userID string, req dto.UpdateProgressRequest) (*dto.ChallengeProgressResponse, error) {
	progress, err := s.repo.GetUserProgress(userID, challengeID)
	if err != nil {
		return nil, errors.New("challenge not started")
	}

	if progress.Status == models.ProgressCompleted {
		return nil, errors.New("challenge already completed")
	}

	progress.Progress = req.Progress
	progress.CurrentStep = req.CurrentStep
	progress.TimeSpentMinutes = req.TimeSpent

	if err := s.repo.UpdateProgress(progress); err != nil {
		return nil, err
	}

	return s.toProgressResponse(progress), nil
}

// SubmitChallenge submits challenge completion
func (s *ChallengeService) SubmitChallenge(challengeID, userID string, req dto.SubmitChallengeRequest) (*dto.SubmitChallengeResponse, error) {
	progress, err := s.repo.GetUserProgress(userID, challengeID)
	if err != nil {
		return nil, errors.New("challenge not started")
	}

	if progress.Status == models.ProgressCompleted {
		return nil, errors.New("challenge already completed")
	}

	challenge, err := s.repo.FindByID(challengeID)
	if err != nil {
		return nil, err
	}

	// Update progress with submission
	progress.SubmissionData = req.SubmissionData
	progress.Feedback = req.Notes

	// Calculate XP (could add bonus logic here)
	baseXP := challenge.XPReward
	bonusXP := 0

	// Check for daily challenge multiplier
	daily, _ := s.repo.GetDailyChallenge()
	if daily != nil && daily.ChallengeID == challengeID {
		bonusXP = int(float64(baseXP) * (daily.XPMultiplier - 1))
	}

	totalXP := baseXP + bonusXP

	// Complete the challenge
	_, err = s.repo.CompleteChallenge(userID, challengeID, totalXP)
	if err != nil {
		return nil, err
	}

	// Add XP to user
	user, levelUp, err := s.userRepo.AddXP(userID, totalXP)
	if err != nil {
		return nil, err
	}

	// Check for new achievements (simplified)
	var unlockedAchievements []dto.AchievementResponse
	// In a real implementation, we'd check achievement criteria here

	return &dto.SubmitChallengeResponse{
		XPEarned:             baseXP,
		BonusXP:              bonusXP,
		TotalXP:              totalXP,
		NewLevel:             user.Level,
		LevelUp:              levelUp,
		AchievementsUnlocked: unlockedAchievements,
	}, nil
}

// GetUserProgress gets user's challenge progress list
func (s *ChallengeService) GetUserProgress(userID, status string, page, limit int) ([]dto.ChallengeProgressResponse, dto.PaginationResponse, error) {
	progresses, total, err := s.repo.GetUserProgressList(userID, status, page, limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.ChallengeProgressResponse, len(progresses))
	for i, p := range progresses {
		responses[i] = *s.toProgressResponse(&p)
	}

	pagination := dto.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	return responses, pagination, nil
}

// Helper functions

func (s *ChallengeService) toChallengeResponse(c *models.Challenge, userID string) dto.ChallengeResponse {
	response := dto.ChallengeResponse{
		ID:               c.ID,
		Title:            c.Title,
		Description:      c.Description,
		Difficulty:       string(c.Difficulty),
		XPReward:         c.XPReward,
		Category:         c.Category,
		Type:             string(c.Type),
		TimeLimitHours:   c.TimeLimitHours,
		MaxParticipants:  c.MaxParticipants,
		IsActive:         c.IsActive,
		StartsAt:         c.StartsAt,
		EndsAt:           c.EndsAt,
		ParticipantCount: int(s.repo.GetParticipantCount(c.ID)),
	}

	// Get time remaining
	if remaining := c.TimeRemaining(); remaining != nil {
		hours := remaining.Hours()
		response.TimeLeftHours = &hours
	}

	// Get user's status on this challenge
	if userID != "" {
		progress, err := s.repo.GetUserProgress(userID, c.ID)
		if err == nil {
			response.UserStatus = string(progress.Status)
			response.UserProgress = progress.Progress
		} else {
			response.UserStatus = "not_started"
			response.UserProgress = 0
		}
	}

	return response
}

func (s *ChallengeService) toProgressResponse(p *models.ChallengeProgress) *dto.ChallengeProgressResponse {
	response := &dto.ChallengeProgressResponse{
		ID:               p.ID,
		ChallengeID:      p.ChallengeID,
		Progress:         p.Progress,
		Status:           string(p.Status),
		CurrentStep:      p.CurrentStep,
		XPEarned:         p.XPEarned,
		TimeSpentMinutes: p.TimeSpentMinutes,
		StartedAt:        p.StartedAt,
		CompletedAt:      p.CompletedAt,
	}

	if p.Challenge != nil {
		challengeResp := s.toChallengeResponse(p.Challenge, "")
		response.Challenge = &challengeResp
	}

	return response
}
