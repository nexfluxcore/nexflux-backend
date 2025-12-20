package models

import (
	"time"

	"gorm.io/datatypes"
)

// ComponentCategory represents component categories
type ComponentCategory struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Slug        string    `gorm:"size:100;uniqueIndex" json:"slug"`
	Icon        string    `gorm:"size:50" json:"icon"`
	Color       string    `gorm:"size:50" json:"color"`
	Description string    `gorm:"type:text" json:"description"`
	Order       int       `gorm:"default:0" json:"order"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Components []Component `gorm:"foreignKey:CategoryID" json:"components,omitempty"`
}

func (ComponentCategory) TableName() string {
	return "component_categories"
}

// Component represents electronic components
type Component struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CategoryID      string         `gorm:"type:uuid;index" json:"category_id"`
	Name            string         `gorm:"size:200;not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description"`
	Manufacturer    string         `gorm:"size:100" json:"manufacturer"`
	PartNumber      string         `gorm:"size:100" json:"part_number"`
	Specs           datatypes.JSON `gorm:"type:jsonb" json:"specs"`
	Price           float64        `gorm:"type:decimal(12,2)" json:"price"`
	Stock           int            `gorm:"default:0" json:"stock"`
	Rating          float64        `gorm:"type:decimal(2,1);default:0" json:"rating"`
	RatingCount     int            `gorm:"default:0" json:"rating_count"`
	ImageURL        string         `gorm:"size:500" json:"image_url"`
	DatasheetURL    string         `gorm:"size:500" json:"datasheet_url"`
	SimulationModel string         `gorm:"size:100" json:"simulation_model"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Category *ComponentCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Component) TableName() string {
	return "components"
}

// RequestPriority represents component request priority
type RequestPriority string

const (
	PriorityLow    RequestPriority = "low"
	PriorityMedium RequestPriority = "medium"
	PriorityHigh   RequestPriority = "high"
	PriorityUrgent RequestPriority = "urgent"
)

// RequestStatus represents component request status
type RequestStatus string

const (
	StatusPending   RequestStatus = "pending"
	StatusReviewing RequestStatus = "reviewing"
	StatusApproved  RequestStatus = "approved"
	StatusRejected  RequestStatus = "rejected"
)

// ComponentRequest represents user component requests
type ComponentRequest struct {
	ID            string          `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID        string          `gorm:"type:uuid;index;not null" json:"user_id"`
	ComponentName string          `gorm:"size:200;not null" json:"component_name"`
	Manufacturer  string          `gorm:"size:100" json:"manufacturer"`
	PartNumber    string          `gorm:"size:100" json:"part_number"`
	Category      string          `gorm:"size:50;not null" json:"category"`
	Description   string          `gorm:"type:text;not null" json:"description"`
	UseCase       string          `gorm:"type:text;not null" json:"use_case"`
	Features      datatypes.JSON  `gorm:"type:jsonb" json:"features"`
	DatasheetURL  string          `gorm:"size:500" json:"datasheet_url"`
	ProductURL    string          `gorm:"size:500" json:"product_url"`
	Priority      RequestPriority `gorm:"type:varchar(20);default:'medium'" json:"priority"`
	Status        RequestStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	AdminNotes    string          `gorm:"type:text" json:"admin_notes"`
	CreatedAt     time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ComponentRequest) TableName() string {
	return "component_requests"
}

// InStock checks if component is in stock
func (c *Component) InStock() bool {
	return c.Stock > 0 && c.IsActive
}

// FormattedPrice returns formatted price string
func (c *Component) FormattedPrice() string {
	if c.Price == 0 {
		return "Free"
	}
	return "Rp " + formatNumber(int64(c.Price))
}

func formatNumber(n int64) string {
	if n < 1000 {
		return string(rune(n))
	}
	// Simple formatting, can be enhanced
	return string(rune(n/1000)) + "." + string(rune(n%1000)) + "K"
}
