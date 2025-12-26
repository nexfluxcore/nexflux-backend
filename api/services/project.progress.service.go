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

// ProjectProgressService handles project progress business logic
type ProjectProgressService struct {
	repo       *repositories.ProjectRepository
	userRepo   *repositories.UserRepository
	gamRepo    *repositories.GamificationRepository
	progressDB *gorm.DB
}

// NewProjectProgressService creates a new ProjectProgressService
func NewProjectProgressService(db *gorm.DB) *ProjectProgressService {
	return &ProjectProgressService{
		repo:       repositories.NewProjectRepository(db),
		userRepo:   repositories.NewUserRepository(db),
		gamRepo:    repositories.NewGamificationRepository(db),
		progressDB: db,
	}
}

// GetProgress gets detailed progress for a project
func (s *ProjectProgressService) GetProgress(projectID, userID string) (*dto.GetProgressResponse, error) {
	// Get project with access check
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if !s.hasAccess(project, userID) {
		return nil, errors.New("access denied")
	}

	// Get progress breakdown
	breakdown := s.calculateBreakdown(project)
	totalProgress := s.calculateTotalProgress(breakdown)

	// Get milestones
	milestones := s.getMilestones(projectID)

	// Calculate total XP earned
	totalXP := s.calculateTotalXPEarned(projectID, userID)

	// Determine next action
	nextAction := s.getNextAction(breakdown)

	return &dto.GetProgressResponse{
		Progress:   totalProgress,
		Breakdown:  breakdown,
		Milestones: milestones,
		NextAction: nextAction,
		TotalXP:    totalXP,
	}, nil
}

// UpdateProgress updates progress for a specific component
func (s *ProjectProgressService) UpdateProgress(projectID, userID string, req dto.ProjectProgressUpdateRequest) (*dto.UpdateProgressResponse, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if !s.hasAccess(project, userID) {
		return nil, errors.New("access denied")
	}

	oldProgress := project.Progress
	xpEarned := 0
	milestonesUnlocked := []string{}

	// Update based on component
	switch req.Component {
	case "schema":
		xp, milestones := s.updateSchemaProgress(project, userID, req.Data)
		xpEarned = xp
		milestonesUnlocked = milestones
	case "code":
		xp, milestones := s.updateCodeProgress(project, userID, req.Data)
		xpEarned = xp
		milestonesUnlocked = milestones
	case "simulation":
		xp, milestones := s.updateSimulationProgress(project, userID, req.Data, req.Action == "run")
		xpEarned = xp
		milestonesUnlocked = milestones
	case "verification":
		xp, milestones := s.updateVerificationProgress(project, userID)
		xpEarned = xp
		milestonesUnlocked = milestones
	}

	// Recalculate total progress
	breakdown := s.calculateBreakdown(project)
	newProgress := s.calculateTotalProgress(breakdown)
	project.Progress = newProgress

	// Update project
	s.progressDB.Save(project)

	// Award XP to user
	if xpEarned > 0 {
		s.awardXP(userID, xpEarned, projectID, "project_progress")
	}

	return &dto.UpdateProgressResponse{
		Progress:           newProgress,
		ProgressChange:     newProgress - oldProgress,
		XPEarned:           xpEarned,
		MilestonesUnlocked: milestonesUnlocked,
		Breakdown:          breakdown,
	}, nil
}

// SaveSchema saves circuit schema data
func (s *ProjectProgressService) SaveSchema(projectID, userID string, req dto.SaveSchemaRequest) (*dto.SaveSchemaResponse, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if !s.hasAccess(project, userID) {
		return nil, errors.New("access denied")
	}

	oldProgress := project.Progress
	isFirstSave := len(project.SchemaData) == 0 || string(project.SchemaData) == "{}" || string(project.SchemaData) == "null"

	// Build schema data
	schemaData := map[string]interface{}{
		"components":      req.Components,
		"connections":     req.Connections,
		"canvas_settings": req.CanvasSettings,
		"last_saved":      time.Now().Format(time.RFC3339),
	}
	schemaJSON, _ := json.Marshal(schemaData)
	project.SchemaData = datatypes.JSON(schemaJSON)

	// Calculate XP
	xpEarned := 0
	if isFirstSave {
		if !s.hasMilestone(projectID, models.MilestoneFirstSchemaSave) {
			xpEarned = models.XPFirstSchemaSave
			s.createMilestone(projectID, userID, models.MilestoneFirstSchemaSave, xpEarned)
		}
	}

	// Recalculate progress
	breakdown := s.calculateBreakdown(project)
	newProgress := s.calculateTotalProgress(breakdown)
	project.Progress = newProgress

	// Save project
	s.progressDB.Save(project)

	// Award XP
	if xpEarned > 0 {
		s.awardXP(userID, xpEarned, projectID, "first_schema_save")
	}

	return &dto.SaveSchemaResponse{
		Success:        true,
		Progress:       newProgress,
		ProgressChange: newProgress - oldProgress,
		XPEarned:       xpEarned,
		IsFirstSave:    isFirstSave,
	}, nil
}

