package models

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Project Progress & XP System Models
// ============================================

// ProjectProgressComponent represents progress component types
type ProjectProgressComponent string

const (
	ProgressComponentSchema       ProjectProgressComponent = "schema"
	ProgressComponentCode         ProjectProgressComponent = "code"
	ProgressComponentSimulation   ProjectProgressComponent = "simulation"
	ProgressComponentVerification ProjectProgressComponent = "verification"
)

// Progress weights (from PROJECT.md)
const (
	ProgressWeightSchema       = 30
	ProgressWeightCode         = 35
	ProgressWeightSimulation   = 25
	ProgressWeightVerification = 10
)

// XP rewards (from PROJECT.md)
const (
	XPProjectCreate             = 10
	XPFirstSchemaSave           = 15
	XPFirstCodeSave             = 20
	XPFirstSimulationRun        = 10
	XPSimulationSuccess         = 25
	XPProjectComplete           = 50
	XPFirstProjectCompleteBonus = 100
)

// MilestoneType represents project milestone types
type MilestoneType string

const (
	MilestoneProjectCreate     MilestoneType = "project_create"
	MilestoneFirstSchemaSave   MilestoneType = "first_schema_save"
	MilestoneFirstCodeSave     MilestoneType = "first_code_save"
	MilestoneFirstSimulation   MilestoneType = "first_simulation_run"
	MilestoneSimulationSuccess MilestoneType = "simulation_success"
	MilestoneProjectComplete   MilestoneType = "project_complete"
	MilestoneFirstComplete     MilestoneType = "first_project_complete"
)

// XPSourceType represents the source of XP transactions
type XPSourceType string

const (
	XPSourceProject   XPSourceType = "project"
	XPSourceChallenge XPSourceType = "challenge"
	XPSourceLab       XPSourceType = "lab"
	XPSourceDaily     XPSourceType = "daily"
)

// ProjectProgress tracks progress for each component of a project
type ProjectProgress struct {
	ID                   string                   `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID            string                   `gorm:"type:uuid;index;not null" json:"project_id"`
	Component            ProjectProgressComponent `gorm:"type:varchar(50);not null" json:"component"`
	IsComplete           bool                     `gorm:"default:false" json:"is_complete"`
	CompletionPercentage int                      `gorm:"default:0" json:"completion_percentage"` // 0-100
	CompletedAt          *time.Time               `json:"completed_at"`
	Data                 datatypes.JSON           `gorm:"type:jsonb" json:"data"`
	CreatedAt            time.Time                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time                `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Project *Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (ProjectProgress) TableName() string {
	return "project_progress"
}

// ProjectMilestone tracks milestones achieved for a project
type ProjectMilestone struct {
	ID            string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID     string        `gorm:"type:uuid;index;not null" json:"project_id"`
	UserID        string        `gorm:"type:uuid;index;not null" json:"user_id"`
	MilestoneType MilestoneType `gorm:"type:varchar(100);not null" json:"milestone_type"`
	XPEarned      int           `gorm:"default:0" json:"xp_earned"`
	UnlockedAt    time.Time     `gorm:"autoCreateTime" json:"unlocked_at"`

	// Relations
	Project *Project `gorm:"foreignKey:ProjectID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"-"`
}

func (ProjectMilestone) TableName() string {
	return "project_milestones"
}

// GetMilestoneXP returns XP reward for a milestone type
func GetMilestoneXP(m MilestoneType) int {
	switch m {
	case MilestoneProjectCreate:
		return XPProjectCreate
	case MilestoneFirstSchemaSave:
		return XPFirstSchemaSave
	case MilestoneFirstCodeSave:
		return XPFirstCodeSave
	case MilestoneFirstSimulation:
		return XPFirstSimulationRun
	case MilestoneSimulationSuccess:
		return XPSimulationSuccess
	case MilestoneProjectComplete:
		return XPProjectComplete
	case MilestoneFirstComplete:
		return XPFirstProjectCompleteBonus
	default:
		return 0
	}
}

// UserXPTransaction tracks all XP transactions for audit trail
type UserXPTransaction struct {
	ID          string       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID      string       `gorm:"type:uuid;index;not null" json:"user_id"`
	XPAmount    int          `gorm:"not null" json:"xp_amount"`
	XPType      XPSourceType `gorm:"type:varchar(50);not null" json:"xp_type"`
	SourceID    *string      `gorm:"type:uuid" json:"source_id"`          // project_id, challenge_id, etc.
	SourceType  string       `gorm:"type:varchar(50)" json:"source_type"` // "project", "challenge", "lab_session"
	Description string       `gorm:"type:text" json:"description"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (UserXPTransaction) TableName() string {
	return "user_xp_transactions"
}

// SimulationStatus represents simulation result status
type SimulationStatus string

const (
	SimulationStatusSuccess SimulationStatus = "success"
	SimulationStatusFailed  SimulationStatus = "failed"
	SimulationStatusTimeout SimulationStatus = "timeout"
	SimulationStatusRunning SimulationStatus = "running"
)

// ProjectSimulationResult stores simulation results
type ProjectSimulationResult struct {
	ID         string           `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID  string           `gorm:"type:uuid;index;not null" json:"project_id"`
	UserID     string           `gorm:"type:uuid;index;not null" json:"user_id"`
	Status     SimulationStatus `gorm:"type:varchar(20);not null" json:"status"`
	DurationMs int              `gorm:"default:0" json:"duration_ms"`
	Results    datatypes.JSON   `gorm:"type:jsonb" json:"results"`
	Errors     datatypes.JSON   `gorm:"type:jsonb" json:"errors"`
	Warnings   datatypes.JSON   `gorm:"type:jsonb" json:"warnings"`
	XPEarned   int              `gorm:"default:0" json:"xp_earned"`
	CreatedAt  time.Time        `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Project *Project `gorm:"foreignKey:ProjectID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"-"`
}

func (ProjectSimulationResult) TableName() string {
	return "project_simulation_results"
}

// ============================================
// Helper Functions for Progress Calculation
// ============================================

// CalculateTotalProgress calculates total progress based on component weights
func CalculateTotalProgress(schemaComplete, codeComplete, simulationComplete, verificationComplete bool, codeCompletionPct int) int {
	var total float64

	// Schema: 30%
	if schemaComplete {
		total += float64(ProgressWeightSchema)
	}

	// Code: 35% (partial based on completion %)
	if codeComplete {
		total += float64(ProgressWeightCode)
	} else if codeCompletionPct > 0 {
		total += float64(ProgressWeightCode) * float64(codeCompletionPct) / 100.0
	}

	// Simulation: 25%
	if simulationComplete {
		total += float64(ProgressWeightSimulation)
	}

	// Verification: 10%
	if verificationComplete {
		total += float64(ProgressWeightVerification)
	}

	return int(total)
}

// GetCodeCompletionPercentage calculates code completion based on character count
func GetCodeCompletionPercentage(charCount int) int {
	if charCount >= 100 {
		return 100
	}
	if charCount >= 50 {
		return 50
	}
	return 0
}
