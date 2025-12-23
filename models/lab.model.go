package models

import (
	"time"

	"gorm.io/datatypes"
)

// LabStatus represents lab status types
type LabStatus string

const (
	LabStatusAvailable   LabStatus = "available"
	LabStatusBusy        LabStatus = "busy"
	LabStatusMaintenance LabStatus = "maintenance"
	LabStatusOffline     LabStatus = "offline"
)

// LabPlatform represents hardware platform types
type LabPlatform string

const (
	PlatformArduinoUno  LabPlatform = "arduino_uno"
	PlatformArduinoMega LabPlatform = "arduino_mega"
	PlatformESP32       LabPlatform = "esp32"
	PlatformESP8266     LabPlatform = "esp8266"
	PlatformRaspberryPi LabPlatform = "raspberry_pi"
	PlatformSTM32       LabPlatform = "stm32"
)

// SessionStatus represents lab session status types
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusExpired   SessionStatus = "expired"
	SessionStatusCancelled SessionStatus = "cancelled"
)

// BookingStatus represents lab booking status types
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusNoShow    BookingStatus = "no_show"
)

// CompilationStatus represents code compilation status types
type CompilationStatus string

const (
	CompilationStatusPending   CompilationStatus = "pending"
	CompilationStatusCompiling CompilationStatus = "compiling"
	CompilationStatusUploading CompilationStatus = "uploading"
	CompilationStatusSuccess   CompilationStatus = "success"
	CompilationStatusFailed    CompilationStatus = "failed"
)

// LabEventType represents hardware log event types
type LabEventType string

const (
	EventTypeSensorRead      LabEventType = "sensor_read"
	EventTypeActuatorControl LabEventType = "actuator_control"
	EventTypeCodeUpload      LabEventType = "code_upload"
	EventTypeError           LabEventType = "error"
	EventTypeConnection      LabEventType = "connection"
	EventTypeSessionStart    LabEventType = "session_start"
	EventTypeSessionEnd      LabEventType = "session_end"
)

// Lab represents the labs table
type Lab struct {
	ID                 string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name               string         `gorm:"size:100;not null" json:"name"`
	Slug               string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description        string         `gorm:"type:text" json:"description"`
	Platform           LabPlatform    `gorm:"type:varchar(50);not null" json:"platform"`
	Status             LabStatus      `gorm:"type:varchar(20);default:'available'" json:"status"`
	HardwareSpecs      datatypes.JSON `gorm:"type:jsonb" json:"hardware_specs"`
	ThumbnailURL       string         `gorm:"size:500" json:"thumbnail_url"`
	LivekitRoomName    string         `gorm:"size:100" json:"livekit_room_name"`
	CurrentUserID      *string        `gorm:"type:uuid" json:"current_user_id"`
	SessionStartedAt   *time.Time     `json:"session_started_at"`
	MaxSessionDuration int            `gorm:"default:30" json:"max_session_duration"` // minutes
	QueueCount         int            `gorm:"default:0" json:"queue_count"`
	TotalSessions      int            `gorm:"default:0" json:"total_sessions"`
	MQTTTopic          string         `gorm:"size:255" json:"mqtt_topic"` // Topic untuk komunikasi hardware
	AgentID            string         `gorm:"size:100" json:"agent_id"`   // ID mini PC agent
	IsOnline           bool           `gorm:"default:false" json:"is_online"`
	LastHeartbeat      *time.Time     `json:"last_heartbeat"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	CurrentUser *User        `gorm:"foreignKey:CurrentUserID" json:"current_user,omitempty"`
	Sessions    []LabSession `gorm:"foreignKey:LabID" json:"sessions,omitempty"`
	Queue       []LabQueue   `gorm:"foreignKey:LabID" json:"queue,omitempty"`
}

func (Lab) TableName() string {
	return "labs"
}

// LabSession represents the lab_sessions table
type LabSession struct {
	ID                string            `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	LabID             string            `gorm:"type:uuid;index;not null" json:"lab_id"`
	UserID            string            `gorm:"type:uuid;index;not null" json:"user_id"`
	ProjectID         *string           `gorm:"type:uuid" json:"project_id"`
	Status            SessionStatus     `gorm:"type:varchar(20);default:'active'" json:"status"`
	StartedAt         time.Time         `gorm:"autoCreateTime" json:"started_at"`
	EndedAt           *time.Time        `json:"ended_at"`
	DurationSeconds   int               `json:"duration_seconds"`
	CodeSubmitted     string            `gorm:"type:text" json:"code_submitted"`
	CompilationStatus CompilationStatus `gorm:"type:varchar(20)" json:"compilation_status"`
	CompilationOutput string            `gorm:"type:text" json:"compilation_output"`
	XPEarned          int               `gorm:"default:0" json:"xp_earned"`
	Feedback          string            `gorm:"type:text" json:"feedback"`
	Rating            int               `json:"rating"`
	LivekitToken      string            `gorm:"size:1000" json:"-"` // Token is sensitive
	CreatedAt         time.Time         `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Lab     *Lab     `gorm:"foreignKey:LabID" json:"lab,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (LabSession) TableName() string {
	return "lab_sessions"
}

