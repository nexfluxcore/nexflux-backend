package models

import (
	"time"

	"gorm.io/datatypes"
)

// NotificationType represents notification types
type NotificationType string

const (
	NotifAchievement    NotificationType = "achievement"
	NotifXP             NotificationType = "xp"
	NotifProject        NotificationType = "project"
	NotifChallenge      NotificationType = "challenge"
	NotifSocial         NotificationType = "social"
	NotifSystem         NotificationType = "system"
	NotifLabAvailable   NotificationType = "lab_available"
	NotifSessionExpired NotificationType = "session_expired"
	NotifQueueExpired   NotificationType = "queue_expired"
	NotifLevelUp        NotificationType = "level_up"
)

// Notification represents user notifications
type Notification struct {
	ID        string           `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID    string           `gorm:"type:uuid;index;not null" json:"user_id"`
	Type      NotificationType `gorm:"type:varchar(20);not null" json:"type"`
	Title     string           `gorm:"size:200;not null" json:"title"`
	Message   string           `gorm:"type:text;not null" json:"message"`
	Data      datatypes.JSON   `gorm:"type:jsonb" json:"data"`
	IsRead    bool             `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time        `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}

// MarkAsRead marks notification as read
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}
