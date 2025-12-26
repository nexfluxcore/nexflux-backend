package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// ============================================
// Circuit Model - For storing reusable circuit schematics
// ============================================

// Circuit represents a circuit schematic created by user
type Circuit struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID          string         `gorm:"type:uuid;index;not null" json:"user_id"`
	ProjectID       *string        `gorm:"type:uuid;index" json:"project_id"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description"`
	ThumbnailURL    string         `gorm:"size:500" json:"thumbnail_url"`
	SchemaData      datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"schema_data"`
	ComponentsCount int            `gorm:"default:0" json:"components_count"`
	WiresCount      int            `gorm:"default:0" json:"wires_count"`
	IsTemplate      bool           `gorm:"default:false" json:"is_template"`
	IsPublic        bool           `gorm:"default:false" json:"is_public"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (Circuit) TableName() string {
	return "circuits"
}

// CircuitTemplate represents pre-built circuit templates for learning
type CircuitTemplate struct {
	ID                   string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name                 string         `gorm:"size:255;not null" json:"name"`
	Description          string         `gorm:"type:text" json:"description"`
	Category             string         `gorm:"size:100;not null" json:"category"` // beginner, intermediate, advanced
	ThumbnailURL         string         `gorm:"size:500" json:"thumbnail_url"`
	SchemaData           datatypes.JSON `gorm:"type:jsonb;not null" json:"schema_data"`
	Difficulty           string         `gorm:"size:50;default:'Beginner'" json:"difficulty"`
	EstimatedTimeMinutes int            `gorm:"default:15" json:"estimated_time_minutes"`
	XPReward             int            `gorm:"default:10" json:"xp_reward"`
	Tags                 pq.StringArray `gorm:"type:text[]" json:"tags"`
	UseCount             int            `gorm:"default:0" json:"use_count"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (CircuitTemplate) TableName() string {
	return "circuit_templates"
}
