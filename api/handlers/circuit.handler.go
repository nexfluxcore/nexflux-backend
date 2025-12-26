package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CircuitHandler handles circuit-related HTTP requests
type CircuitHandler struct {
	service *services.CircuitService
}

// NewCircuitHandler creates a new CircuitHandler
func NewCircuitHandler(db *gorm.DB) *CircuitHandler {
	return &CircuitHandler{
		service: services.NewCircuitService(db),
	}
}

// ListCircuits godoc
// @Summary List user's circuits
// @Description Get list of current user's circuits
// @Tags Circuits
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param project_id query string false "Filter by project ID"
// @Param search query string false "Search by name"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of circuits"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /circuits [get]
func (h *CircuitHandler) ListCircuits(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CircuitListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = dto.CircuitListRequest{Page: 1, Limit: 20}
	}

	circuits, pagination, err := h.service.ListCircuits(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    circuits,
		"meta":    pagination,
	})
}

// GetCircuit godoc
// @Summary Get circuit detail
// @Description Get detailed information about a circuit
// @Tags Circuits
// @Produce json
// @Param id path string true "Circuit ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.CircuitDetailResponse "Circuit details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Circuit not found"
// @Router /circuits/{id} [get]
func (h *CircuitHandler) GetCircuit(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	circuitID := c.Param("id")
	circuit, err := h.service.GetCircuit(circuitID, userID.(string))
	if err != nil {
		if err.Error() == "circuit not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		} else {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    circuit,
	})
}

// CreateCircuit godoc
// @Summary Create new circuit
// @Description Create a new circuit
// @Tags Circuits
// @Accept json
// @Produce json
// @Param circuit body dto.CreateCircuitRequest true "Circuit data"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "Circuit created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /circuits [post]
func (h *CircuitHandler) CreateCircuit(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CreateCircuitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	circuit, xpEarned, err := h.service.CreateCircuit(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"data":      circuit,
		"message":   "Circuit created successfully",
		"xp_earned": xpEarned,
	})
}

// UpdateCircuit godoc
// @Summary Update circuit
// @Description Update an existing circuit
// @Tags Circuits
// @Accept json
// @Produce json
// @Param id path string true "Circuit ID (UUID)"
// @Param circuit body dto.UpdateCircuitRequest true "Circuit data"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Circuit updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Router /circuits/{id} [put]
func (h *CircuitHandler) UpdateCircuit(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	circuitID := c.Param("id")
	var req dto.UpdateCircuitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	circuit, err := h.service.UpdateCircuit(circuitID, userID.(string), req)
	if err != nil {
		if err.Error() == "circuit not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		} else if err.Error() == "access denied" {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    circuit,
		"message": "Circuit saved",
	})
}

// DeleteCircuit godoc
// @Summary Delete circuit
// @Description Delete a circuit
// @Tags Circuits
// @Param id path string true "Circuit ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]string "Circuit deleted"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Circuit not found"
// @Router /circuits/{id} [delete]
func (h *CircuitHandler) DeleteCircuit(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	circuitID := c.Param("id")
	if err := h.service.DeleteCircuit(circuitID, userID.(string)); err != nil {
		if err.Error() == "access denied" {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Circuit deleted successfully",
	})
}

// DuplicateCircuit godoc
// @Summary Duplicate circuit
// @Description Create a copy of an existing circuit
// @Tags Circuits
// @Param id path string true "Circuit ID to duplicate"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "Circuit duplicated"
// @Failure 404 {object} map[string]string "Circuit not found"
// @Router /circuits/{id}/duplicate [post]
func (h *CircuitHandler) DuplicateCircuit(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	circuitID := c.Param("id")
	circuit, err := h.service.DuplicateCircuit(circuitID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    circuit,
		"message": "Circuit duplicated",
	})
}

// UploadThumbnail godoc
// @Summary Upload circuit thumbnail
// @Description Upload a thumbnail image for a circuit
// @Tags Circuits
// @Accept json
// @Produce json
// @Param id path string true "Circuit ID"
// @Param thumbnail body dto.CircuitThumbnailRequest true "Base64 image data"
// @Security Bearer
// @Success 200 {object} dto.CircuitThumbnailResponse "Thumbnail uploaded"
// @Router /circuits/{id}/thumbnail [post]
func (h *CircuitHandler) UploadThumbnail(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	circuitID := c.Param("id")
	var req dto.CircuitThumbnailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	thumbnailURL, err := h.service.UploadThumbnail(circuitID, userID.(string), req.ImageData)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"thumbnail_url": thumbnailURL,
		},
	})
}

// ListTemplates godoc
// @Summary List circuit templates
// @Description Get list of available circuit templates
// @Tags Circuits
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param difficulty query string false "Filter by difficulty"
// @Param search query string false "Search by name"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of templates"
// @Router /circuits/templates [get]
func (h *CircuitHandler) ListTemplates(c *gin.Context) {
	var req dto.CircuitTemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = dto.CircuitTemplateListRequest{Page: 1, Limit: 20}
	}

	templates, pagination, err := h.service.ListTemplates(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
		"meta":    pagination,
	})
}

// UseTemplate godoc
// @Summary Use circuit template
// @Description Clone a template to user's circuits
// @Tags Circuits
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param body body dto.UseTemplateRequest true "Optional name and project ID"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "Template applied"
// @Router /circuits/templates/{id}/use [post]
func (h *CircuitHandler) UseTemplate(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	templateID := c.Param("id")
	var req dto.UseTemplateRequest
	c.ShouldBindJSON(&req) // Optional fields

	circuit, xpEarned, err := h.service.UseTemplate(templateID, userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"data":      circuit,
		"message":   "Template applied",
		"xp_earned": xpEarned,
	})
}

// ExportCircuit godoc
// @Summary Export circuit
// @Description Export circuit in different formats (json, spice)
// @Tags Circuits
// @Produce json
// @Param id path string true "Circuit ID"
// @Param format query string false "Export format: json, spice" default(json)
// @Security Bearer
// @Success 200 {object} dto.CircuitExportResponse "Export data"
// @Router /circuits/{id}/export [get]
func (h *CircuitHandler) ExportCircuit(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	circuitID := c.Param("id")
	format := c.DefaultQuery("format", "json")

	export, err := h.service.ExportCircuit(circuitID, userID.(string), format)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    export,
	})
}
