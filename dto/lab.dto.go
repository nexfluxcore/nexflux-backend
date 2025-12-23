package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================
// Lab DTOs
// ============================================

// LabListRequest for listing labs
type LabListRequest struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
	Platform string `form:"platform"`
	Status   string `form:"status"`
	Search   string `form:"search"`
}

// LabResponse for single lab response
type LabResponse struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	Description        string                 `json:"description"`
	Platform           string                 `json:"platform"`
	Status             string                 `json:"status"`
	HardwareSpecs      map[string]interface{} `json:"hardware_specs"`
	ThumbnailURL       string                 `json:"thumbnail_url"`
	MaxSessionDuration int                    `json:"max_session_duration"`
	QueueCount         int                    `json:"queue_count"`
	TotalSessions      int                    `json:"total_sessions"`
	IsOnline           bool                   `json:"is_online"`
	CurrentUser        *LabCurrentUserInfo    `json:"current_user,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// LabCurrentUserInfo for current user using the lab
type LabCurrentUserInfo struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	AvatarURL        string `json:"avatar_url"`
	SessionRemaining int    `json:"session_remaining"` // seconds
}

// LabDetailResponse includes full hardware specs and queue
type LabDetailResponse struct {
	LabResponse
	HardwareSpecsDetailed *LabHardwareSpecs    `json:"hardware_specs_detailed"`
	Queue                 []LabQueueEntryBrief `json:"queue"`
}

// LabHardwareSpecs for detailed hardware specs
type LabHardwareSpecs struct {
	Board     string         `json:"board"`
	Camera    string         `json:"camera"`
	Sensors   []SensorSpec   `json:"sensors"`
	Actuators []ActuatorSpec `json:"actuators"`
}

// SensorSpec for sensor specification
type SensorSpec struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	Pin  string            `json:"pin,omitempty"`
	Pins map[string]string `json:"pins,omitempty"`
}

// ActuatorSpec for actuator specification
type ActuatorSpec struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	Pin  string            `json:"pin,omitempty"`
	Pins map[string]string `json:"pins,omitempty"`
}

// LabQueueEntryBrief for queue display
type LabQueueEntryBrief struct {
	Position  int    `json:"position"`
	UserName  string `json:"user_name"`
	BidAmount int    `json:"bid_amount"`
}

// ============================================
// Lab Queue DTOs
// ============================================

// JoinQueueRequest for joining lab queue
type JoinQueueRequest struct {
	BidAmount int `json:"bid_amount"` // XP to bid for priority (optional)
}

// JoinQueueResponse for queue join result
type JoinQueueResponse struct {
	QueueID       string    `json:"queue_id"`
	Position      int       `json:"position"`
	EstimatedWait int       `json:"estimated_wait"` // seconds
	ExpiresAt     time.Time `json:"expires_at"`
}

// QueueStatusResponse for queue status
type QueueStatusResponse struct {
	Position      int       `json:"position"`
	EstimatedWait int       `json:"estimated_wait"` // seconds
	ExpiresAt     time.Time `json:"expires_at"`
	BidAmount     int       `json:"bid_amount"`
}

// ============================================
// Lab Session DTOs
// ============================================

// StartSessionResponse for starting a lab session
type StartSessionResponse struct {
	SessionID      string          `json:"session_id"`
	LivekitToken   string          `json:"livekit_token"`
	LivekitURL     string          `json:"livekit_url"`
	RoomName       string          `json:"room_name"`
	ExpiresAt      time.Time       `json:"expires_at"`
	HardwareConfig *HardwareConfig `json:"hardware_config"`
}

// HardwareConfig for hardware configuration
type HardwareConfig struct {
	Board      string `json:"board"`
	SerialPort string `json:"serial_port"`
	BaudRate   int    `json:"baud_rate"`
}

// EndSessionRequest for ending a lab session
type EndSessionRequest struct {
	Feedback string `json:"feedback"`
	Rating   int    `json:"rating" binding:"omitempty,min=1,max=5"`
}

// EndSessionResponse for session end result
type EndSessionResponse struct {
	SessionID       string `json:"session_id"`
	DurationSeconds int    `json:"duration_seconds"`
	XPEarned        int    `json:"xp_earned"`
	Message         string `json:"message"`
}

// SessionResponse for session info
type SessionResponse struct {
	ID                string     `json:"id"`
	LabID             string     `json:"lab_id"`
	LabName           string     `json:"lab_name"`
	Status            string     `json:"status"`
	StartedAt         time.Time  `json:"started_at"`
	EndedAt           *time.Time `json:"ended_at"`
	DurationSeconds   int        `json:"duration_seconds"`
	CompilationStatus string     `json:"compilation_status"`
	XPEarned          int        `json:"xp_earned"`
	Rating            int        `json:"rating"`
}

// ActiveSessionResponse for active session detail
type ActiveSessionResponse struct {
	SessionID        string          `json:"session_id"`
	LabID            string          `json:"lab_id"`
	LabName          string          `json:"lab_name"`
	LivekitToken     string          `json:"livekit_token"`
	LivekitURL       string          `json:"livekit_url"`
	RoomName         string          `json:"room_name"`
	RemainingSeconds int             `json:"remaining_seconds"`
	HardwareConfig   *HardwareConfig `json:"hardware_config"`
}

// ============================================
// Code Execution DTOs
// ============================================

// SubmitCodeRequest for submitting code
type SubmitCodeRequest struct {
	SessionID string `json:"session_id" binding:"required,uuid"`
	Code      string `json:"code" binding:"required"`
	Language  string `json:"language" binding:"required,oneof=arduino micropython c cpp"`
	Filename  string `json:"filename"`
}

// SubmitCodeResponse for code submission result
type SubmitCodeResponse struct {
	CompilationID string `json:"compilation_id"`
	Status        string `json:"status"`
}

// CompilationStatusResponse for compilation status
type CompilationStatusResponse struct {
	Status     string     `json:"status"`
	Output     string     `json:"output"`
	Errors     string     `json:"errors,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
}

