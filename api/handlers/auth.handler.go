package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	service *services.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		service: services.NewAuthService(),
	}
}

// Register godoc
// @Summary Register new user
// @Description Create a new user account with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.AuthResponse "Registration successful"
// @Failure 400 {object} map[string]string "Invalid input or email exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	response, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "email already registered" {
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registration successful",
		"data":    response,
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse "Login successful"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	response, err := h.service.Login(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data":    response,
	})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Email address"
// @Success 200 {object} map[string]interface{} "Reset email sent (always returns success for security)"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Always return success to prevent email enumeration
	_ = h.service.ForgotPassword(req)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If an account with that email exists, a password reset link has been sent.",
	})
}

// VerifyResetToken godoc
// @Summary Verify password reset token
// @Description Check if a password reset token is valid
// @Tags Auth
// @Produce json
// @Param token query string true "Reset token"
// @Success 200 {object} map[string]interface{} "Token is valid"
// @Failure 400 {object} map[string]string "Invalid or expired token"
// @Router /auth/verify-reset-token [get]
func (h *AuthHandler) VerifyResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Token is required")
		return
	}

	if err := h.service.VerifyResetToken(token); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token is valid",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset user password with valid token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Token and new password"
// @Success 200 {object} map[string]interface{} "Password reset successful"
// @Failure 400 {object} map[string]string "Invalid token or password"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if err := h.service.ResetPassword(req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password has been reset successfully",
	})
}

// ChangePassword godoc
// @Summary Change password
// @Description Change password for authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Current and new password"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Password changed"
// @Failure 400 {object} map[string]string "Invalid password"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if err := h.service.ChangePassword(userID.(string), req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.RefreshTokenResponse "New tokens"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	response, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate user's tokens
// @Tags Auth
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Logged out"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_ = h.service.Logout(userID.(string))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetMe godoc
// @Summary Get current user
// @Description Get profile of authenticated user
// @Tags Auth
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.UserAuthInfo "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.service.GetCurrentUser(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}
