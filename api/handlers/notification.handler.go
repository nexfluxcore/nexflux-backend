package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	service *services.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{
		service: services.NewNotificationService(db),
	}
}

// ListNotifications godoc
// @Summary List notifications
// @Description Get user's notifications with optional filters
// @Tags Notifications
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Filter by type (achievement, challenge, system, social)"
// @Param unread_only query bool false "Only unread notifications"
// @Security Bearer
// @Success 200 {object} dto.NotificationListResponse "List of notifications"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications [get]
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.NotificationListRequest
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

	response, err := h.service.ListNotifications(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags Notifications
// @Param id path string true "Notification ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	notificationID := c.Param("id")

	if err := h.service.MarkAsRead(notificationID, userID.(string)); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all user's unread notifications as read
// @Tags Notifications
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.MarkAllAsRead(userID.(string)); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All notifications marked as read",
	})
}

// DeleteNotification godoc
// @Summary Delete notification
// @Description Delete a specific notification
// @Tags Notifications
// @Param id path string true "Notification ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	notificationID := c.Param("id")

	if err := h.service.DeleteNotification(notificationID, userID.(string)); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification deleted",
	})
}
