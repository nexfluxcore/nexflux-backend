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
