package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ComponentHandler handles component-related HTTP requests
type ComponentHandler struct {
	service *services.ComponentService
}

// NewComponentHandler creates a new ComponentHandler
func NewComponentHandler(db *gorm.DB) *ComponentHandler {
	return &ComponentHandler{
		service: services.NewComponentService(db),
	}
}

// ListComponents godoc
// @Summary List components
// @Description Get list of electronic components with optional filters
// @Tags Components
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Category ID or slug"
// @Param search query string false "Search query"
// @Param in_stock query bool false "Filter in-stock only"
// @Param sort query string false "Sort field (name, price, rating)" default(name)
// @Param order query string false "Sort order (asc, desc)" default(asc)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of components with pagination"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components [get]
func (h *ComponentHandler) ListComponents(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	var req dto.ComponentListRequest
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

	components, pagination, err := h.service.ListComponents(req, userIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"components": components,
			"pagination": pagination,
		},
	})
}

// GetComponent godoc
// @Summary Get component details
// @Description Get detailed information about a specific electronic component
// @Tags Components
// @Produce json
// @Param id path string true "Component ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.ComponentResponse "Component details"
// @Failure 404 {object} map[string]string "Component not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/{id} [get]
func (h *ComponentHandler) GetComponent(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	componentID := c.Param("id")

	component, err := h.service.GetComponent(componentID, userIDStr)
	if err != nil {
		if err.Error() == "component not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    component,
	})
}

// SearchComponents godoc
// @Summary Search components
// @Description Search components by name, part number, or description
// @Tags Components
// @Produce json
// @Param q query string true "Search query (min 2 characters)"
// @Param category query string false "Category filter"
// @Param limit query int false "Max results" default(10)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Search results"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/search [get]
func (h *ComponentHandler) SearchComponents(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	var req dto.ComponentSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid parameters: "+err.Error())
		return
	}

	if req.Limit < 1 {
		req.Limit = 10
	}

	components, err := h.service.SearchComponents(req, userIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    components,
	})
}

// GetCategories godoc
// @Summary Get component categories
// @Description Get list of all component categories (no auth required)
// @Tags Components
// @Produce json
// @Success 200 {object} map[string]interface{} "List of categories"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/categories [get]
func (h *ComponentHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

// GetFavorites godoc
// @Summary Get favorite components
// @Description Get user's favorite/saved components
// @Tags Components
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of favorite components"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/favorites [get]
func (h *ComponentHandler) GetFavorites(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	page := 1
	limit := 20

	components, pagination, err := h.service.GetFavorites(userID.(string), page, limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"components": components,
			"pagination": pagination,
		},
	})
}

// ToggleFavorite godoc
// @Summary Toggle component favorite
// @Description Add or remove component from user's favorites
// @Tags Components
// @Param id path string true "Component ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Favorite status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Component not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/{id}/favorite [post]
func (h *ComponentHandler) ToggleFavorite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	componentID := c.Param("id")

	isFavorite, err := h.service.ToggleFavorite(componentID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"is_favorite": isFavorite,
	})
}

// CreateRequest godoc
// @Summary Request new component
// @Description Submit a request to add a new component to the library
// @Tags Components
// @Accept json
// @Produce json
// @Param request body dto.ComponentRequestCreateRequest true "Component request data"
// @Security Bearer
// @Success 201 {object} dto.ComponentRequestResponse "Created request"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/request [post]
func (h *ComponentHandler) CreateRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.ComponentRequestCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	request, err := h.service.CreateRequest(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Component request submitted successfully",
		"data":    request,
	})
}

// GetUserRequests godoc
// @Summary Get my component requests
// @Description Get list of component requests submitted by the current user
// @Tags Components
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of requests"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/requests [get]
func (h *ComponentHandler) GetUserRequests(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	requests, err := h.service.GetUserRequests(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
	})
}
