package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ChallengeHandler handles challenge-related HTTP requests
type ChallengeHandler struct {
	service *services.ChallengeService
}

// NewChallengeHandler creates a new ChallengeHandler
func NewChallengeHandler(db *gorm.DB) *ChallengeHandler {
	return &ChallengeHandler{
		service: services.NewChallengeService(db),
	}
}

// ListChallenges godoc
// @Summary List challenges
// @Description Get list of available challenges with optional filters
// @Tags Challenges
// @Produce json
// @Param type query string false "Challenge type (all, daily, weekly, special, regular)" default(all)
// @Param difficulty query string false "Difficulty filter (Easy, Medium, Hard, Expert)"
// @Param category query string false "Category filter"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Security Bearer
// @Success 200 {object} dto.ChallengeListResponse "List of challenges with stats"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /challenges [get]
func (h *ChallengeHandler) ListChallenges(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.ChallengeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid parameters: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	response, err := h.service.ListChallenges(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetChallenge godoc
// @Summary Get challenge details
// @Description Get detailed information about a specific challenge including instructions
// @Tags Challenges
// @Produce json
// @Param id path string true "Challenge ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.ChallengeDetailResponse "Challenge details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Challenge not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /challenges/{id} [get]
func (h *ChallengeHandler) GetChallenge(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	challengeID := c.Param("id")

	challenge, err := h.service.GetChallenge(challengeID, userID.(string))
	if err != nil {
		if err.Error() == "challenge not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    challenge,
	})
}

// GetDailyChallenge godoc
// @Summary Get today's daily challenge
// @Description Get the daily challenge for today with bonus XP multiplier
// @Tags Challenges
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.DailyChallengeResponse "Daily challenge details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "No daily challenge today"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /challenges/daily [get]
func (h *ChallengeHandler) GetDailyChallenge(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	daily, err := h.service.GetDailyChallenge(userID.(string))
	if err != nil {
		if err.Error() == "no daily challenge today" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    daily,
	})
}

// StartChallenge godoc
// @Summary Start a challenge
// @Description Start working on a challenge (creates progress record)
// @Tags Challenges
// @Param id path string true "Challenge ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.ChallengeProgressResponse "Challenge progress"
// @Failure 400 {object} map[string]string "Challenge not available or already started"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Challenge not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /challenges/{id}/start [post]
func (h *ChallengeHandler) StartChallenge(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	challengeID := c.Param("id")

	progress, err := h.service.StartChallenge(challengeID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Challenge started",
		"data":    progress,
	})
}

// UpdateProgress godoc
// @Summary Update challenge progress
// @Description Update user's progress on an active challenge
// @Tags Challenges
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID (UUID)"
// @Param progress body dto.UpdateProgressRequest true "Progress data"
// @Security Bearer
// @Success 200 {object} dto.ChallengeProgressResponse "Updated progress"
// @Failure 400 {object} map[string]string "Invalid input or challenge not started"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /challenges/{id}/progress [put]
func (h *ChallengeHandler) UpdateProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	challengeID := c.Param("id")

	var req dto.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	progress, err := h.service.UpdateProgress(challengeID, userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    progress,
	})
}

// SubmitChallenge godoc
// @Summary Submit challenge completion
// @Description Submit challenge for completion and earn XP rewards
// @Tags Challenges
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID (UUID)"
// @Param submission body dto.SubmitChallengeRequest true "Submission data"
// @Security Bearer
// @Success 200 {object} dto.SubmitChallengeResponse "Completion result with XP earned"
// @Failure 400 {object} map[string]string "Challenge not started or already completed"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /challenges/{id}/submit [post]
func (h *ChallengeHandler) SubmitChallenge(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	challengeID := c.Param("id")

	var req dto.SubmitChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.SubmitChallenge(challengeID, userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Challenge completed!",
		"data":    result,
	})
}

// GetUserProgress godoc
// @Summary Get my challenge progress
// @Description Get list of challenges the user has started or completed
// @Tags Challenges
// @Produce json
// @Param status query string false "Filter by status (in_progress, completed)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of challenge progress"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /challenges/progress [get]
func (h *ChallengeHandler) GetUserProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := c.Query("status")
	page := 1
	limit := 20

	progresses, pagination, err := h.service.GetUserProgress(userID.(string), status, page, limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"progress":   progresses,
			"pagination": pagination,
		},
	})
}