// ============================================
// Sensor/Actuator DTOs
// ============================================

// SensorDataResponse for sensor readings
type SensorDataResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Sensors   map[string]interface{} `json:"sensors"`
}

// ControlActuatorRequest for controlling actuators
type ControlActuatorRequest struct {
	Actuator string                 `json:"actuator" binding:"required"`
	Action   string                 `json:"action" binding:"required"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// ControlActuatorResponse for actuator control result
type ControlActuatorResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Actuator string `json:"actuator"`
	Action   string `json:"action"`
}

// SerialCommandRequest for sending serial commands
type SerialCommandRequest struct {
	Command string `json:"command" binding:"required"`
}

// SerialCommandResponse for serial command result
type SerialCommandResponse struct {
	Success  bool   `json:"success"`
	Response string `json:"response"`
}

// ============================================
// Lab Booking DTOs
// ============================================

// CreateBookingRequest for creating a booking
type CreateBookingRequest struct {
	LabID           string    `json:"lab_id" binding:"required,uuid"`
	ScheduledAt     time.Time `json:"scheduled_at" binding:"required"`
	DurationMinutes int       `json:"duration_minutes" binding:"omitempty,min=15,max=60"`
	BidAmount       int       `json:"bid_amount"`
}

// BookingResponse for booking info
type BookingResponse struct {
	ID              string    `json:"id"`
	LabID           string    `json:"lab_id"`
	LabName         string    `json:"lab_name"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	DurationMinutes int       `json:"duration_minutes"`
	Status          string    `json:"status"`
	BidAmount       int       `json:"bid_amount"`
	QueuePosition   int       `json:"queue_position"`
	CreatedAt       time.Time `json:"created_at"`
}

// ============================================
// Lab Hardware Log DTOs
// ============================================

// HardwareLogResponse for hardware log entry
type HardwareLogResponse struct {
	ID        string         `json:"id"`
	EventType string         `json:"event_type"`
	EventData datatypes.JSON `json:"event_data"`
	CreatedAt time.Time      `json:"created_at"`
}