// SaveCode saves code data
func (s *ProjectProgressService) SaveCode(projectID, userID string, req dto.SaveCodeRequest) (*dto.SaveCodeResponse, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if !s.hasAccess(project, userID) {
		return nil, errors.New("access denied")
	}

	oldProgress := project.Progress
	isFirstSave := len(project.CodeData) == 0 || string(project.CodeData) == "{}" || string(project.CodeData) == "null"

	// Build code data
	codeData := map[string]interface{}{
		"content":    req.Content,
		"language":   req.Language,
		"filename":   req.Filename,
		"last_saved": time.Now().Format(time.RFC3339),
	}
	codeJSON, _ := json.Marshal(codeData)
	project.CodeData = datatypes.JSON(codeJSON)

	// Calculate XP
	xpEarned := 0
	if isFirstSave && len(req.Content) >= 100 {
		if !s.hasMilestone(projectID, models.MilestoneFirstCodeSave) {
			xpEarned = models.XPFirstCodeSave
			s.createMilestone(projectID, userID, models.MilestoneFirstCodeSave, xpEarned)
		}
	}

	// Recalculate progress
	breakdown := s.calculateBreakdown(project)
	newProgress := s.calculateTotalProgress(breakdown)
	project.Progress = newProgress

	// Save project
	s.progressDB.Save(project)

	// Award XP
	if xpEarned > 0 {
		s.awardXP(userID, xpEarned, projectID, "first_code_save")
	}

	return &dto.SaveCodeResponse{
		Success:        true,
		Progress:       newProgress,
		ProgressChange: newProgress - oldProgress,
		XPEarned:       xpEarned,
		IsFirstSave:    isFirstSave,
		CharCount:      len(req.Content),
	}, nil
}

// RunSimulation runs simulation and returns results
func (s *ProjectProgressService) RunSimulation(projectID, userID string, req dto.RunSimulationRequest) (*dto.RunSimulationResponse, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if !s.hasAccess(project, userID) {
		return nil, errors.New("access denied")
	}

	// Check if schema and code exist
	if len(project.SchemaData) == 0 || string(project.SchemaData) == "{}" {
		return nil, errors.New("schema data required")
	}
	if len(project.CodeData) == 0 || string(project.CodeData) == "{}" {
		return nil, errors.New("code data required")
	}

	oldProgress := project.Progress

	// Simulate the circuit (for now, always success - in production, actual simulation logic)
	simulationSuccess := true // TODO: Implement actual simulation engine

	xpEarned := 0
	milestonesUnlocked := []string{}

	// First simulation milestone
	if !s.hasMilestone(projectID, models.MilestoneFirstSimulation) {
		xpEarned += models.XPFirstSimulationRun
		milestonesUnlocked = append(milestonesUnlocked, string(models.MilestoneFirstSimulation))
		s.createMilestone(projectID, userID, models.MilestoneFirstSimulation, models.XPFirstSimulationRun)
	}

	// Simulation success milestone
	if simulationSuccess {
		if !s.hasMilestone(projectID, models.MilestoneSimulationSuccess) {
			xpEarned += models.XPSimulationSuccess
			milestonesUnlocked = append(milestonesUnlocked, string(models.MilestoneSimulationSuccess))
			s.createMilestone(projectID, userID, models.MilestoneSimulationSuccess, models.XPSimulationSuccess)
		}
	}

	// Update simulation data
	simData := map[string]interface{}{
		"last_run_at":      time.Now().Format(time.RFC3339),
		"last_run_success": simulationSuccess,
		"duration_ms":      req.DurationMs,
	}
	simJSON, _ := json.Marshal(simData)
	project.SimulationData = datatypes.JSON(simJSON)

	// Recalculate progress
	breakdown := s.calculateBreakdown(project)
	newProgress := s.calculateTotalProgress(breakdown)
	project.Progress = newProgress

	// Save project
	s.progressDB.Save(project)

	// Award XP
	if xpEarned > 0 {
		s.awardXP(userID, xpEarned, projectID, "simulation_run")
	}

	// Create simulation record
	simResult := &models.ProjectSimulationResult{
		ProjectID:  projectID,
		UserID:     userID,
		Status:     models.SimulationStatusSuccess,
		DurationMs: req.DurationMs,
		XPEarned:   xpEarned,
	}
	s.progressDB.Create(simResult)

	status := "success"
	if !simulationSuccess {
		status = "failed"
	}

	return &dto.RunSimulationResponse{
		SimulationID: simResult.ID,
		Status:       status,
		Results: &dto.SimulationResults{
			OutputData: datatypes.JSON(`{}`),
			Errors:     []string{},
			Warnings:   []string{},
		},
		XPEarned: xpEarned,
		ProgressUpdate: &dto.ProgressUpdateInfo{
			Old: oldProgress,
			New: newProgress,
		},
	}, nil
}

