package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Project DTOs
// ============================================

// ProjectListRequest for listing projects
type ProjectListRequest struct {
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
	Search     string `form:"search"`
	Difficulty string `form:"difficulty"`
	IsFavorite *bool  `form:"is_favorite"`
	IsPublic   *bool  `form:"is_public"`
	Sort       string `form:"sort,default=created_at"`
	Order      string `form:"order,default=desc"`
}

// ProjectResponse for single project response
type ProjectResponse struct {
	ID                 string        `json:"id"`
	UserID             string        `json:"user_id"`
	Name               string        `json:"name"`
	Description        string        `json:"description"`
	ThumbnailURL       string        `json:"thumbnail_url"`
	Difficulty         string        `json:"difficulty"`
	Progress           int           `json:"progress"`
	XPReward           int           `json:"xp_reward"`
	IsPublic           bool          `json:"is_public"`
	IsFavorite         bool          `json:"is_favorite"`
	IsTemplate         bool          `json:"is_template"`
	HardwarePlatform   string        `json:"hardware_platform"`
	Tags               []string      `json:"tags"`
	ComponentsCount    int           `json:"components_count"`
	CollaboratorsCount int           `json:"collaborators_count"`
	CompletedAt        *time.Time    `json:"completed_at"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	User               *UserResponse `json:"user,omitempty"`
}

// ProjectDetailResponse includes schema/code data
type ProjectDetailResponse struct {
	ProjectResponse
	SchemaData     datatypes.JSON             `json:"schema_data"`
	CodeData       datatypes.JSON             `json:"code_data"`
	SimulationData datatypes.JSON             `json:"simulation_data"`
	Components     []ProjectComponentResponse `json:"components"`
	Collaborators  []CollaboratorResponse     `json:"collaborators"`
}

// ProjectCreateRequest for creating new project
type ProjectCreateRequest struct {
	Name             string   `json:"name" binding:"required,max=200"`
	Description      string   `json:"description"`
	Difficulty       string   `json:"difficulty" binding:"omitempty,oneof=Beginner Intermediate Advanced"`
	HardwarePlatform string   `json:"hardware_platform"`
	Tags             []string `json:"tags"`
	IsPublic         bool     `json:"is_public"`
	IsTemplate       bool     `json:"is_template"`
}

// ProjectUpdateRequest for updating project
type ProjectUpdateRequest struct {
	Name             string         `json:"name" binding:"omitempty,max=200"`
	Description      string         `json:"description"`
	ThumbnailURL     string         `json:"thumbnail_url"`
	Difficulty       string         `json:"difficulty" binding:"omitempty,oneof=Beginner Intermediate Advanced"`
	Progress         *int           `json:"progress" binding:"omitempty,min=0,max=100"`
	HardwarePlatform string         `json:"hardware_platform"`
	Tags             []string       `json:"tags"`
	IsPublic         *bool          `json:"is_public"`
	IsFavorite       *bool          `json:"is_favorite"`
	SchemaData       datatypes.JSON `json:"schema_data"`
	CodeData         datatypes.JSON `json:"code_data"`
	SimulationData   datatypes.JSON `json:"simulation_data"`
}

// ============================================
// Project Collaborator DTOs
// ============================================

// CollaboratorResponse for collaborator info
type CollaboratorResponse struct {
	ID         string        `json:"id"`
	UserID     string        `json:"user_id"`
	Role       string        `json:"role"`
	InvitedAt  time.Time     `json:"invited_at"`
	AcceptedAt *time.Time    `json:"accepted_at"`
	User       *UserResponse `json:"user,omitempty"`
}

// AddCollaboratorRequest for adding collaborator
type AddCollaboratorRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=editor viewer"`
}

// ============================================
// Project Component DTOs
// ============================================

// ProjectComponentResponse for component in project
type ProjectComponentResponse struct {
	ID          string             `json:"id"`
	ComponentID string             `json:"component_id"`
	Quantity    int                `json:"quantity"`
	PositionX   float64            `json:"position_x"`
	PositionY   float64            `json:"position_y"`
	Rotation    float64            `json:"rotation"`
	ConfigData  datatypes.JSON     `json:"config_data"`
	Component   *ComponentResponse `json:"component,omitempty"`
}

// AddProjectComponentRequest for adding component to project
type AddProjectComponentRequest struct {
	ComponentID string         `json:"component_id" binding:"required,uuid"`
	Quantity    int            `json:"quantity" binding:"omitempty,min=1"`
	PositionX   float64        `json:"position_x"`
	PositionY   float64        `json:"position_y"`
	Rotation    float64        `json:"rotation"`
	ConfigData  datatypes.JSON `json:"config_data"`
}

// ============================================
// Project Progress DTOs
// ============================================

