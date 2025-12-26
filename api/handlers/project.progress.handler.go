package handlers

import (
	"encoding/json"
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProjectProgressHandler handles project progress-related HTTP requests
type ProjectProgressHandler struct {
	service *services.ProjectProgressService
}

// NewProjectProgressHandler creates a new ProjectProgressHandler
func NewProjectProgressHandler(db *gorm.DB) *ProjectProgressHandler {
	return &ProjectProgressHandler{
		service: services.NewProjectProgressService(db),
	}
}

// GetProgress godoc
// @Summary Get project progress detail
// @Description Get detailed progress breakdown for a project including milestones and XP earned
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.GetProgressResponse "Progress details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/progress [get]
func (h *ProjectProgressHandler) GetProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	result, err := h.service.GetProgress(projectID, userID.(string))
	if err != nil {
		switch err.Error() {
		case "project not found":
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

// UpdateProgress godoc
// @Summary Update project progress
// @Description Update progress for a specific component (schema, code, simulation, verification)
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Param body body dto.ProjectProgressUpdateRequest true "Progress update data"
// @Security Bearer
// @Success 200 {object} dto.UpdateProgressResponse "Progress updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/progress [put]
func (h *ProjectProgressHandler) UpdateProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var req dto.ProjectProgressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.UpdateProgress(projectID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "project not found":
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

// SaveSchema godoc
// @Summary Save project schema/circuit data
// @Description Save circuit schema data including components and connections
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Param body body dto.SaveSchemaRequest true "Schema data"
// @Security Bearer
// @Success 200 {object} dto.SaveSchemaResponse "Schema saved"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/schema [put]
func (h *ProjectProgressHandler) SaveSchema(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var req dto.SaveSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.SaveSchema(projectID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "project not found":
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

// SaveCode godoc
// @Summary Save project code
// @Description Save code/firmware data for the project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Param body body dto.SaveCodeRequest true "Code data"
// @Security Bearer
// @Success 200 {object} dto.SaveCodeResponse "Code saved"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/code [put]
func (h *ProjectProgressHandler) SaveCode(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var req dto.SaveCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.SaveCode(projectID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "project not found":
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

// RunSimulation godoc
// @Summary Run project simulation
// @Description Execute simulation for the project circuit and code
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Param body body dto.RunSimulationRequest true "Simulation parameters"
// @Security Bearer
// @Success 200 {object} dto.RunSimulationResponse "Simulation results"
// @Failure 400 {object} map[string]string "Invalid input or missing schema/code"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/simulate [post]
func (h *ProjectProgressHandler) RunSimulation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var req dto.RunSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body with defaults
		req = dto.RunSimulationRequest{
			DurationMs:      5000,
			SpeedMultiplier: 1.0,
		}
	}

	result, err := h.service.RunSimulation(projectID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "project not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		case "schema data required", "code data required":
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

// CompleteProject godoc
// @Summary Complete project
// @Description Mark project as complete and award completion XP
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.CompleteProjectResponse "Project completed"
// @Failure 400 {object} map[string]string "Project not ready for completion"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/complete [post]
func (h *ProjectProgressHandler) CompleteProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	result, err := h.service.CompleteProject(projectID, userID.(string))
	if err != nil {
		switch err.Error() {
		case "project not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		case "project already completed", "project not ready for completion":
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

// GetSchemaData godoc
// @Summary Get project schema data
// @Description Get the circuit schema data for a project
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Schema data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/schema [get]
func (h *ProjectProgressHandler) GetSchemaData(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	schemaData, err := h.service.GetSchemaData(projectID, userID.(string))
	if err != nil {
		switch err.Error() {
		case "project not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Parse JSON to return as object
	var data map[string]interface{}
	if schemaData != nil {
		json.Unmarshal(schemaData, &data)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetCodeData godoc
// @Summary Get project code data
// @Description Get the code/firmware data for a project
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Code data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/code [get]
func (h *ProjectProgressHandler) GetCodeData(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	codeData, err := h.service.GetCodeData(projectID, userID.(string))
	if err != nil {
		switch err.Error() {
		case "project not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Parse JSON to return as object
	var data map[string]interface{}
	if codeData != nil {
		json.Unmarshal(codeData, &data)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
