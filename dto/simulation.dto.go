package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Simulation DTOs
// ============================================

// SimulationListRequest for listing simulations
type SimulationListRequest struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
	Status    string `form:"status"` // draft, running, completed, paused, error
	Type      string `form:"type"`   // Basic Electronics, IoT, etc.
	ProjectID string `form:"project_id"`
	Search    string `form:"search"`
	Sort      string `form:"sort,default=created_at"`
	Order     string `form:"order,default=desc"`
}

// SimulationResponse for simulation list response
type SimulationResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	ThumbnailURL    string     `json:"thumbnail_url"`
	ComponentsCount int        `json:"components_count"`
	WiresCount      int        `json:"wires_count"`
	RunCount        int        `json:"run_count"`
	TotalRuntimeMs  int64      `json:"total_runtime_ms"`
	LastRunAt       *time.Time `json:"last_run_at"`
	ProjectID       *string    `json:"project_id"`
	ProjectName     string     `json:"project_name,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// SimulationDetailResponse includes full data
type SimulationDetailResponse struct {
	SimulationResponse
	SchemaData         datatypes.JSON `json:"schema_data"`
	SimulationSettings datatypes.JSON `json:"simulation_settings"`
	LastResult         datatypes.JSON `json:"last_result"`
	ErrorMessage       string         `json:"error_message,omitempty"`
}

// CreateSimulationRequest for creating simulation
type CreateSimulationRequest struct {
	Name               string         `json:"name" binding:"required,max=255"`
	Description        string         `json:"description"`
	Type               string         `json:"type" binding:"omitempty,oneof='Basic Electronics' IoT 'Power Electronics' Wireless 'Renewable Energy' Audio 'Digital Logic'"`
	ProjectID          *string        `json:"project_id"`
	SchemaData         datatypes.JSON `json:"schema_data"`
	SimulationSettings datatypes.JSON `json:"simulation_settings"`
}

// UpdateSimulationRequest for updating simulation
type UpdateSimulationRequest struct {
	Name               string         `json:"name" binding:"omitempty,max=255"`
	Description        string         `json:"description"`
	Type               string         `json:"type"`
	SchemaData         datatypes.JSON `json:"schema_data"`
	SimulationSettings datatypes.JSON `json:"simulation_settings"`
}

// SimulationStatsResponse for stats endpoint
type SimulationStatsResponse struct {
	TotalSimulations    int            `json:"total_simulations"`
	RunningNow          int            `json:"running_now"`
	Completed           int            `json:"completed"`
	Paused              int            `json:"paused"`
	Error               int            `json:"error"`
	TotalRuntimeHours   float64        `json:"total_runtime_hours"`
	SuccessRate         float64        `json:"success_rate"`
	SimulationsThisWeek int            `json:"simulations_this_week"`
	ByType              map[string]int `json:"by_type"`
}

// ============================================
// Simulation Run DTOs
// ============================================

// RunSimulationRequestDTO for starting a simulation (renamed to avoid conflict)
type RunSimulationRequestDTO struct {
	DurationMs       int            `json:"duration_ms" binding:"omitempty,min=100,max=300000"`
	SettingsOverride datatypes.JSON `json:"settings_override"`
}

// RunSimulationResponseDTO for run response
type RunSimulationResponseDTO struct {
	RunID     string    `json:"run_id"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

// StopSimulationResponse for stop response
type StopSimulationResponse struct {
	Status     string `json:"status"`
	DurationMs int    `json:"duration_ms"`
}

// SimulationRunResponse for run history
type SimulationRunResponse struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	DurationMs  int            `json:"duration_ms"`
	ResultData  datatypes.JSON `json:"result_data"`
	Errors      datatypes.JSON `json:"errors"`
	Warnings    datatypes.JSON `json:"warnings"`
	StartedAt   time.Time      `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
}

// SaveSimulationResultRequest for saving simulation result
type SaveSimulationResultRequest struct {
	DurationMs int            `json:"duration_ms" binding:"required"`
	ResultData datatypes.JSON `json:"result_data" binding:"required"`
	Errors     datatypes.JSON `json:"errors"`
	Warnings   datatypes.JSON `json:"warnings"`
}

// SaveSimulationResultResponse for save result response
type SaveSimulationResultResponse struct {
	RunID    string `json:"run_id"`
	XPEarned int    `json:"xp_earned"`
}
