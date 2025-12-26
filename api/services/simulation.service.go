package services

import (
	"encoding/json"
	"errors"
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SimulationService handles simulation business logic
type SimulationService struct {
	repo     *repositories.SimulationRepository
	userRepo *repositories.UserRepository
	db       *gorm.DB
}

// NewSimulationService creates a new SimulationService
func NewSimulationService(db *gorm.DB) *SimulationService {
	return &SimulationService{
		repo:     repositories.NewSimulationRepository(db),
		userRepo: repositories.NewUserRepository(db),
		db:       db,
	}
}

// ListSimulations lists user's simulations
func (s *SimulationService) ListSimulations(userID string, req dto.SimulationListRequest) ([]dto.SimulationResponse, dto.PaginationResponse, error) {
	filter := repositories.SimulationFilter{
		UserID:    userID,
		ProjectID: req.ProjectID,
		Status:    req.Status,
		Type:      req.Type,
		Search:    req.Search,
	}

	simulations, total, err := s.repo.FindAll(filter, req.Page, req.Limit, req.Sort, req.Order)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.SimulationResponse, len(simulations))
	for i, sim := range simulations {
		responses[i] = s.toSimulationResponse(&sim)
	}

	return responses, dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      int(total),
		TotalPages: (int(total) + req.Limit - 1) / req.Limit,
	}, nil
}

// GetStats gets simulation statistics for a user
func (s *SimulationService) GetStats(userID string) (*dto.SimulationStatsResponse, error) {
	stats, err := s.repo.GetStats(userID)
	if err != nil {
		return nil, err
	}

	return &dto.SimulationStatsResponse{
		TotalSimulations:    stats.TotalSimulations,
		RunningNow:          stats.RunningNow,
		Completed:           stats.Completed,
		Paused:              stats.Paused,
		Error:               stats.Error,
		TotalRuntimeHours:   stats.TotalRuntimeHours,
		SuccessRate:         stats.SuccessRate,
		SimulationsThisWeek: stats.SimulationsThisWeek,
		ByType:              stats.ByType,
	}, nil
}

// GetSimulation gets simulation by ID
func (s *SimulationService) GetSimulation(simulationID, userID string) (*dto.SimulationDetailResponse, error) {
	simulation, err := s.repo.FindByID(simulationID)
	if err != nil {
		return nil, errors.New("simulation not found")
	}

	// Check access (only owner can view non-public simulations)
	if simulation.UserID != userID {
		return nil, errors.New("access denied")
	}

	return s.toSimulationDetailResponse(simulation), nil
}

// CreateSimulation creates a new simulation
func (s *SimulationService) CreateSimulation(userID string, req dto.CreateSimulationRequest) (*dto.SimulationResponse, error) {
	// Parse schema to count components and wires
	componentsCount, wiresCount := s.countSchemaElements(req.SchemaData)

	simType := models.SimulationType(req.Type)
	if simType == "" {
		simType = models.SimTypeBasicElectronics
	}

	simulation := &models.Simulation{
		UserID:             userID,
		ProjectID:          req.ProjectID,
		Name:               req.Name,
		Description:        req.Description,
		Type:               simType,
		SchemaData:         req.SchemaData,
		SimulationSettings: req.SimulationSettings,
		ComponentsCount:    componentsCount,
		WiresCount:         wiresCount,
		Status:             models.SimStatusDraft,
	}

	if err := s.repo.Create(simulation); err != nil {
		return nil, err
	}

	// Award XP for first simulation
	s.awardXP(userID, 10, simulation.ID, "simulation_create")

	return s.toSimulationResponsePtr(simulation), nil
}

