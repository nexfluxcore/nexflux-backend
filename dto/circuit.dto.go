package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Circuit DTOs
// ============================================

// CircuitListRequest for listing circuits
type CircuitListRequest struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
	ProjectID string `form:"project_id"`
	Search    string `form:"search"`
	Sort      string `form:"sort,default=created_at"`
	Order     string `form:"order,default=desc"`
}

// CircuitResponse for circuit response
type CircuitResponse struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	ThumbnailURL    string         `json:"thumbnail_url"`
	SchemaData      datatypes.JSON `json:"schema_data,omitempty"`
	ComponentsCount int            `json:"components_count"`
	WiresCount      int            `json:"wires_count"`
	ProjectID       *string        `json:"project_id"`
	ProjectName     string         `json:"project_name,omitempty"`
	IsTemplate      bool           `json:"is_template"`
	IsPublic        bool           `json:"is_public"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// CircuitDetailResponse includes full schema data
type CircuitDetailResponse struct {
	CircuitResponse
	SchemaData datatypes.JSON `json:"schema_data"`
}

// CreateCircuitRequest for creating circuit
type CreateCircuitRequest struct {
	Name        string         `json:"name" binding:"required,max=255"`
	Description string         `json:"description"`
	ProjectID   *string        `json:"project_id"`
	SchemaData  datatypes.JSON `json:"schema_data" binding:"required"`
}

// UpdateCircuitRequest for updating circuit
type UpdateCircuitRequest struct {
	Name        string         `json:"name" binding:"omitempty,max=255"`
	Description string         `json:"description"`
	SchemaData  datatypes.JSON `json:"schema_data"`
	IsPublic    *bool          `json:"is_public"`
}

// CircuitThumbnailRequest for uploading thumbnail
type CircuitThumbnailRequest struct {
	ImageData string `json:"image_data" binding:"required"` // base64 encoded
}

// CircuitThumbnailResponse for thumbnail upload response
type CircuitThumbnailResponse struct {
	ThumbnailURL string `json:"thumbnail_url"`
}

// CircuitExportResponse for export response
type CircuitExportResponse struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// ============================================
// Circuit Template DTOs
// ============================================

// CircuitTemplateListRequest for listing templates
type CircuitTemplateListRequest struct {
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
	Search     string `form:"search"`
}

// CircuitTemplateResponse for template response
type CircuitTemplateResponse struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	Description          string         `json:"description"`
	Category             string         `json:"category"`
	Difficulty           string         `json:"difficulty"`
	ThumbnailURL         string         `json:"thumbnail_url"`
	SchemaData           datatypes.JSON `json:"schema_data,omitempty"`
	EstimatedTimeMinutes int            `json:"estimated_time_minutes"`
	XPReward             int            `json:"xp_reward"`
	Tags                 []string       `json:"tags"`
	UseCount             int            `json:"use_count"`
	CreatedAt            time.Time      `json:"created_at"`
}

// UseTemplateRequest for using a template
type UseTemplateRequest struct {
	Name      string  `json:"name"`
	ProjectID *string `json:"project_id"`
}
