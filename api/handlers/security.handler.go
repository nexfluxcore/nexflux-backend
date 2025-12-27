package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SecurityHandler handles security-related HTTP requests
type SecurityHandler struct {
	service *services.SecurityService
}

// NewSecurityHandler creates a new SecurityHandler
func NewSecurityHandler(db *gorm.DB) *SecurityHandler {
	return &SecurityHandler{
		service: services.NewSecurityService(db),
	}
}

// ============================================
// Password Management
// ============================================

// ChangePassword godoc
// @Summary Change password
// @Description Change user's password
// @Tags Security
// @Accept json
// @Produce json
// @Param body body dto.ChangePasswordRequest true "Password change request"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Password changed"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Current password incorrect"
// @Failure 422 {object} map[string]string "Same password"
// @Router /auth/password [put]
func (h *SecurityHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.service.ChangePassword(userID.(string), req, ipAddress, userAgent); err != nil {
		switch err.Error() {
		case "current password is incorrect":
			utils.RespondWithError(c, http.StatusUnauthorized, err.Error())
		case "new password cannot be the same as current password":
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "code": "SAME_PASSWORD"})
		case "password has been used recently, please choose a different one":
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "code": "PASSWORD_REUSED"})
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// ============================================
// Two-Factor Authentication
// ============================================

// Enable2FA godoc
// @Summary Enable 2FA
// @Description Enable two-factor authentication
// @Tags Security
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.Enable2FAResponse "2FA enabled"
// @Failure 400 {object} map[string]string "Already enabled"
// @Router /auth/2fa/enable [post]
func (h *SecurityHandler) Enable2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user email from context or fetch from DB
	email, _ := c.Get("email")
	emailStr := ""
	if email != nil {
		emailStr = email.(string)
	}

	result, err := h.service.Enable2FA(userID.(string), emailStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Two-factor authentication enabled. Please save your backup codes.",
	})
}

// Verify2FA godoc
// @Summary Verify 2FA
// @Description Verify 2FA code during setup
// @Tags Security
// @Accept json
// @Produce json
// @Param body body dto.Verify2FARequest true "Verification code"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "2FA verified"
// @Failure 400 {object} map[string]string "Invalid code"
// @Router /auth/2fa/verify [post]
func (h *SecurityHandler) Verify2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.service.Verify2FA(userID.(string), req.Code, ipAddress, userAgent); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Two-factor authentication verified and activated",
	})
}

// Disable2FA godoc
// @Summary Disable 2FA
// @Description Disable two-factor authentication
// @Tags Security
// @Accept json
// @Produce json
// @Param body body dto.Disable2FARequest true "Password and code"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "2FA disabled"
// @Failure 400 {object} map[string]string "Invalid code or password"
// @Router /auth/2fa/disable [post]
func (h *SecurityHandler) Disable2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.service.Disable2FA(userID.(string), req, ipAddress, userAgent); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Two-factor authentication disabled",
	})
}

// Get2FAStatus godoc
// @Summary Get 2FA status
// @Description Check 2FA status
// @Tags Security
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.TwoFAStatusResponse "2FA status"
// @Router /auth/2fa/status [get]
func (h *SecurityHandler) Get2FAStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status, err := h.service.Get2FAStatus(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
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

// GetSessions godoc
// @Summary Get active sessions
// @Description Get all active sessions for the user
// @Tags Security
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Active sessions"
// @Router /auth/sessions [get]
func (h *SecurityHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get current session ID from context (set by middleware)
	currentSessionID := ""
	if sid, exists := c.Get("sessionID"); exists {
		currentSessionID = sid.(string)
	}

	sessions, err := h.service.GetSessions(userID.(string), currentSessionID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessions,
	})
}

// RevokeSession godoc
// @Summary Revoke session
// @Description Revoke a specific session
// @Tags Security
// @Param id path string true "Session ID"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Session revoked"
// @Failure 400 {object} map[string]string "Cannot revoke current session"
// @Router /auth/sessions/{id} [delete]
func (h *SecurityHandler) RevokeSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionID := c.Param("id")

	// Check if trying to revoke current session
	if currentSessionID, exists := c.Get("sessionID"); exists && currentSessionID == sessionID {
		utils.RespondWithError(c, http.StatusBadRequest, "Cannot revoke current session. Use logout instead.")
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.service.RevokeSession(userID.(string), sessionID, ipAddress, userAgent); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessions godoc
// @Summary Revoke all sessions
// @Description Revoke all sessions except current
// @Tags Security
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Sessions revoked"
// @Router /auth/sessions [delete]
func (h *SecurityHandler) RevokeAllSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	currentSessionID := ""
	if sid, exists := c.Get("sessionID"); exists {
		currentSessionID = sid.(string)
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	count, err := h.service.RevokeAllSessions(userID.(string), currentSessionID, ipAddress, userAgent)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All other sessions have been revoked",
		"data": gin.H{
			"revoked_count": count,
		},
	})
}

// ============================================
// Login History
// ============================================

// GetLoginHistory godoc
// @Summary Get login history
// @Description Get recent login attempts
// @Tags Security
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Login history"
// @Router /auth/login-history [get]
func (h *SecurityHandler) GetLoginHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.LoginHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = dto.LoginHistoryRequest{Page: 1, Limit: 20}
	}

	history, pagination, err := h.service.GetLoginHistory(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       history,
		"pagination": pagination,
	})
}

// ============================================
// Security Settings
// ============================================

// GetSecuritySettings godoc
// @Summary Get security settings
// @Description Get overall security settings for the user
// @Tags Security
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.SecuritySettingsResponse "Security settings"
// @Router /auth/security [get]
func (h *SecurityHandler) GetSecuritySettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	settings, err := h.service.GetSecuritySettings(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}
