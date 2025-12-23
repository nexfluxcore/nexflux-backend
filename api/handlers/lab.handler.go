package handlers

import (
	"fmt"
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LabHandler handles lab-related HTTP requests
type LabHandler struct {
	service *services.LabService
}

// NewLabHandler creates a new LabHandler
func NewLabHandler(db *gorm.DB) *LabHandler {
	return &LabHandler{
		service: services.NewLabService(db),
	}
}

// ============================================
// Lab Management
// ============================================

// ListLabs godoc
// @Summary List available labs
// @Description Get list of available labs with optional filters
// @Tags Labs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param platform query string false "Filter by platform (arduino_uno, esp32, etc.)"
// @Param status query string false "Filter by status (available, busy, maintenance, offline)"
// @Param search query string false "Search by name or description"
// @Success 200 {object} map[string]interface{} "List of labs with pagination"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs [get]
func (h *LabHandler) ListLabs(c *gin.Context) {
	var req dto.LabListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid parameters: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	labs, pagination, err := h.service.ListLabs(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"labs":  labs,
			"total": pagination.Total,
			"page":  pagination.Page,
			"limit": pagination.Limit,
		},
	})
}

// GetLab godoc
// @Summary Get lab details
// @Description Get detailed information about a specific lab including hardware specs and queue
// @Tags Labs
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Success 200 {object} dto.LabDetailResponse "Lab details"
// @Failure 404 {object} map[string]string "Lab not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id} [get]
func (h *LabHandler) GetLab(c *gin.Context) {
	labID := c.Param("id")

	lab, err := h.service.GetLab(labID)
	if err != nil {
		if err.Error() == "lab not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    lab,
	})
}

// GetLabBySlug godoc
// @Summary Get lab details by slug
// @Description Get detailed information about a specific lab by its slug
// @Tags Labs
// @Produce json
// @Param slug path string true "Lab Slug"
// @Success 200 {object} dto.LabDetailResponse "Lab details"
// @Failure 404 {object} map[string]string "Lab not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/slug/{slug} [get]
func (h *LabHandler) GetLabBySlug(c *gin.Context) {
	slug := c.Param("slug")

	lab, err := h.service.GetLabBySlug(slug)
	if err != nil {
		if err.Error() == "lab not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    lab,
	})
}

// ============================================
// Queue Management
// ============================================

// JoinQueue godoc
// @Summary Join lab queue
// @Description Join the queue for a specific lab to reserve a session
// @Tags Labs
// @Accept json
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Param body body dto.JoinQueueRequest true "Queue join request"
// @Security Bearer
// @Success 200 {object} dto.JoinQueueResponse "Queue join result"
// @Failure 400 {object} map[string]string "Invalid input or already in queue"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Lab not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/queue [post]
func (h *LabHandler) JoinQueue(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	var req dto.JoinQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body (bid_amount defaults to 0)
		req = dto.JoinQueueRequest{BidAmount: 0}
	}

	result, err := h.service.JoinQueue(labID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "lab not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "already in queue for this lab", "you already have an active lab session",
			"lab is not accepting queue entries", "insufficient XP for bid amount":
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// LeaveQueue godoc
// @Summary Leave lab queue
// @Description Leave the queue for a specific lab
// @Tags Labs
// @Param id path string true "Lab ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Successfully left queue"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not in queue"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/queue [delete]
func (h *LabHandler) LeaveQueue(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	if err := h.service.LeaveQueue(labID, userID.(string)); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully left queue",
	})
}