// ProjectProgressUpdateRequest for updating project progress
type ProjectProgressUpdateRequest struct {
	Component string         `json:"component" binding:"required,oneof=schema code simulation verification"`
	Data      datatypes.JSON `json:"data"`
	Action    string         `json:"action" binding:"required,oneof=save run verify"`
}

// UpdateProgressResponse for progress update response
type UpdateProgressResponse struct {
	Progress           int                `json:"progress"`
	ProgressChange     int                `json:"progress_change"`
	XPEarned           int                `json:"xp_earned"`
	MilestonesUnlocked []string           `json:"milestones_unlocked"`
	Breakdown          *ProgressBreakdown `json:"breakdown,omitempty"`
}

// ProgressBreakdown shows progress for each component
type ProgressBreakdown struct {
	Schema       *ComponentProgress `json:"schema"`
	Code         *ComponentProgress `json:"code"`
	Simulation   *ComponentProgress `json:"simulation"`
	Verification *ComponentProgress `json:"verification"`
}

// ComponentProgress shows progress for a single component
type ComponentProgress struct {
	Complete   bool `json:"complete"`
	Weight     int  `json:"weight"`
	Earned     int  `json:"earned"`
	Percentage int  `json:"percentage"`
}

// GetProgressResponse for detailed progress info
type GetProgressResponse struct {
	Progress   int                `json:"progress"`
	Breakdown  *ProgressBreakdown `json:"breakdown"`
	Milestones []MilestoneInfo    `json:"milestones"`
	NextAction string             `json:"next_action"`
	TotalXP    int                `json:"total_xp_earned"`
}

// MilestoneInfo shows milestone details
type MilestoneInfo struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	UnlockedAt time.Time `json:"unlocked_at"`
	XPEarned   int       `json:"xp_earned"`
}

// ============================================
// Schema Save DTOs
// ============================================

// SaveSchemaRequest for saving schema/circuit data
type SaveSchemaRequest struct {
	Components     datatypes.JSON `json:"components" binding:"required"`
	Connections    datatypes.JSON `json:"connections"`
	CanvasSettings datatypes.JSON `json:"canvas_settings"`
}

// SaveSchemaResponse for schema save response
type SaveSchemaResponse struct {
	Success        bool `json:"success"`
	Progress       int  `json:"progress"`
	ProgressChange int  `json:"progress_change"`
	XPEarned       int  `json:"xp_earned"`
	IsFirstSave    bool `json:"is_first_save"`
}

// ============================================
// Code Save DTOs
// ============================================

// SaveCodeRequest for saving code data
type SaveCodeRequest struct {
	Content  string `json:"content" binding:"required"`
	Language string `json:"language" binding:"required,oneof=arduino micropython c cpp python"`
	Filename string `json:"filename"`
}

// SaveCodeResponse for code save response
type SaveCodeResponse struct {
	Success        bool `json:"success"`
	Progress       int  `json:"progress"`
	ProgressChange int  `json:"progress_change"`
	XPEarned       int  `json:"xp_earned"`
	IsFirstSave    bool `json:"is_first_save"`
	CharCount      int  `json:"char_count"`
}

// ============================================
// Simulation DTOs
// ============================================

// RunSimulationRequest for running simulation
type RunSimulationRequest struct {
	DurationMs      int     `json:"duration_ms" binding:"omitempty,min=100,max=60000"`
	SpeedMultiplier float64 `json:"speed_multiplier" binding:"omitempty,min=0.1,max=10"`
}

// RunSimulationResponse for simulation response
type RunSimulationResponse struct {
	SimulationID   string              `json:"simulation_id"`
	Status         string              `json:"status"` // success, failed, timeout
	Results        *SimulationResults  `json:"results"`
	XPEarned       int                 `json:"xp_earned"`
	ProgressUpdate *ProgressUpdateInfo `json:"progress_update"`
}

// SimulationResults contains simulation output
type SimulationResults struct {
	OutputData datatypes.JSON `json:"output_data"`
	Errors     []string       `json:"errors"`
	Warnings   []string       `json:"warnings"`
}

// ProgressUpdateInfo shows progress change
type ProgressUpdateInfo struct {
	Old int `json:"old"`
	New int `json:"new"`
}

// ============================================
// Complete Project DTOs
// ============================================

// CompleteProjectResponse for project completion
type CompleteProjectResponse struct {
	CompletedAt          time.Time    `json:"completed_at"`
	XPEarned             int          `json:"xp_earned"`
	TotalXP              int          `json:"total_xp"`
	AchievementsUnlocked []string     `json:"achievements_unlocked"`
	LevelUp              *LevelUpInfo `json:"level_up,omitempty"`
}

// LevelUpInfo shows level up details
type LevelUpInfo struct {
	OldLevel int `json:"old_level"`
	NewLevel int `json:"new_level"`
	XPToNext int `json:"xp_to_next"`
}
