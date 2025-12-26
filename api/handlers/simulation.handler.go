package handlers

import (
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/dto"
	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SimulationHandler handles simulation-related HTTP requests
type SimulationHandler struct {
	service *services.SimulationService
}

// NewSimulationHandler creates a new SimulationHandler
func NewSimulationHandler(db *gorm.DB) *SimulationHandler {
	return &SimulationHandler{
		service: services.NewSimulationService(db),
	}
}

// ListSimulations godoc
// @Summary List user's simulations
// @Description Get list of current user's simulations
// @Tags Simulations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by type"
// @Param project_id query string false "Filter by project ID"
// @Param search query string false "Search by name"
// @Param sort query string false "Sort by field" default(created_at)
// @Param order query string false "Sort order" default(desc)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of simulations"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /simulations [get]
func (h *SimulationHandler) ListSimulations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.SimulationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = dto.SimulationListRequest{Page: 1, Limit: 20}
	}

	simulations, pagination, err := h.service.ListSimulations(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    simulations,
		"meta":    pagination,
	})
}

// GetStats godoc
// @Summary Get simulation statistics
// @Description Get aggregated simulation statistics for current user
// @Tags Simulations
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.SimulationStatsResponse "Simulation statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /simulations/stats [get]
func (h *SimulationHandler) GetStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.service.GetStats(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSimulation godoc
// @Summary Get simulation detail
// @Description Get detailed information about a simulation
// @Tags Simulations
// @Produce json
// @Param id path string true "Simulation ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.SimulationDetailResponse "Simulation details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Simulation not found"
// @Router /simulations/{id} [get]
func (h *SimulationHandler) GetSimulation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	simulationID := c.Param("id")
	simulation, err := h.service.GetSimulation(simulationID, userID.(string))
	if err != nil {
		if err.Error() == "simulation not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		} else {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    simulation,
	})
}

// CreateSimulation godoc
// @Summary Create new simulation
// @Description Create a new simulation
// @Tags Simulations
// @Accept json
// @Produce json
// @Param simulation body dto.CreateSimulationRequest true "Simulation data"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "Simulation created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /simulations [post]
func (h *SimulationHandler) CreateSimulation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CreateSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	simulation, err := h.service.CreateSimulation(userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    simulation,
		"message": "Simulation created successfully",
	})
}

// UpdateSimulation godoc
// @Summary Update simulation
// @Description Update an existing simulation
// @Tags Simulations
// @Accept json
// @Produce json
// @Param id path string true "Simulation ID (UUID)"
// @Param simulation body dto.UpdateSimulationRequest true "Simulation data"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Simulation updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Router /simulations/{id} [put]
func (h *SimulationHandler) UpdateSimulation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	simulationID := c.Param("id")
	var req dto.UpdateSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	simulation, err := h.service.UpdateSimulation(simulationID, userID.(string), req)
	if err != nil {
		if err.Error() == "simulation not found" {
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
		"data":    simulation,
		"message": "Simulation updated successfully",
	})
}

// DeleteSimulation godoc
// @Summary Delete simulation
// @Description Delete a simulation
// @Tags Simulations
// @Param id path string true "Simulation ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]string "Simulation deleted"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Simulation not found"
// @Router /simulations/{id} [delete]
func (h *SimulationHandler) DeleteSimulation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	simulationID := c.Param("id")
	if err := h.service.DeleteSimulation(simulationID, userID.(string)); err != nil {
		if err.Error() == "access denied" {
			utils.RespondWithError(c, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Simulation deleted successfully",
	})
}

// RunSimulation godoc
// @Summary Run simulation
// @Description Start a new simulation run
// @Tags Simulations
// @Accept json
// @Produce json
// @Param id path string true "Simulation ID (UUID)"
// @Param body body dto.RunSimulationRequestDTO false "Run parameters"
// @Security Bearer
// @Success 200 {object} dto.RunSimulationResponseDTO "Simulation started"
// @Failure 400 {object} map[string]string "Already running"
// @Failure 404 {object} map[string]string "Simulation not found"
// @Router /simulations/{id}/run [post]
func (h *SimulationHandler) RunSimulation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	simulationID := c.Param("id")
	var req dto.RunSimulationRequestDTO
	c.ShouldBindJSON(&req) // Optional

	result, err := h.service.RunSimulation(simulationID, userID.(string), req)
	if err != nil {
		if err.Error() == "simulation not found" {
			utils.RespondWithError(c, http.StatusNotFound, err.Error())
		} else if err.Error() == "simulation already running" {
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Simulation started",
	})
}

// StopSimulation godoc
// @Summary Stop simulation
// @Description Stop a running simulation
// @Tags Simulations
// @Produce json
// @Param id path string true "Simulation ID (UUID)"
// @Security Bearer
// @Success 200 {object} dto.StopSimulationResponse "Simulation stopped"
// @Failure 404 {object} map[string]string "Simulation not found"
// @Router /simulations/{id}/stop [post]
func (h *SimulationHandler) StopSimulation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	simulationID := c.Param("id")
	result, err := h.service.StopSimulation(simulationID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Simulation stopped",
	})
}

// GetRuns godoc
// @Summary Get simulation run history
// @Description Get run history for a simulation
// @Tags Simulations
// @Produce json
// @Param id path string true "Simulation ID (UUID)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Run history"
// @Failure 404 {object} map[string]string "Simulation not found"
// @Router /simulations/{id}/runs [get]
func (h *SimulationHandler) GetRuns(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	simulationID := c.Param("id")
	runs, err := h.service.GetRuns(simulationID, userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    runs,
	})
}

// SaveResult godoc
// @Summary Save simulation result
// @Description Save simulation result from frontend
// @Tags Simulations
// @Accept json
// @Produce json
// @Param id path string true "Simulation ID (UUID)"
// @Param body body dto.SaveSimulationResultRequest true "Result data"
// @Security Bearer
// @Success 200 {object} dto.SaveSimulationResultResponse "Result saved"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Simulation not found"
// @Router /simulations/{id}/result [post]
func (h *SimulationHandler) SaveResult(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	simulationID := c.Param("id")
	var req dto.SaveSimulationResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := h.service.SaveResult(simulationID, userID.(string), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Simulation result saved",
	})
}
