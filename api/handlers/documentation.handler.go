package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DocHandler handles documentation-related HTTP requests
type DocHandler struct {
	service *services.DocService
}

// NewDocHandler creates a new DocHandler
func NewDocHandler(db *gorm.DB) *DocHandler {
	return &DocHandler{
		service: services.NewDocService(db),
	}
}

// ==================================================
// CATEGORY ENDPOINTS
// ==================================================

// GetCategories godoc
// @Summary Get documentation categories
// @Description Get list of all documentation categories with article counts
// @Tags Documentation
// @Produce json
// @Success 200 {object} map[string]interface{} "List of categories"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/categories [get]
func (h *DocHandler) GetCategories(c *gin.Context) {
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

// GetCategoryBySlug godoc
// @Summary Get category details
// @Description Get category details with its articles
// @Tags Documentation
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} dto.DocCategoryDetailResponse "Category details"
// @Failure 404 {object} map[string]string "Category not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/categories/{slug} [get]
func (h *DocHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.service.GetCategoryBySlug(slug)
	if err != nil {
		if err.Error() == "record not found" {
			utils.RespondWithError(c, http.StatusNotFound, "Category not found")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    category,
	})
}

// ==================================================
// ARTICLE ENDPOINTS
// ==================================================

// ListArticles godoc
// @Summary List articles
// @Description Get list of documentation articles with filters
// @Tags Documentation
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Category slug filter"
// @Param difficulty query string false "Difficulty filter (beginner, intermediate, advanced)"
// @Param search query string false "Search query"
// @Param sort query string false "Sort by (newest, popular, title)" default(newest)
// @Success 200 {object} dto.DocArticleListResponse "List of articles"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/articles [get]
func (h *DocHandler) ListArticles(c *gin.Context) {
	var req dto.DocArticleListRequest
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

	response, err := h.service.ListArticles(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetArticleBySlug godoc
// @Summary Get article details
// @Description Get full article content with related articles and table of contents
// @Tags Documentation
// @Produce json
// @Param slug path string true "Article slug"
// @Success 200 {object} dto.DocArticleDetailResponse "Article details"
// @Failure 404 {object} map[string]string "Article not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/articles/{slug} [get]
func (h *DocHandler) GetArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")

	article, err := h.service.GetArticleBySlug(slug)
	if err != nil {
		if err.Error() == "record not found" {
			utils.RespondWithError(c, http.StatusNotFound, "Article not found")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    article,
	})
}

// GetPopularArticles godoc
// @Summary Get popular articles
// @Description Get most viewed articles
// @Tags Documentation
// @Produce json
// @Param limit query int false "Number of articles" default(10)
// @Success 200 {object} map[string]interface{} "List of popular articles"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/articles/popular [get]
func (h *DocHandler) GetPopularArticles(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	articles, err := h.service.GetPopularArticles(limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    articles,
	})
}

// GetFeaturedArticles godoc
// @Summary Get featured articles
// @Description Get featured/highlighted articles
// @Tags Documentation
// @Produce json
// @Param limit query int false "Number of articles" default(10)
// @Success 200 {object} map[string]interface{} "List of featured articles"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/articles/featured [get]
func (h *DocHandler) GetFeaturedArticles(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	articles, err := h.service.GetFeaturedArticles(limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    articles,
	})
}

// IncrementArticleView godoc
// @Summary Record article view
// @Description Increment view count when user opens an article
// @Tags Documentation
// @Produce json
// @Param slug path string true "Article slug"
// @Success 200 {object} map[string]interface{} "Updated view count"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/articles/{slug}/view [post]
func (h *DocHandler) IncrementArticleView(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.service.IncrementArticleView(slug)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "View recorded",
		"data":    result,
	})
}

// ==================================================
// VIDEO ENDPOINTS
// ==================================================

// ListVideos godoc
// @Summary List video tutorials
// @Description Get list of video tutorials with filters
// @Tags Documentation
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Category slug filter"
// @Param difficulty query string false "Difficulty filter"
// @Success 200 {object} dto.DocVideoListResponse "List of videos"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/videos [get]
func (h *DocHandler) ListVideos(c *gin.Context) {
	var req dto.DocVideoListRequest
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

	response, err := h.service.ListVideos(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetVideoByID godoc
// @Summary Get video details
// @Description Get video tutorial details by ID
// @Tags Documentation
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 200 {object} dto.DocVideoResponse "Video details"
// @Failure 404 {object} map[string]string "Video not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/videos/{id} [get]
func (h *DocHandler) GetVideoByID(c *gin.Context) {
	id := c.Param("id")

	video, err := h.service.GetVideoByID(id)
	if err != nil {
		if err.Error() == "record not found" {
			utils.RespondWithError(c, http.StatusNotFound, "Video not found")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    video,
	})
}

// ==================================================
// SEARCH ENDPOINT
// ==================================================

// Search godoc
// @Summary Search documentation
// @Description Search articles and videos
// @Tags Documentation
// @Produce json
// @Param q query string true "Search query (min 2 characters)"
// @Param type query string false "Search type (all, articles, videos)" default(all)
// @Param limit query int false "Results limit per type" default(10)
// @Success 200 {object} dto.DocSearchResponse "Search results"
// @Failure 400 {object} map[string]string "Invalid query"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /docs/search [get]
func (h *DocHandler) Search(c *gin.Context) {
	var req dto.DocSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Search query is required (min 2 characters)")
		return
	}

	if req.Limit < 1 {
		req.Limit = 10
	}

	response, err := h.service.Search(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