// UpdateSimulation updates a simulation
func (s *SimulationService) UpdateSimulation(simulationID, userID string, req dto.UpdateSimulationRequest) (*dto.SimulationResponse, error) {
	simulation, err := s.repo.FindByID(simulationID)
	if err != nil {
		return nil, errors.New("simulation not found")
	}

	if simulation.UserID != userID {
		return nil, errors.New("access denied")
	}

	if req.Name != "" {
		simulation.Name = req.Name
	}
	if req.Description != "" {
		simulation.Description = req.Description
	}
	if req.Type != "" {
		simulation.Type = models.SimulationType(req.Type)
	}
	if req.SchemaData != nil {
		simulation.SchemaData = req.SchemaData
		simulation.ComponentsCount, simulation.WiresCount = s.countSchemaElements(req.SchemaData)
	}
	if req.SimulationSettings != nil {
		simulation.SimulationSettings = req.SimulationSettings
	}

	if err := s.repo.Update(simulation); err != nil {
		return nil, err
	}

	return s.toSimulationResponsePtr(simulation), nil
}

// DeleteSimulation deletes a simulation
func (s *SimulationService) DeleteSimulation(simulationID, userID string) error {
	isOwner, err := s.repo.IsOwner(simulationID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("access denied")
	}

	return s.repo.Delete(simulationID)
}

// RunSimulation starts a simulation
func (s *SimulationService) RunSimulation(simulationID, userID string, req dto.RunSimulationRequestDTO) (*dto.RunSimulationResponseDTO, error) {
	simulation, err := s.repo.FindByID(simulationID)
	if err != nil {
		return nil, errors.New("simulation not found")
	}

	if simulation.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Check if already running
	if simulation.Status == models.SimStatusRunning {
		return nil, errors.New("simulation already running")
	}

	// Update status to running
	simulation.Status = models.SimStatusRunning
	if err := s.repo.Update(simulation); err != nil {
		return nil, err
	}

	// Create run record
	run := &models.SimulationRun{
		SimulationID: simulationID,
		UserID:       userID,
		Status:       "running",
	}

	if err := s.repo.CreateRun(run); err != nil {
		return nil, err
	}

	return &dto.RunSimulationResponseDTO{
		RunID:     run.ID,
		Status:    "running",
		StartedAt: run.StartedAt,
	}, nil
}

// StopSimulation stops a running simulation
func (s *SimulationService) StopSimulation(simulationID, userID string) (*dto.StopSimulationResponse, error) {
	simulation, err := s.repo.FindByID(simulationID)
	if err != nil {
		return nil, errors.New("simulation not found")
	}

	if simulation.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Update status
	simulation.Status = models.SimStatusPaused
	if err := s.repo.Update(simulation); err != nil {
		return nil, err
	}

	// Get latest run
	run, err := s.repo.GetLatestRun(simulationID)
	durationMs := 0
	if err == nil && run != nil {
		now := time.Now()
		run.Status = "stopped"
		run.CompletedAt = &now
		run.DurationMs = int(time.Since(run.StartedAt).Milliseconds())
		durationMs = run.DurationMs
		s.repo.UpdateRun(run)
	}

	return &dto.StopSimulationResponse{
		Status:     "paused",
		DurationMs: durationMs,
	}, nil
}

// GetRuns gets simulation run history
func (s *SimulationService) GetRuns(simulationID, userID string) ([]dto.SimulationRunResponse, error) {
	simulation, err := s.repo.FindByID(simulationID)
	if err != nil {
		return nil, errors.New("simulation not found")
	}

	if simulation.UserID != userID {
		return nil, errors.New("access denied")
	}

	runs, err := s.repo.GetRuns(simulationID, 50)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.SimulationRunResponse, len(runs))
	for i, run := range runs {
		responses[i] = dto.SimulationRunResponse{
			ID:          run.ID,
			Status:      run.Status,
			DurationMs:  run.DurationMs,
			ResultData:  run.ResultData,
			Errors:      run.Errors,
			Warnings:    run.Warnings,
			StartedAt:   run.StartedAt,
			CompletedAt: run.CompletedAt,
		}
	}

	return responses, nil
}

