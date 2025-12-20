package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth-related HTTP requests
type OAuthHandler struct {
	oauthService *services.OAuthService
}

// NewOAuthHandler creates a new OAuthHandler instance
func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{
		oauthService: services.NewOAuthService(),
	}
}

// RegisterWithPassword godoc
// @Summary Register a new user with email and password
// @Description Register a new user with email, password, and optional username
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.RegisterRequest true "Register Request"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *OAuthHandler) RegisterWithPassword(c *gin.Context) {
	var input dto.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	user, err := services.RegisterUserWithPassword(&input)
	if err != nil {
		if strings.Contains(err.Error(), "already registered") {
			utils.RespondWithError(c, http.StatusConflict, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to register user: "+err.Error())
		return
	}

	// Auto-login after registration
	loginInput := dto.LoginRequest{
		Email:    input.Email,
		Password: input.Password,
	}

	authResponse, err := services.LoginWithPassword(&loginInput)
	if err != nil {
		// User created but auto-login failed, return user info without token
		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully. Please login.",
			"user": dto.UserAuthInfo{
				ID:         user.ID,
				Email:      user.Email,
				Username:   user.Username,
				Provider:   user.Provider,
				IsVerified: user.EmailVerified,
			},
		})
		return
	}

	c.JSON(http.StatusCreated, authResponse)
}

// LoginWithPassword godoc
// @Summary Login with email and password
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.LoginRequest true "Login Request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *OAuthHandler) LoginWithPassword(c *gin.Context) {
	var input dto.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	authResponse, err := services.LoginWithPassword(&input)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// GetOAuthURL godoc
// @Summary Get OAuth authorization URL
// @Description Get the authorization URL for the specified OAuth provider
// @Tags Auth
// @Produce json
// @Param provider path string true "OAuth Provider (google, github, apple)"
// @Success 200 {object} dto.OAuthURLResponse
// @Failure 400 {object} map[string]string
// @Failure 501 {object} map[string]string
// @Router /auth/{provider} [get]
func (h *OAuthHandler) GetOAuthURL(c *gin.Context) {
	provider := c.Param("provider")

	response, err := h.oauthService.GetOAuthURL(provider)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			utils.RespondWithError(c, http.StatusNotImplemented, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleOAuthCallback godoc
// @Summary Handle OAuth callback
// @Description Process OAuth callback and authenticate/register user
// @Tags Auth
// @Produce json
// @Param provider path string true "OAuth Provider (google, github, apple)"
// @Param code query string true "Authorization code from OAuth provider"
// @Param state query string false "State parameter for CSRF protection"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/{provider}/callback [get]
func (h *OAuthHandler) HandleOAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")

	// For Apple, also check POST form data (Apple uses form_post response mode)
	if code == "" {
		code = c.PostForm("code")
	}

	if code == "" {
		// Check for error from OAuth provider
		errorMsg := c.Query("error")
		errorDesc := c.Query("error_description")
		if errorMsg != "" {
			utils.RespondWithError(c, http.StatusBadRequest, "OAuth error: "+errorMsg+". "+errorDesc)
			return
		}
		utils.RespondWithError(c, http.StatusBadRequest, "Authorization code is required")
		return
	}

	var authResponse *dto.AuthResponse
	var err error

	// Apple requires special handling because it sends user data in the callback
	if provider == "apple" {
		idToken := c.PostForm("id_token")
		userData := c.PostForm("user") // Apple sends user data as JSON string
		authResponse, err = h.oauthService.HandleAppleCallback(code, idToken, userData)
	} else {
		authResponse, err = h.oauthService.HandleOAuthCallback(provider, code)
	}

	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "OAuth authentication failed: "+err.Error())
		return
	}

	// Check if frontend URL is configured for redirect
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" {
		// Redirect to frontend with token
		redirectURL := frontendURL + "/auth/callback?token=" + authResponse.Token + "&provider=" + provider
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	// Return JSON response if no frontend URL configured
	c.JSON(http.StatusOK, authResponse)
}

// HandleOAuthCallbackPost godoc
// @Summary Handle OAuth callback via POST
// @Description Process OAuth callback via POST request (for mobile/desktop apps and Apple Sign In)
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth Provider (google, github, apple)"
// @Param input body dto.OAuthCallbackRequest true "OAuth Callback Request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/{provider}/callback [post]
func (h *OAuthHandler) HandleOAuthCallbackPost(c *gin.Context) {
	provider := c.Param("provider")

	// Check content type to determine how to parse the request
	contentType := c.GetHeader("Content-Type")

	var code, idToken, userData string

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Apple Sign In sends form data
		code = c.PostForm("code")
		idToken = c.PostForm("id_token")
		userData = c.PostForm("user")
	} else {
		// JSON request (mobile apps)
		var input dto.OAuthCallbackRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
			return
		}
		code = input.Code
	}

	if code == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Authorization code is required")
		return
	}

	var authResponse *dto.AuthResponse
	var err error

	if provider == "apple" {
		authResponse, err = h.oauthService.HandleAppleCallback(code, idToken, userData)
	} else {
		authResponse, err = h.oauthService.HandleOAuthCallback(provider, code)
	}

	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "OAuth authentication failed: "+err.Error())
		return
	}

	// Check if frontend URL is configured for redirect (for form POST from Apple)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" && strings.Contains(contentType, "application/x-www-form-urlencoded") {
		redirectURL := frontendURL + "/auth/callback?token=" + authResponse.Token + "&provider=" + provider
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// GetSupportedProviders godoc
// @Summary Get list of supported OAuth providers
// @Description Returns list of OAuth providers that are configured and available
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string][]string
// @Router /auth/providers [get]
func (h *OAuthHandler) GetSupportedProviders(c *gin.Context) {
	providers := []map[string]interface{}{}

	// Check each provider
	if os.Getenv("GOOGLE_CLIENT_ID") != "" {
		providers = append(providers, map[string]interface{}{
			"name":    "google",
			"label":   "Google",
			"enabled": true,
		})
	}

	if os.Getenv("GITHUB_CLIENT_ID") != "" {
		providers = append(providers, map[string]interface{}{
			"name":    "github",
			"label":   "GitHub",
			"enabled": true,
		})
	}

	if os.Getenv("APPLE_CLIENT_ID") != "" {
		providers = append(providers, map[string]interface{}{
			"name":    "apple",
			"label":   "Apple",
			"enabled": true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}