// ============================================
// WebSocket Event DTOs
// ============================================

// WSLabStatusEvent for lab status WebSocket event
type WSLabStatusEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

// WSQueueUpdateEvent for queue update WebSocket event
type WSQueueUpdateEvent struct {
	Event string `json:"event"`
	Data  struct {
		Position      int `json:"position"`
		EstimatedWait int `json:"estimated_wait"`
	} `json:"data"`
}

// WSSensorDataEvent for sensor data WebSocket event
type WSSensorDataEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

// WSCompilationEvent for compilation WebSocket event
type WSCompilationEvent struct {
	Event string `json:"event"`
	Data  struct {
		Status string `json:"status"`
		Output string `json:"output"`
	} `json:"data"`
}

// WSTimeWarningEvent for session time warning WebSocket event
type WSTimeWarningEvent struct {
	Event string `json:"event"`
	Data  struct {
		RemainingSeconds int `json:"remaining_seconds"`
	} `json:"data"`
}

// WSSessionEndedEvent for session ended WebSocket event
type WSSessionEndedEvent struct {
	Event string `json:"event"`
	Data  struct {
		Reason   string `json:"reason"`
		XPEarned int    `json:"xp_earned"`
	} `json:"data"`
}

// ============================================
// MQTT Message DTOs
// ============================================

// MQTTSensorMessage for sensor data from hardware
type MQTTSensorMessage struct {
	LabID     string                 `json:"lab_id"`
	SessionID string                 `json:"session_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Sensors   map[string]interface{} `json:"sensors"`
}

// MQTTActuatorCommand for actuator control to hardware
type MQTTActuatorCommand struct {
	LabID     string                 `json:"lab_id"`
	SessionID string                 `json:"session_id"`
	Actuator  string                 `json:"actuator"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params,omitempty"`
}

// MQTTCodeUploadCommand for code upload to hardware
type MQTTCodeUploadCommand struct {
	LabID         string `json:"lab_id"`
	SessionID     string `json:"session_id"`
	CompilationID string `json:"compilation_id"`
	Code          string `json:"code"`
	Language      string `json:"language"`
	Filename      string `json:"filename"`
}

// MQTTCompilationResult for compilation result from hardware
type MQTTCompilationResult struct {
	LabID         string     `json:"lab_id"`
	SessionID     string     `json:"session_id"`
	CompilationID string     `json:"compilation_id"`
	Status        string     `json:"status"`
	Output        string     `json:"output"`
	Errors        string     `json:"errors,omitempty"`
	UploadedAt    *time.Time `json:"uploaded_at,omitempty"`
}

// MQTTHeartbeat for agent heartbeat
type MQTTHeartbeat struct {
	LabID     string    `json:"lab_id"`
	AgentID   string    `json:"agent_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
	Metrics   struct {
		CPUUsage    float64 `json:"cpu_usage"`
		MemoryUsage float64 `json:"memory_usage"`
		Temperature float64 `json:"temperature"`
	} `json:"metrics,omitempty"`
}

// ============================================
// RabbitMQ Job DTOs
// ============================================

// CodeCompilationJob for compilation queue
type CodeCompilationJob struct {
	CompilationID string    `json:"compilation_id"`
	LabID         string    `json:"lab_id"`
	SessionID     string    `json:"session_id"`
	UserID        string    `json:"user_id"`
	Code          string    `json:"code"`
	Language      string    `json:"language"`
	Filename      string    `json:"filename"`
	CreatedAt     time.Time `json:"created_at"`
}

// SessionExpirationJob for session expiration queue
type SessionExpirationJob struct {
	SessionID string    `json:"session_id"`
	LabID     string    `json:"lab_id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// QueueExpirationJob for queue expiration
type QueueExpirationJob struct {
	QueueID   string    `json:"queue_id"`
	LabID     string    `json:"lab_id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// XPRewardJob for XP reward processing
type XPRewardJob struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Action    string    `json:"action"`
	XPAmount  int       `json:"xp_amount"`
	CreatedAt time.Time `json:"created_at"`
}