// LabBooking represents the lab_bookings table
type LabBooking struct {
	ID              string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	LabID           string        `gorm:"type:uuid;index;not null" json:"lab_id"`
	UserID          string        `gorm:"type:uuid;index;not null" json:"user_id"`
	ScheduledAt     time.Time     `gorm:"not null" json:"scheduled_at"`
	DurationMinutes int           `gorm:"default:30" json:"duration_minutes"`
	Status          BookingStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	BidAmount       int           `gorm:"default:0" json:"bid_amount"` // XP bid for priority
	QueuePosition   int           `json:"queue_position"`
	CreatedAt       time.Time     `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Lab  *Lab  `gorm:"foreignKey:LabID" json:"lab,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (LabBooking) TableName() string {
	return "lab_bookings"
}

// LabQueue represents the lab_queue table
type LabQueue struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	LabID     string    `gorm:"type:uuid;index;not null" json:"lab_id"`
	UserID    string    `gorm:"type:uuid;index;not null" json:"user_id"`
	Position  int       `gorm:"not null" json:"position"`
	BidAmount int       `gorm:"default:0" json:"bid_amount"`
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
	ExpiresAt time.Time `json:"expires_at"`

	// Relations
	Lab  *Lab  `gorm:"foreignKey:LabID" json:"lab,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (LabQueue) TableName() string {
	return "lab_queue"
}

// LabHardwareLog represents the lab_hardware_logs table
type LabHardwareLog struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	LabID     string         `gorm:"type:uuid;index;not null" json:"lab_id"`
	SessionID *string        `gorm:"type:uuid;index" json:"session_id"`
	EventType LabEventType   `gorm:"type:varchar(50);not null" json:"event_type"`
	EventData datatypes.JSON `gorm:"type:jsonb" json:"event_data"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Lab     *Lab        `gorm:"foreignKey:LabID" json:"lab,omitempty"`
	Session *LabSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

func (LabHardwareLog) TableName() string {
	return "lab_hardware_logs"
}

// CodeCompilation represents pending code compilation jobs
type CodeCompilation struct {
	ID          string            `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	SessionID   string            `gorm:"type:uuid;index;not null" json:"session_id"`
	LabID       string            `gorm:"type:uuid;index;not null" json:"lab_id"`
	UserID      string            `gorm:"type:uuid;index;not null" json:"user_id"`
	Code        string            `gorm:"type:text;not null" json:"code"`
	Language    string            `gorm:"size:20;not null" json:"language"` // arduino, micropython, c
	Filename    string            `gorm:"size:255" json:"filename"`
	Status      CompilationStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Output      string            `gorm:"type:text" json:"output"`
	Errors      string            `gorm:"type:text" json:"errors"`
	StartedAt   *time.Time        `json:"started_at"`
	CompletedAt *time.Time        `json:"completed_at"`
	UploadedAt  *time.Time        `json:"uploaded_at"`
	CreatedAt   time.Time         `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Session *LabSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Lab     *Lab        `gorm:"foreignKey:LabID" json:"lab,omitempty"`
	User    *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (CodeCompilation) TableName() string {
	return "code_compilations"
}

// Helper methods

// IsAvailable checks if lab is available
func (l *Lab) IsAvailable() bool {
	return l.Status == LabStatusAvailable && l.CurrentUserID == nil
}

// SessionRemainingSeconds calculates remaining session time in seconds
func (l *Lab) SessionRemainingSeconds() int {
	if l.SessionStartedAt == nil {
		return 0
	}
	elapsed := time.Since(*l.SessionStartedAt)
	maxDuration := time.Duration(l.MaxSessionDuration) * time.Minute
	remaining := maxDuration - elapsed
	if remaining < 0 {
		return 0
	}
	return int(remaining.Seconds())
}