// GetQueueStatus godoc
// @Summary Get queue status
// @Description Get user's current queue status for a lab
// @Tags Labs
// @Param id path string true "Lab ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.QueueStatusResponse "Queue status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not in queue"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/queue/status [get]
func (h *LabHandler) GetQueueStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	status, err := h.service.GetQueueStatus(labID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// ============================================
// Session Management
// ============================================

// StartSession godoc
// @Summary Start lab session
// @Description Start a lab session when the lab is available or user is first in queue
// @Tags Labs
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.StartSessionResponse "Session started successfully"
// @Failure 400 {object} map[string]string "Lab not available or already has session"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Lab not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/session/start [post]
func (h *LabHandler) StartSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	result, err := h.service.StartSession(labID, userID.(string))
	if err != nil {
		switch err.Error() {
		case "lab not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "lab is not available, please join the queue",
			"you already have an active lab session":
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// EndSession godoc
// @Summary End lab session
// @Description End the current lab session
// @Tags Labs
// @Accept json
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Param body body dto.EndSessionRequest true "Session end request"
// @Security Bearer
// @Success 200 {object} dto.EndSessionResponse "Session ended successfully"
// @Failure 400 {object} map[string]string "No active session"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/session/end [post]
func (h *LabHandler) EndSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	var req dto.EndSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = dto.EndSessionRequest{}
	}

	result, err := h.service.EndSession(labID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "no active session found":
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetActiveSession godoc
// @Summary Get active session
// @Description Get user's current active lab session
// @Tags Labs
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.ActiveSessionResponse "Active session details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "No active session"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/session/active [get]
func (h *LabHandler) GetActiveSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session, err := h.service.GetActiveSession(userID.(string))
	if err != nil {
		if err.Error() == "no active session found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

// GetSessionHistory godoc
// @Summary Get session history
// @Description Get user's lab session history
// @Tags Labs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Session history with pagination"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/session/history [get]
func (h *LabHandler) GetSessionHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	sessions, pagination, err := h.service.GetUserSessionHistory(userID.(string), page, limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sessions":   sessions,
			"pagination": pagination,
		},
	})
}

// ============================================
// Code Execution
// ============================================

// SubmitCode godoc
// @Summary Submit code for compilation
// @Description Submit code to be compiled and uploaded to the lab hardware
// @Tags Labs
// @Accept json
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Param body body dto.SubmitCodeRequest true "Code submission request"
// @Security Bearer
// @Success 200 {object} dto.SubmitCodeResponse "Code submitted successfully"
// @Failure 400 {object} map[string]string "Invalid input or no active session"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/code/submit [post]
func (h *LabHandler) SubmitCode(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	var req dto.SubmitCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.SubmitCode(labID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "session not found", "session is not active":
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetCompilationStatus godoc
// @Summary Get compilation status
// @Description Get the status of a code compilation job
// @Tags Labs
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Param compilation_id path string true "Compilation ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.CompilationStatusResponse "Compilation status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Compilation not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/code/status/{compilation_id} [get]
func (h *LabHandler) GetCompilationStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")
	compilationID := c.Param("compilation_id")

	result, err := h.service.GetCompilationStatus(labID, compilationID, userID.(string))
	if err != nil {
		switch err.Error() {
		case "compilation not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ============================================
// Sensor/Actuator Control
// ============================================

// GetSensors godoc
// @Summary Get sensor data
// @Description Get current sensor readings from the lab hardware
// @Tags Labs
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.SensorDataResponse "Current sensor data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "No active session"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/sensors [get]
func (h *LabHandler) GetSensors(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	result, err := h.service.GetSensorData(labID, userID.(string))
	if err != nil {
		switch err.Error() {
		case "no active session on this lab":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ControlActuator godoc
// @Summary Control actuator
// @Description Send a control command to a lab actuator (LED, servo, buzzer, etc.)
// @Tags Labs
// @Accept json
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Param body body dto.ControlActuatorRequest true "Actuator control request"
// @Security Bearer
// @Success 200 {object} dto.ControlActuatorResponse "Actuator control result"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "No active session"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/actuators/control [post]
func (h *LabHandler) ControlActuator(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	var req dto.ControlActuatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.ControlActuator(labID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "no active session on this lab":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// SendSerialCommand godoc
// @Summary Send serial command
// @Description Send a serial command directly to the microcontroller
// @Tags Labs
// @Accept json
// @Produce json
// @Param id path string true "Lab ID (UUID)"
// @Param body body dto.SerialCommandRequest true "Serial command request"
// @Security Bearer
// @Success 200 {object} dto.SerialCommandResponse "Serial command result"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "No active session"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /labs/{id}/serial [post]
func (h *LabHandler) SendSerialCommand(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	labID := c.Param("id")

	var req dto.SerialCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.SendSerialCommand(labID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "no active session on this lab":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// Helper function
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
