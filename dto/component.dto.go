package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Component Category DTOs
// ============================================

// CategoryResponse for component category
type CategoryResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Icon           string `json:"icon"`
	Color          string `json:"color"`
	Description    string `json:"description"`
	Order          int    `json:"order"`
	ComponentCount int    `json:"component_count,omitempty"`
}

// ============================================
// Component DTOs
// ============================================

// ComponentListRequest for listing components
type ComponentListRequest struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Category string `form:"category"`
	Search   string `form:"search"`
	InStock  *bool  `form:"in_stock"`
	Sort     string `form:"sort,default=name"`
	Order    string `form:"order,default=asc"`
}

// ComponentResponse for single component
type ComponentResponse struct {
	ID              string            `json:"id"`
	CategoryID      string            `json:"category_id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Manufacturer    string            `json:"manufacturer"`
	PartNumber      string            `json:"part_number"`
	Specs           datatypes.JSON    `json:"specs"`
	Price           float64           `json:"price"`
	Stock           int               `json:"stock"`
	Rating          float64           `json:"rating"`
	RatingCount     int               `json:"rating_count"`
	ImageURL        string            `json:"image_url"`
	DatasheetURL    string            `json:"datasheet_url"`
	SimulationModel string            `json:"simulation_model"`
	IsActive        bool              `json:"is_active"`
	IsFavorite      bool              `json:"is_favorite,omitempty"`
	Category        *CategoryResponse `json:"category,omitempty"`
}

// ComponentSearchRequest for searching components
type ComponentSearchRequest struct {
	Query    string `form:"q" binding:"required,min=2"`
	Category string `form:"category"`
	Limit    int    `form:"limit,default=10"`
}

// ============================================
// Component Request DTOs
// ============================================

// ComponentRequestCreateRequest for requesting new component
type ComponentRequestCreateRequest struct {
	ComponentName string   `json:"component_name" binding:"required,max=200"`
	Manufacturer  string   `json:"manufacturer"`
	PartNumber    string   `json:"part_number"`
	Category      string   `json:"category" binding:"required"`
	Description   string   `json:"description" binding:"required"`
	UseCase       string   `json:"use_case" binding:"required"`
	Features      []string `json:"features"`
	DatasheetURL  string   `json:"datasheet_url"`
	ProductURL    string   `json:"product_url"`
	Priority      string   `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
}

// ComponentRequestResponse for component request
type ComponentRequestResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ComponentName string    `json:"component_name"`
	Manufacturer  string    `json:"manufacturer"`
	PartNumber    string    `json:"part_number"`
	Category      string    `json:"category"`
	Description   string    `json:"description"`
	UseCase       string    `json:"use_case"`
	Features      []string  `json:"features"`
	DatasheetURL  string    `json:"datasheet_url"`
	ProductURL    string    `json:"product_url"`
	Priority      string    `json:"priority"`
	Status        string    `json:"status"`
	AdminNotes    string    `json:"admin_notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
