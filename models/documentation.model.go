package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// DocCategory difficulty types
type DocDifficulty string

const (
	DocDifficultyBeginner     DocDifficulty = "beginner"
	DocDifficultyIntermediate DocDifficulty = "intermediate"
	DocDifficultyAdvanced     DocDifficulty = "advanced"
)

// DocCategory represents a documentation category
type DocCategory struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Icon        string    `gorm:"size:50" json:"icon"`
	Color       string    `gorm:"size:100" json:"color"`
	Order       int       `gorm:"default:0" json:"order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Articles []DocArticle `gorm:"foreignKey:CategoryID" json:"articles,omitempty"`
	Videos   []DocVideo   `gorm:"foreignKey:CategoryID" json:"videos,omitempty"`
}

func (DocCategory) TableName() string {
	return "doc_categories"
}

// DocArticle represents a documentation article
type DocArticle struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CategoryID      string         `gorm:"type:uuid;index;not null" json:"category_id"`
	Title           string         `gorm:"size:300;not null" json:"title"`
	Slug            string         `gorm:"uniqueIndex;size:300;not null" json:"slug"`
	Excerpt         string         `gorm:"type:text" json:"excerpt"`
	Content         string         `gorm:"type:text;not null" json:"content"`
	AuthorID        string         `gorm:"type:uuid;index" json:"author_id"`
	ReadTimeMinutes int            `gorm:"default:5" json:"read_time_minutes"`
	Difficulty      DocDifficulty  `gorm:"type:varchar(20);default:'beginner'" json:"difficulty"`
	Tags            pq.StringArray `gorm:"type:text[]" json:"tags"`
	Views           int            `gorm:"default:0" json:"views"`
	IsFeatured      bool           `gorm:"default:false" json:"is_featured"`
	IsPublished     bool           `gorm:"default:true" json:"is_published"`
	PublishedAt     *time.Time     `json:"published_at"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Category *DocCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Author   *User        `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

func (DocArticle) TableName() string {
	return "doc_articles"
}

// BeforeCreate hook for DocArticle
func (a *DocArticle) BeforeCreate(tx *gorm.DB) error {
	if a.PublishedAt == nil && a.IsPublished {
		now := time.Now()
		a.PublishedAt = &now
	}
	return nil
}

// DocVideo represents a video tutorial
type DocVideo struct {
	ID              string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CategoryID      string        `gorm:"type:uuid;index" json:"category_id"`
	Title           string        `gorm:"size:300;not null" json:"title"`
	Description     string        `gorm:"type:text" json:"description"`
	VideoURL        string        `gorm:"size:500;not null" json:"video_url"`
	ThumbnailURL    string        `gorm:"size:500" json:"thumbnail_url"`
	DurationSeconds int           `json:"duration_seconds"`
	Difficulty      DocDifficulty `gorm:"type:varchar(20);default:'beginner'" json:"difficulty"`
	Views           int           `gorm:"default:0" json:"views"`
	IsFeatured      bool          `gorm:"default:false" json:"is_featured"`
	IsPublished     bool          `gorm:"default:true" json:"is_published"`
	CreatedAt       time.Time     `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Category *DocCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (DocVideo) TableName() string {
	return "doc_videos"
}