// SaveResult saves simulation result from frontend
func (s *SimulationService) SaveResult(simulationID, userID string, req dto.SaveSimulationResultRequest) (*dto.SaveSimulationResultResponse, error) {
	simulation, err := s.repo.FindByID(simulationID)
	if err != nil {
		return nil, errors.New("simulation not found")
	}

	if simulation.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Determine status from result
	var resultData struct {
		Success bool `json:"success"`
	}
	json.Unmarshal(req.ResultData, &resultData)

	status := models.SimStatusCompleted
	if !resultData.Success {
		status = models.SimStatusError
	}

	// Update simulation
	simulation.Status = status
	simulation.LastResult = req.ResultData
	now := time.Now()
	simulation.LastRunAt = &now

	if err := s.repo.Update(simulation); err != nil {
		return nil, err
	}

	// Update run counts
	s.repo.IncrementRunCount(simulationID, req.DurationMs)

	// Create run record
	run := &models.SimulationRun{
		SimulationID: simulationID,
		UserID:       userID,
		Status:       "completed",
		DurationMs:   req.DurationMs,
		ResultData:   req.ResultData,
		Errors:       req.Errors,
		Warnings:     req.Warnings,
		CompletedAt:  &now,
	}
	s.repo.CreateRun(run)

	// Award XP
	xpEarned := 5 // Simulation complete XP
	if resultData.Success {
		xpEarned = 10
	}
	s.awardXP(userID, xpEarned, simulationID, "simulation_complete")

	return &dto.SaveSimulationResultResponse{
		RunID:    run.ID,
		XPEarned: xpEarned,
	}, nil
}

// ============================================
// Helper Functions
// ============================================

func (s *SimulationService) countSchemaElements(schemaData datatypes.JSON) (components, wires int) {
	var schema struct {
		Components []interface{} `json:"components"`
		Wires      []interface{} `json:"wires"`
	}
	json.Unmarshal(schemaData, &schema)
	return len(schema.Components), len(schema.Wires)
}

func (s *SimulationService) toSimulationResponse(sim *models.Simulation) dto.SimulationResponse {
	resp := dto.SimulationResponse{
		ID:              sim.ID,
		Name:            sim.Name,
		Description:     sim.Description,
		Type:            string(sim.Type),
		Status:          string(sim.Status),
		ThumbnailURL:    sim.ThumbnailURL,
		ComponentsCount: sim.ComponentsCount,
		WiresCount:      sim.WiresCount,
		RunCount:        sim.RunCount,
		TotalRuntimeMs:  sim.TotalRuntimeMs,
		LastRunAt:       sim.LastRunAt,
		ProjectID:       sim.ProjectID,
		CreatedAt:       sim.CreatedAt,
		UpdatedAt:       sim.UpdatedAt,
	}

	if sim.Project != nil {
		resp.ProjectName = sim.Project.Name
	}

	return resp
}

func (s *SimulationService) toSimulationResponsePtr(sim *models.Simulation) *dto.SimulationResponse {
	resp := s.toSimulationResponse(sim)
	return &resp
}

func (s *SimulationService) toSimulationDetailResponse(sim *models.Simulation) *dto.SimulationDetailResponse {
	return &dto.SimulationDetailResponse{
		SimulationResponse: s.toSimulationResponse(sim),
		SchemaData:         sim.SchemaData,
		SimulationSettings: sim.SimulationSettings,
		LastResult:         sim.LastResult,
		ErrorMessage:       sim.ErrorMessage,
	}
}

func (s *SimulationService) awardXP(userID string, xpAmount int, sourceID, description string) {
	tx := &models.UserXPTransaction{
		UserID:      userID,
		XPAmount:    xpAmount,
		XPType:      models.XPSourceProject,
		SourceID:    &sourceID,
		SourceType:  "simulation",
		Description: description,
	}
	s.db.Create(tx)

	s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"current_xp": gorm.Expr("current_xp + ?", xpAmount),
			"total_xp":   gorm.Expr("total_xp + ?", xpAmount),
		})
}
