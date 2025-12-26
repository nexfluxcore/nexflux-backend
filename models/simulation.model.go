package models

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Simulation Models - For simulation management
// ============================================

// SimulationType represents different types of simulations
type SimulationType string

const (
	SimTypeBasicElectronics SimulationType = "Basic Electronics"
	SimTypeIoT              SimulationType = "IoT"
	SimTypePowerElectronics SimulationType = "Power Electronics"
	SimTypeWireless         SimulationType = "Wireless"
	SimTypeRenewableEnergy  SimulationType = "Renewable Energy"
	SimTypeAudio            SimulationType = "Audio"
	SimTypeDigitalLogic     SimulationType = "Digital Logic"
)

// SimulationStatusType represents simulation status
type SimulationStatusType string

const (
	SimStatusDraft     SimulationStatusType = "draft"
	SimStatusRunning   SimulationStatusType = "running"
	SimStatusCompleted SimulationStatusType = "completed"
	SimStatusPaused    SimulationStatusType = "paused"
	SimStatusError     SimulationStatusType = "error"
)

// Simulation represents a simulation created by user
type Simulation struct {
	ID                 string               `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID             string               `gorm:"type:uuid;index;not null" json:"user_id"`
	ProjectID          *string              `gorm:"type:uuid;index" json:"project_id"`
	Name               string               `gorm:"size:255;not null" json:"name"`
	Description        string               `gorm:"type:text" json:"description"`
	Type               SimulationType       `gorm:"type:varchar(100);default:'Basic Electronics'" json:"type"`
	ThumbnailURL       string               `gorm:"size:500" json:"thumbnail_url"`
	SchemaData         datatypes.JSON       `gorm:"type:jsonb" json:"schema_data"`
	SimulationSettings datatypes.JSON       `gorm:"type:jsonb" json:"simulation_settings"`
	LastResult         datatypes.JSON       `gorm:"type:jsonb" json:"last_result"`
	ComponentsCount    int                  `gorm:"default:0" json:"components_count"`
	WiresCount         int                  `gorm:"default:0" json:"wires_count"`
	RunCount           int                  `gorm:"default:0" json:"run_count"`
	TotalRuntimeMs     int64                `gorm:"default:0" json:"total_runtime_ms"`
	Status             SimulationStatusType `gorm:"type:varchar(50);default:'draft'" json:"status"`
	ErrorMessage       string               `gorm:"type:text" json:"error_message"`
	LastRunAt          *time.Time           `json:"last_run_at"`
	CreatedAt          time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time            `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User    *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Runs    []SimulationRun `gorm:"foreignKey:SimulationID" json:"runs,omitempty"`
}

func (Simulation) TableName() string {
	return "simulations"
}

// SimulationRun represents a single simulation run history
type SimulationRun struct {
	ID           string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	SimulationID string         `gorm:"type:uuid;index;not null" json:"simulation_id"`
	UserID       string         `gorm:"type:uuid;index;not null" json:"user_id"`
	Status       string         `gorm:"type:varchar(50);not null" json:"status"`
	DurationMs   int            `gorm:"default:0" json:"duration_ms"`
	ResultData   datatypes.JSON `gorm:"type:jsonb" json:"result_data"`
	Errors       datatypes.JSON `gorm:"type:jsonb" json:"errors"`
	Warnings     datatypes.JSON `gorm:"type:jsonb" json:"warnings"`
	StartedAt    time.Time      `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt  *time.Time     `json:"completed_at"`

	// Relations
	Simulation *Simulation `gorm:"foreignKey:SimulationID" json:"-"`
	User       *User       `gorm:"foreignKey:UserID" json:"-"`
}

func (SimulationRun) TableName() string {
	return "simulation_runs"
}

// SimulationStats represents aggregated simulation statistics
type SimulationStats struct {
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
