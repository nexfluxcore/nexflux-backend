package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// Difficulty represents project/challenge difficulty levels
type Difficulty string

const (
	DifficultyBeginner     Difficulty = "Beginner"
	DifficultyIntermediate Difficulty = "Intermediate"
	DifficultyAdvanced     Difficulty = "Advanced"
)

// CollaboratorRole represents project collaborator roles
type CollaboratorRole string

const (
	CollabRoleOwner  CollaboratorRole = "owner"
	CollabRoleEditor CollaboratorRole = "editor"
	CollabRoleViewer CollaboratorRole = "viewer"
)

// Project represents user projects
type Project struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID           string         `gorm:"type:uuid;index;not null" json:"user_id"`
	Name             string         `gorm:"size:200;not null" json:"name"`
	Description      string         `gorm:"type:text" json:"description"`
	ThumbnailURL     string         `gorm:"size:500" json:"thumbnail_url"`
	Difficulty       Difficulty     `gorm:"type:varchar(20);default:'Beginner'" json:"difficulty"`
	Progress         int            `gorm:"default:0" json:"progress"`
	XPReward         int            `gorm:"default:100" json:"xp_reward"`
	IsPublic         bool           `gorm:"default:false" json:"is_public"`
	IsFavorite       bool           `gorm:"default:false" json:"is_favorite"`
	IsTemplate       bool           `gorm:"default:false" json:"is_template"`
	HardwarePlatform string         `gorm:"size:50" json:"hardware_platform"`
	Tags             pq.StringArray `gorm:"type:text[]" json:"tags"`
	SchemaData       datatypes.JSON `gorm:"type:jsonb" json:"schema_data"`
	CodeData         datatypes.JSON `gorm:"type:jsonb" json:"code_data"`
	SimulationData   datatypes.JSON `gorm:"type:jsonb" json:"simulation_data"`
	CompletedAt      *time.Time     `json:"completed_at"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User          *User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Components    []ProjectComponent    `gorm:"foreignKey:ProjectID" json:"components,omitempty"`
	Collaborators []ProjectCollaborator `gorm:"foreignKey:ProjectID" json:"collaborators,omitempty"`
}

func (Project) TableName() string {
	return "projects"
}

// ProjectCollaborator represents project collaborators
type ProjectCollaborator struct {
	ID         string           `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID  string           `gorm:"type:uuid;index;not null" json:"project_id"`
	UserID     string           `gorm:"type:uuid;index;not null" json:"user_id"`
	Role       CollaboratorRole `gorm:"type:varchar(20);default:'viewer'" json:"role"`
	InvitedAt  time.Time        `gorm:"autoCreateTime" json:"invited_at"`
	AcceptedAt *time.Time       `json:"accepted_at"`

	Project *Project `gorm:"foreignKey:ProjectID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ProjectCollaborator) TableName() string {
	return "project_collaborators"
}

// ProjectComponent represents components used in a project
type ProjectComponent struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID   string         `gorm:"type:uuid;index;not null" json:"project_id"`
	ComponentID string         `gorm:"type:uuid;index;not null" json:"component_id"`
	Quantity    int            `gorm:"default:1" json:"quantity"`
	PositionX   float64        `json:"position_x"`
	PositionY   float64        `json:"position_y"`
	Rotation    float64        `gorm:"default:0" json:"rotation"`
	ConfigData  datatypes.JSON `gorm:"type:jsonb" json:"config_data"`

	Project   *Project   `gorm:"foreignKey:ProjectID" json:"-"`
	Component *Component `gorm:"foreignKey:ComponentID" json:"component,omitempty"`
}

func (ProjectComponent) TableName() string {
	return "project_components"
}

// IsCompleted checks if project is completed
func (p *Project) IsCompleted() bool {
	return p.Progress >= 100 && p.CompletedAt != nil
}

// ComponentsCount returns total components used
func (p *Project) ComponentsCount() int {
	total := 0
	for _, c := range p.Components {
		total += c.Quantity
	}
	return total
}
