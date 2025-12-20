package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GamificationHandler handles gamification-related HTTP requests
type GamificationHandler struct {
	service *services.GamificationService
}

// NewGamificationHandler creates a new GamificationHandler
func NewGamificationHandler(db *gorm.DB) *GamificationHandler {
	return &GamificationHandler{
		service: services.NewGamificationService(db),
	}
}

// GetAllAchievements godoc
// @Summary Get all achievements
// @Description Get list of all achievements with user's unlock status and progress
// @Tags Gamification
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.AchievementListResponse "List of achievements with stats"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /achievements [get]
func (h *GamificationHandler) GetAllAchievements(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response, err := h.service.GetAllAchievements(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetUserAchievements godoc
// @Summary Get my achievements
// @Description Get list of achievements unlocked by the current user
// @Tags Gamification
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of unlocked achievements"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /achievements/user [get]
func (h *GamificationHandler) GetUserAchievements(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	achievements, err := h.service.GetUserAchievements(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    achievements,
	})
}

// GetLeaderboard godoc
// @Summary Get leaderboard
// @Description Get leaderboard rankings (weekly, monthly, or all-time)
// @Tags Gamification
// @Produce json
// @Param type query string false "Leaderboard type (weekly, monthly, all_time)" default(weekly)
// @Param limit query int false "Number of entries to return" default(10)
// @Security Bearer
// @Success 200 {object} dto.LeaderboardResponse "Leaderboard with rankings"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /leaderboard [get]
func (h *GamificationHandler) GetLeaderboard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.LeaderboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid parameters: "+err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "weekly"
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	response, err := h.service.GetLeaderboard(req, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetStreak godoc
// @Summary Get streak info
// @Description Get user's current streak information and history
// @Tags Gamification
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.StreakResponse "Streak information"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /streak [get]
func (h *GamificationHandler) GetStreak(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	streak, err := h.service.GetStreak(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    streak,
	})
}

// GetUserStats godoc
// @Summary Get user stats
// @Description Get comprehensive user gamification statistics including XP, level, streak, etc.
// @Tags Gamification
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.UserStatsResponse "User statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me/stats [get]
func (h *GamificationHandler) GetUserStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.service.GetUserStats(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