// CompleteProject marks project as complete
func (s *ProjectProgressService) CompleteProject(projectID, userID string) (*dto.CompleteProjectResponse, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if project.UserID != userID {
		return nil, errors.New("access denied")
	}

	if project.CompletedAt != nil {
		return nil, errors.New("project already completed")
	}

	// Check if project is ready for completion (progress >= 90%)
	if project.Progress < 90 {
		return nil, errors.New("project not ready for completion")
	}

	// Mark as complete
	now := time.Now()
	project.CompletedAt = &now
	project.Progress = 100

	xpEarned := 0
	achievementsUnlocked := []string{}

	// Project complete milestone
	if !s.hasMilestone(projectID, models.MilestoneProjectComplete) {
		xpEarned += models.XPProjectComplete
		s.createMilestone(projectID, userID, models.MilestoneProjectComplete, models.XPProjectComplete)
	}

	// Check if first project completion ever
	var completedCount int64
	s.progressDB.Model(&models.Project{}).Where("user_id = ? AND completed_at IS NOT NULL", userID).Count(&completedCount)

	if completedCount == 0 { // This is the first project completion
		if !s.hasMilestone(projectID, models.MilestoneFirstComplete) {
			xpEarned += models.XPFirstProjectCompleteBonus
			achievementsUnlocked = append(achievementsUnlocked, "first_complete")
			s.createMilestone(projectID, userID, models.MilestoneFirstComplete, models.XPFirstProjectCompleteBonus)
		}
	}

	// Save project
	s.progressDB.Save(project)

	// Award XP and check for level up
	var levelUp *dto.LevelUpInfo
	if xpEarned > 0 {
		levelUp = s.awardXPWithLevelCheck(userID, xpEarned, projectID, "project_complete")
	}

	// Calculate total XP earned for this project
	totalXP := s.calculateTotalXPEarned(projectID, userID)

	return &dto.CompleteProjectResponse{
		CompletedAt:          now,
		XPEarned:             xpEarned,
		TotalXP:              totalXP,
		AchievementsUnlocked: achievementsUnlocked,
		LevelUp:              levelUp,
	}, nil
}

// GetSchemaData returns schema data for a project
func (s *ProjectProgressService) GetSchemaData(projectID, userID string) (datatypes.JSON, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if !s.hasAccess(project, userID) {
		return nil, errors.New("access denied")
	}

	return project.SchemaData, nil
}

// GetCodeData returns code data for a project
func (s *ProjectProgressService) GetCodeData(projectID, userID string) (datatypes.JSON, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if !s.hasAccess(project, userID) {
		return nil, errors.New("access denied")
	}

	return project.CodeData, nil
}

// ============================================
// Helper Functions
// ============================================

func (s *ProjectProgressService) hasAccess(project *models.Project, userID string) bool {
	if project.UserID == userID {
		return true
	}
	// Check collaborators
	for _, c := range project.Collaborators {
		if c.UserID == userID {
			return true
		}
	}
	// Public projects are accessible too
	return project.IsPublic
}

