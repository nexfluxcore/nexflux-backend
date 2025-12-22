package handlers

import (
	"net/http"
	"nexfi-backend/api/repositories"
	"nexfi-backend/database"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"nexfi-backend/pkg/storage"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	repo *repositories.UserRepository
}

func NewUserHandler(repo *repositories.UserRepository) *UserHandler {
	// If no repo provided, create one with global DB
	if repo == nil {
		repo = repositories.NewUserRepository(database.DB)
	}
	return &UserHandler{repo: repo}
}

// GetUserProfile godoc
// @Summary Get my profile
// @Description Get the profile details of the authenticated user
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.UserResponse "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.repo.FindByID(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	response := dto.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Username:      user.Username,
		Role:          string(user.Role),
		AvatarURL:     user.AvatarURL,
		Bio:           user.Bio,
		Level:         user.Level,
		TotalXP:       user.TotalXP,
		CurrentXP:     user.CurrentXP,
		TargetXP:      user.TargetXP,
		StreakDays:    user.StreakDays,
		Language:      user.Language,
		Theme:         string(user.Theme),
		IsPro:         user.IsPro,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateUserProfile godoc
// @Summary Update my profile
// @Description Update the profile details of the authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param input body dto.UserUpdateRequest true "Profile update data"
// @Success 200 {object} dto.UserResponse "Updated profile"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 409 {object} map[string]string "Username already taken"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	user, err := h.repo.FindByID(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	// Update fields
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Username != "" {
		// Check if username is taken
		if existing, _ := h.repo.FindByUsername(input.Username); existing != nil && existing.ID != user.ID {
			utils.RespondWithError(c, http.StatusConflict, "Username already taken")
			return
		}
		user.Username = input.Username
	}
	if input.Bio != "" {
		user.Bio = input.Bio
	}
	if input.AvatarURL != "" {
		// Delete old avatar if exists and different from new one
		if user.AvatarURL != "" && user.AvatarURL != input.AvatarURL {
			storage.DeleteOldAvatar(user.AvatarURL)
		}
		user.AvatarURL = input.AvatarURL
	}
	if input.Language != "" {
		user.Language = input.Language
	}
	if input.Theme != "" {
		user.Theme = models.Theme(input.Theme)
	}

	if err := h.repo.Update(user); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	response := dto.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Username:      user.Username,
		Role:          string(user.Role),
		AvatarURL:     user.AvatarURL,
		Bio:           user.Bio,
		Level:         user.Level,
		TotalXP:       user.TotalXP,
		CurrentXP:     user.CurrentXP,
		TargetXP:      user.TargetXP,
		StreakDays:    user.StreakDays,
		Language:      user.Language,
		Theme:         string(user.Theme),
		IsPro:         user.IsPro,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data":    response,
	})
}

// GetUserSettings godoc
// @Summary Get my settings
// @Description Get notification and other settings for user
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.UserSettingsResponse "User settings"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me/settings [get]
func (h *UserHandler) GetUserSettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	settings, err := h.repo.GetSettings(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get settings")
		return
	}

	response := dto.UserSettingsResponse{
		NotificationEmail:     settings.NotificationEmail,
		NotificationPush:      settings.NotificationPush,
		NotificationMarketing: settings.NotificationMarketing,
		NotificationUpdates:   settings.NotificationUpdates,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateUserSettings godoc
// @Summary Update my settings
// @Description Update notification and other settings
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param input body dto.UserSettingsUpdateRequest true "Settings update data"
// @Success 200 {object} dto.UserSettingsResponse "Updated settings"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me/settings [put]
func (h *UserHandler) UpdateUserSettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input dto.UserSettingsUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	settings, err := h.repo.GetSettings(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get settings")
		return
	}

	// Update only provided fields
	if input.NotificationEmail != nil {
		settings.NotificationEmail = *input.NotificationEmail
	}
	if input.NotificationPush != nil {
		settings.NotificationPush = *input.NotificationPush
	}
	if input.NotificationMarketing != nil {
		settings.NotificationMarketing = *input.NotificationMarketing
	}
	if input.NotificationUpdates != nil {
		settings.NotificationUpdates = *input.NotificationUpdates
	}

	if err := h.repo.UpdateSettings(settings); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
		"data": dto.UserSettingsResponse{
			NotificationEmail:     settings.NotificationEmail,
			NotificationPush:      settings.NotificationPush,
			NotificationMarketing: settings.NotificationMarketing,
			NotificationUpdates:   settings.NotificationUpdates,
		},
	})
}

// UploadAvatar godoc
// @Summary Upload avatar
// @Description Upload a new avatar image for the authenticated user
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image (JPEG, PNG, GIF, WebP, max 5MB)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Avatar URL"
// @Failure 400 {object} map[string]string "Invalid file or file too large"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user
	user, err := h.repo.FindByID(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	// Get file from form
	file, err := c.FormFile("avatar")
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "No file uploaded")
		return
	}

	// Delete old avatar if exists
	storage.DeleteOldAvatar(user.AvatarURL)

	// Upload new avatar
	avatarURL, err := storage.UploadAvatar(file, userID.(string))
	if err != nil {
		switch err {
		case storage.ErrFileTooLarge:
			utils.RespondWithError(c, http.StatusBadRequest, "File too large (max 5MB)")
		case storage.ErrInvalidFileType:
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid file type. Allowed: JPEG, PNG, GIF, WebP")
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to upload avatar")
		}
		return
	}

	// Update user avatar URL in database
	user.AvatarURL = avatarURL
	if err := h.repo.Update(user); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Avatar uploaded successfully",
		"avatar_url": avatarURL,
	})
}

// DeleteAvatar godoc
// @Summary Delete avatar
// @Description Remove the avatar image for the authenticated user
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me/avatar [delete]
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user
	user, err := h.repo.FindByID(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	// Delete avatar file
	storage.DeleteOldAvatar(user.AvatarURL)

	// Clear avatar URL in database
	user.AvatarURL = ""
	if err := h.repo.Update(user); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Avatar deleted successfully",
	})
}
