package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	service *services.ProjectService
}

// NewProjectHandler creates a new ProjectHandler
func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{
		service: services.NewProjectService(db),
	}
}

// ListProjects godoc
// @Summary List user's projects
// @Description Get list of current user's projects with optional filters
// @Tags Projects
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search query"
// @Param difficulty query string false "Filter by difficulty (Beginner, Intermediate, Advanced)"
// @Param is_favorite query bool false "Filter favorites only"
// @Param is_public query bool false "Filter public/private"
// @Param sort query string false "Sort field (created_at, updated_at, name)" default(created_at)
// @Param order query string false "Sort order (asc, desc)" default(desc)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of projects with pagination"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.ProjectListRequest
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

	projects, pagination, err := h.service.ListProjects(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"projects":   projects,
			"pagination": pagination,
		},
	})
}

// GetProject godoc
// @Summary Get project details
// @Description Get detailed information about a specific project including components and collaborators
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.ProjectDetailResponse "Project details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	project, err := h.service.GetProject(projectID, userID.(string))
	if err != nil {
		if err.Error() == "project not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "access denied" {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    project,
	})
}

// CreateProject godoc
// @Summary Create new project
// @Description Create a new project for the current user
// @Tags Projects
// @Accept json
// @Produce json
// @Param project body dto.ProjectCreateRequest true "Project data"
// @Security Bearer
// @Success 201 {object} dto.ProjectResponse "Created project"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	project, err := h.service.CreateProject(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Project created successfully",
		"data":    project,
	})
}

// UpdateProject godoc
// @Summary Update project
// @Description Update an existing project (owner only)
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Param project body dto.ProjectUpdateRequest true "Project update data"
// @Security Bearer
// @Success 200 {object} dto.ProjectResponse "Updated project"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var req dto.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	project, err := h.service.UpdateProject(projectID, userID.(string), req)
	if err != nil {
		if err.Error() == "access denied" {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Project updated successfully",
		"data":    project,
	})
}

// DeleteProject godoc
// @Summary Delete project
// @Description Delete a project and all its components (owner only)
// @Tags Projects
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Project deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	if err := h.service.DeleteProject(projectID, userID.(string)); err != nil {
		if err.Error() == "access denied" {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Project deleted successfully",
	})
}

// DuplicateProject godoc
// @Summary Duplicate project
// @Description Create a copy of an existing project (including components)
// @Tags Projects
// @Param id path string true "Project ID to duplicate (UUID)"
// @Security Bearer
// @Success 201 {object} dto.ProjectResponse "Duplicated project"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/duplicate [post]
func (h *ProjectHandler) DuplicateProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	project, err := h.service.DuplicateProject(projectID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Project duplicated successfully",
		"data":    project,
	})
}

// ToggleFavorite godoc
// @Summary Toggle project favorite
// @Description Toggle the favorite status of a project
// @Tags Projects
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Favorite status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/favorite [put]
func (h *ProjectHandler) ToggleFavorite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	isFavorite, err := h.service.ToggleFavorite(projectID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"is_favorite": isFavorite,
	})
}

// GetTemplates godoc
// @Summary Get project templates
// @Description Get list of public project templates that can be used as starting points
// @Tags Projects
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of templates"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/templates [get]
func (h *ProjectHandler) GetTemplates(c *gin.Context) {
	page := 1
	limit := 20

	projects, pagination, err := h.service.GetTemplates(page, limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"templates":  projects,
			"pagination": pagination,
		},
	})
}

// GetCollaborators godoc
// @Summary Get project collaborators
// @Description Get list of collaborators for a specific project
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of collaborators"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/collaborators [get]
func (h *ProjectHandler) GetCollaborators(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	collaborators, err := h.service.GetCollaborators(projectID, userID.(string))
	if err != nil {
		if err.Error() == "project not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "access denied" {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    collaborators,
	})
}

// AddCollaborator godoc
// @Summary Add collaborator to project
// @Description Invite a user to collaborate on a project by email
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Param collaborator body dto.AddCollaboratorRequest true "Collaborator data"
// @Security Bearer
// @Success 201 {object} dto.CollaboratorResponse "Added collaborator"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied (owner only)"
// @Failure 404 {object} map[string]string "Project or user not found"
// @Failure 409 {object} map[string]string "User already a collaborator"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/collaborators [post]
func (h *ProjectHandler) AddCollaborator(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var req dto.AddCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	collaborator, err := h.service.AddCollaborator(projectID, userID.(string), req)
	if err != nil {
		switch err.Error() {
		case "project not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "user not found":
			utils.RespondWithError(c, http.StatusNotFound, "User with this email not found")
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, "Only project owner can add collaborators")
		case "already a collaborator":
			utils.RespondWithError(c, http.StatusConflict, "User is already a collaborator")
		case "cannot add yourself":
			utils.RespondWithError(c, http.StatusBadRequest, "Cannot add yourself as a collaborator")
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Collaborator added successfully",
		"data":    collaborator,
	})
}

// RemoveCollaborator godoc
// @Summary Remove collaborator from project
// @Description Remove a collaborator from a project
// @Tags Projects
// @Param id path string true "Project ID (UUID)"
// @Param userId path string true "User ID to remove (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Project or collaborator not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id}/collaborators/{userId} [delete]
func (h *ProjectHandler) RemoveCollaborator(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := c.Param("id")
	collaboratorUserID := c.Param("userId")

	err := h.service.RemoveCollaborator(projectID, userID.(string), collaboratorUserID)
	if err != nil {
		switch err.Error() {
		case "project not found":
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		case "collaborator not found":
			utils.RespondWithError(c, http.StatusNotFound, "Collaborator not found")
		case "access denied":
			utils.RespondWithError(c, http.StatusForbidden, "Only project owner can remove collaborators")
		default:
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Collaborator removed successfully",
	})
}