func (s *ProjectProgressService) calculateBreakdown(project *models.Project) *dto.ProgressBreakdown {
	breakdown := &dto.ProgressBreakdown{
		Schema:       &dto.ComponentProgress{Weight: models.ProgressWeightSchema},
		Code:         &dto.ComponentProgress{Weight: models.ProgressWeightCode},
		Simulation:   &dto.ComponentProgress{Weight: models.ProgressWeightSimulation},
		Verification: &dto.ComponentProgress{Weight: models.ProgressWeightVerification},
	}

	// Schema check
	if len(project.SchemaData) > 0 && string(project.SchemaData) != "{}" && string(project.SchemaData) != "null" {
		var schemaData map[string]interface{}
		json.Unmarshal(project.SchemaData, &schemaData)
		if components, ok := schemaData["components"]; ok && components != nil {
			breakdown.Schema.Complete = true
			breakdown.Schema.Earned = models.ProgressWeightSchema
			breakdown.Schema.Percentage = 100
		}
	}

	// Code check
	if len(project.CodeData) > 0 && string(project.CodeData) != "{}" && string(project.CodeData) != "null" {
		var codeData map[string]interface{}
		json.Unmarshal(project.CodeData, &codeData)
		if content, ok := codeData["content"].(string); ok {
			charCount := len(content)
			pct := models.GetCodeCompletionPercentage(charCount)
			breakdown.Code.Percentage = pct
			breakdown.Code.Earned = models.ProgressWeightCode * pct / 100
			breakdown.Code.Complete = pct >= 100
		}
	}

	// Simulation check
	if len(project.SimulationData) > 0 && string(project.SimulationData) != "{}" && string(project.SimulationData) != "null" {
		var simData map[string]interface{}
		json.Unmarshal(project.SimulationData, &simData)
		if success, ok := simData["last_run_success"].(bool); ok && success {
			breakdown.Simulation.Complete = true
			breakdown.Simulation.Earned = models.ProgressWeightSimulation
			breakdown.Simulation.Percentage = 100
		}
	}

	// Verification - based on completion
	if project.CompletedAt != nil {
		breakdown.Verification.Complete = true
		breakdown.Verification.Earned = models.ProgressWeightVerification
		breakdown.Verification.Percentage = 100
	}

	return breakdown
}

func (s *ProjectProgressService) calculateTotalProgress(breakdown *dto.ProgressBreakdown) int {
	total := 0
	if breakdown.Schema != nil {
		total += breakdown.Schema.Earned
	}
	if breakdown.Code != nil {
		total += breakdown.Code.Earned
	}
	if breakdown.Simulation != nil {
		total += breakdown.Simulation.Earned
	}
	if breakdown.Verification != nil {
		total += breakdown.Verification.Earned
	}
	return total
}

func (s *ProjectProgressService) getMilestones(projectID string) []dto.MilestoneInfo {
	var milestones []models.ProjectMilestone
	s.progressDB.Where("project_id = ?", projectID).Order("unlocked_at ASC").Find(&milestones)

	result := make([]dto.MilestoneInfo, len(milestones))
	for i, m := range milestones {
		result[i] = dto.MilestoneInfo{
			ID:         m.ID,
			Type:       string(m.MilestoneType),
			UnlockedAt: m.UnlockedAt,
			XPEarned:   m.XPEarned,
		}
	}
	return result
}

func (s *ProjectProgressService) calculateTotalXPEarned(projectID, userID string) int {
	var total int
	s.progressDB.Model(&models.UserXPTransaction{}).
		Where("user_id = ? AND source_id = ?", userID, projectID).
		Select("COALESCE(SUM(xp_amount), 0)").
		Scan(&total)
	return total
}

func (s *ProjectProgressService) getNextAction(breakdown *dto.ProgressBreakdown) string {
	if !breakdown.Schema.Complete {
		return "Add components to your circuit to earn +15 XP"
	}
	if !breakdown.Code.Complete {
		return "Write at least 100 characters of code to earn +20 XP"
	}
	if !breakdown.Simulation.Complete {
		return "Run simulation to earn +25 XP"
	}
	if !breakdown.Verification.Complete {
		return "Complete your project to earn +50 XP bonus"
	}
	return "Project complete! Great job!"
}

