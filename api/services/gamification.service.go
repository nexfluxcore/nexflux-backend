package services

import (
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// GamificationService handles gamification business logic
type GamificationService struct {
	repo     *repositories.GamificationRepository
	userRepo *repositories.UserRepository
}

// NewGamificationService creates a new GamificationService
func NewGamificationService(db *gorm.DB) *GamificationService {
	return &GamificationService{
		repo:     repositories.NewGamificationRepository(db),
		userRepo: repositories.NewUserRepository(db),
	}
}

// ============================================
// Achievement Methods
// ============================================

// GetAllAchievements gets all achievements with user's unlock status
func (s *GamificationService) GetAllAchievements(userID string) (*dto.AchievementListResponse, error) {
	achievements, err := s.repo.GetAllAchievements()
	if err != nil {
		return nil, err
	}

	// Get user's unlocked achievements
	userAchievements, err := s.repo.GetUserAchievements(userID)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup
	unlocked := make(map[string]time.Time)
	for _, ua := range userAchievements {
		unlocked[ua.AchievementID] = ua.UnlockedAt
	}

	// Get user's progress towards achievements
	progress, _ := s.repo.GetAchievementProgress(userID)

	responses := make([]dto.AchievementResponse, len(achievements))
	for i, a := range achievements {
		unlockedAt, isUnlocked := unlocked[a.ID]

		response := dto.AchievementResponse{
			ID:               a.ID,
			Name:             a.Name,
			Description:      a.Description,
			Icon:             a.Icon,
			Rarity:           string(a.Rarity),
			RequirementType:  a.RequirementType,
			RequirementValue: a.RequirementValue,
			XPReward:         a.XPReward,
			IsUnlocked:       isUnlocked,
		}

		if isUnlocked {
			response.UnlockedAt = &unlockedAt
		}

		// Calculate progress
		if !isUnlocked && a.RequirementType != "" {
			if currentProgress, ok := progress[a.RequirementType]; ok {
				percentage := int(float64(currentProgress) / float64(a.RequirementValue) * 100)
				if percentage > 100 {
					percentage = 100
				}
				response.Progress = percentage
			}
		}

		responses[i] = response
	}

	// Count unlocked
	unlockedCount := len(unlocked)

	return &dto.AchievementListResponse{
		Achievements: responses,
		Stats: struct {
			Unlocked int `json:"unlocked"`
			Total    int `json:"total"`
		}{
			Unlocked: unlockedCount,
			Total:    len(achievements),
		},
	}, nil
}

// GetUserAchievements gets user's unlocked achievements
func (s *GamificationService) GetUserAchievements(userID string) ([]dto.AchievementResponse, error) {
	userAchievements, err := s.repo.GetUserAchievements(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AchievementResponse, len(userAchievements))
	for i, ua := range userAchievements {
		a := ua.Achievement
		responses[i] = dto.AchievementResponse{
			ID:          a.ID,
			Name:        a.Name,
			Description: a.Description,
			Icon:        a.Icon,
			Rarity:      string(a.Rarity),
			XPReward:    a.XPReward,
			IsUnlocked:  true,
			UnlockedAt:  &ua.UnlockedAt,
		}
	}

	return responses, nil
}

// CheckAndUnlockAchievements checks if user has earned any new achievements
func (s *GamificationService) CheckAndUnlockAchievements(userID string) ([]dto.AchievementResponse, error) {
	achievements, err := s.repo.GetAllAchievements()
	if err != nil {
		return nil, err
	}

	progress, err := s.repo.GetAchievementProgress(userID)
	if err != nil {
		return nil, err
	}

	var newUnlocks []dto.AchievementResponse

	for _, a := range achievements {
		// Skip if already unlocked
		if s.repo.HasAchievement(userID, a.ID) {
			continue
		}

		// Check if requirement is met
		if a.RequirementType != "" && a.RequirementValue > 0 {
			if currentProgress, ok := progress[a.RequirementType]; ok {
				if currentProgress >= a.RequirementValue {
					// Unlock achievement
					ua, err := s.repo.UnlockAchievement(userID, a.ID)
					if err == nil && ua != nil {
						// Add XP reward
						s.userRepo.AddXP(userID, a.XPReward)

						newUnlocks = append(newUnlocks, dto.AchievementResponse{
							ID:          a.ID,
							Name:        a.Name,
							Description: a.Description,
							Icon:        a.Icon,
							Rarity:      string(a.Rarity),
							XPReward:    a.XPReward,
							IsUnlocked:  true,
							UnlockedAt:  &ua.UnlockedAt,
						})
					}
				}
			}
		}
	}

	return newUnlocks, nil
}

// ============================================
// Leaderboard Methods
// ============================================

// GetLeaderboard gets leaderboard
func (s *GamificationService) GetLeaderboard(req dto.LeaderboardRequest, userID string) (*dto.LeaderboardResponse, error) {
	leaderboard, err := s.repo.GetOrCreateLeaderboard(req.Type)
	if err != nil {
		return nil, err
	}

	entries, err := s.repo.GetLeaderboardEntries(leaderboard.ID, req.Limit)
	if err != nil {
		return nil, err
	}

	entryResponses := make([]dto.LeaderboardEntryResponse, len(entries))
	for i, e := range entries {
		entryResponses[i] = dto.LeaderboardEntryResponse{
			Rank: e.Rank,
			User: dto.UserBriefResponse{
				ID:        e.User.ID,
				Name:      e.User.Name,
				Username:  e.User.Username,
				AvatarURL: e.User.AvatarURL,
				Level:     e.User.Level,
			},
			XP:            e.XP,
			IsCurrentUser: e.UserID == userID,
		}
	}

	userRank, _ := s.repo.GetUserRank(leaderboard.ID, userID)

	response := &dto.LeaderboardResponse{
		Type:            req.Type,
		Entries:         entryResponses,
		CurrentUserRank: userRank,
	}

	if leaderboard.PeriodStart != nil && leaderboard.PeriodEnd != nil {
		response.Period = &dto.PeriodResponse{
			Start: *leaderboard.PeriodStart,
			End:   *leaderboard.PeriodEnd,
		}
	}

	return response, nil
}

// ============================================
// Streak Methods
// ============================================

// GetStreak gets user's streak info
func (s *GamificationService) GetStreak(userID string) (*dto.StreakResponse, error) {
	streak, err := s.repo.GetStreak(userID)
	if err != nil {
		return nil, err
	}

	lastActive := ""
	if streak.LastActivityDate != nil {
		lastActive = *streak.LastActivityDate
	}

	// Check if active today
	today := time.Now().Format("2006-01-02")
	isActiveToday := lastActive == today

	return &dto.StreakResponse{
		CurrentStreak:    streak.CurrentStreak,
		LongestStreak:    streak.LongestStreak,
		LastActivityDate: lastActive,
		StreakProtects:   streak.StreakProtects,
		IsActiveToday:    isActiveToday,
	}, nil
}

// UpdateStreak updates user's streak
func (s *GamificationService) UpdateStreak(userID string) (*models.UserStreak, error) {
	return s.repo.UpdateStreak(userID)
}

// ============================================
// User Stats Methods
// ============================================

// GetUserStats gets comprehensive user stats
func (s *GamificationService) GetUserStats(userID string) (*dto.UserStatsResponse, error) {
	stats, err := s.userRepo.GetUserStats(userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserStatsResponse{
		Level:         stats["level"].(int),
		TotalXP:       stats["total_xp"].(int),
		CurrentXP:     stats["current_xp"].(int),
		TargetXP:      stats["target_xp"].(int),
		XPToNextLevel: stats["xp_to_next_level"].(int),
		Streak: dto.StreakInfo{
			Current:    stats["streak_current"].(int),
			Longest:    stats["streak_longest"].(int),
			LastActive: getStringFromMap(stats, "streak_last_active"),
		},
		Challenges: dto.ChallengeStats{
			Completed:      int(stats["challenges_completed"].(int64)),
			InProgress:     int(stats["challenges_in_progress"].(int64)),
			TotalAvailable: 0, // Would need to query
		},
		Projects: dto.ProjectStats{
			Total:      int(stats["projects_total"].(int64)),
			Completed:  int(stats["projects_completed"].(int64)),
			InProgress: int(stats["projects_in_progress"].(int64)),
		},
		Achievements: dto.AchievementStats{
			Unlocked: int(stats["achievements_unlocked"].(int64)),
			Total:    int(stats["achievements_total"].(int64)),
		},
	}, nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(*string); ok && s != nil {
			return *s
		}
	}
	return ""
}