func (s *ProjectProgressService) hasMilestone(projectID string, milestoneType models.MilestoneType) bool {
	var count int64
	s.progressDB.Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND milestone_type = ?", projectID, milestoneType).
		Count(&count)
	return count > 0
}

func (s *ProjectProgressService) createMilestone(projectID, userID string, milestoneType models.MilestoneType, xpEarned int) {
	milestone := &models.ProjectMilestone{
		ProjectID:     projectID,
		UserID:        userID,
		MilestoneType: milestoneType,
		XPEarned:      xpEarned,
	}
	s.progressDB.Create(milestone)
}

func (s *ProjectProgressService) awardXP(userID string, xpAmount int, sourceID, description string) {
	// Create transaction record
	tx := &models.UserXPTransaction{
		UserID:      userID,
		XPAmount:    xpAmount,
		XPType:      models.XPSourceProject,
		SourceID:    &sourceID,
		SourceType:  "project",
		Description: description,
	}
	s.progressDB.Create(tx)

	// Update user XP
	s.progressDB.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"current_xp": gorm.Expr("current_xp + ?", xpAmount),
			"total_xp":   gorm.Expr("total_xp + ?", xpAmount),
		})
}

func (s *ProjectProgressService) awardXPWithLevelCheck(userID string, xpAmount int, sourceID, description string) *dto.LevelUpInfo {
	// Get current user
	var user models.User
	s.progressDB.First(&user, "id = ?", userID)
	oldLevel := user.Level

	// Award XP
	s.awardXP(userID, xpAmount, sourceID, description)

	// Check for level up
	newXP := user.CurrentXP + xpAmount
	targetXP := user.TargetXP

	if newXP >= targetXP {
		// Level up!
		newLevel := user.Level + 1
		newTargetXP := int(float64(targetXP) * 1.5) // 1.5x multiplier
		remainingXP := newXP - targetXP

		s.progressDB.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"level":      newLevel,
				"current_xp": remainingXP,
				"target_xp":  newTargetXP,
			})

		return &dto.LevelUpInfo{
			OldLevel: oldLevel,
			NewLevel: newLevel,
			XPToNext: newTargetXP,
		}
	}

	return nil
}

func (s *ProjectProgressService) updateSchemaProgress(project *models.Project, userID string, data datatypes.JSON) (int, []string) {
	project.SchemaData = data

	xpEarned := 0
	milestones := []string{}

	isFirstSave := !s.hasMilestone(project.ID, models.MilestoneFirstSchemaSave)
	if isFirstSave {
		xpEarned = models.XPFirstSchemaSave
		milestones = append(milestones, string(models.MilestoneFirstSchemaSave))
		s.createMilestone(project.ID, userID, models.MilestoneFirstSchemaSave, xpEarned)
	}

	return xpEarned, milestones
}

func (s *ProjectProgressService) updateCodeProgress(project *models.Project, userID string, data datatypes.JSON) (int, []string) {
	project.CodeData = data

	xpEarned := 0
	milestones := []string{}

	isFirstSave := !s.hasMilestone(project.ID, models.MilestoneFirstCodeSave)
	if isFirstSave {
		xpEarned = models.XPFirstCodeSave
		milestones = append(milestones, string(models.MilestoneFirstCodeSave))
		s.createMilestone(project.ID, userID, models.MilestoneFirstCodeSave, xpEarned)
	}

	return xpEarned, milestones
}

func (s *ProjectProgressService) updateSimulationProgress(project *models.Project, userID string, data datatypes.JSON, isRun bool) (int, []string) {
	project.SimulationData = data

	xpEarned := 0
	milestones := []string{}

	if isRun {
		if !s.hasMilestone(project.ID, models.MilestoneFirstSimulation) {
			xpEarned += models.XPFirstSimulationRun
			milestones = append(milestones, string(models.MilestoneFirstSimulation))
			s.createMilestone(project.ID, userID, models.MilestoneFirstSimulation, models.XPFirstSimulationRun)
		}
	}

	return xpEarned, milestones
}

func (s *ProjectProgressService) updateVerificationProgress(project *models.Project, userID string) (int, []string) {
	xpEarned := 0
	milestones := []string{}

	// Verification is usually automatic when project is complete
	// Or can be triggered by admin/mentor

	return xpEarned, milestones
}
