package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"nexfi-backend/pkg/livekit"
	"nexfi-backend/pkg/mqtt"
	"nexfi-backend/pkg/rabbitmq"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LabService handles lab business logic
type LabService struct {
	repo     *repositories.LabRepository
	userRepo *repositories.UserRepository
	gamRepo  *repositories.GamificationRepository
}

// NewLabService creates a new LabService
func NewLabService(db *gorm.DB) *LabService {
	return &LabService{
		repo:     repositories.NewLabRepository(db),
		userRepo: repositories.NewUserRepository(db),
		gamRepo:  repositories.NewGamificationRepository(db),
	}
}

// ============================================
// Lab Management
// ============================================

// ListLabs lists available labs with filters
func (s *LabService) ListLabs(req dto.LabListRequest) ([]dto.LabResponse, dto.PaginationResponse, error) {
	filter := repositories.LabFilter{
		Platform: req.Platform,
		Status:   req.Status,
		Search:   req.Search,
	}

	labs, total, err := s.repo.FindAll(filter, req.Page, req.Limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.LabResponse, len(labs))
	for i, lab := range labs {
		responses[i] = s.toLabResponse(&lab)
	}

	pagination := dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      int(total),
		TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
	}

	return responses, pagination, nil
}

// GetLab gets lab details by ID
func (s *LabService) GetLab(labID string) (*dto.LabDetailResponse, error) {
	lab, err := s.repo.FindByID(labID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lab not found")
		}
		return nil, err
	}

	return s.toLabDetailResponse(lab), nil
}

// GetLabBySlug gets lab details by slug
func (s *LabService) GetLabBySlug(slug string) (*dto.LabDetailResponse, error) {
	lab, err := s.repo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lab not found")
		}
		return nil, err
	}

	return s.toLabDetailResponse(lab), nil
}

// ============================================
// Queue Management
// ============================================

// JoinQueue adds user to lab queue
func (s *LabService) JoinQueue(labID, userID string, req dto.JoinQueueRequest) (*dto.JoinQueueResponse, error) {
	// Get lab
	lab, err := s.repo.FindByIDSimple(labID)
	if err != nil {
		return nil, errors.New("lab not found")
	}

	// Check if lab is available or busy (can queue for both)
	if lab.Status == models.LabStatusMaintenance || lab.Status == models.LabStatusOffline {
		return nil, errors.New("lab is not accepting queue entries")
	}

	// Check if user already has an active session
	activeSession, _ := s.repo.FindActiveSession(userID)
	if activeSession != nil {
		return nil, errors.New("you already have an active lab session")
	}

	// Check if user has enough XP for bid
	if req.BidAmount > 0 {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return nil, errors.New("user not found")
		}
		if user.CurrentXP < req.BidAmount {
			return nil, errors.New("insufficient XP for bid amount")
		}
	}

	// Join queue
	queueEntry, err := s.repo.JoinQueue(labID, userID, req.BidAmount)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("already in queue for this lab")
		}
		return nil, err
	}

	// Publish queue expiration job
	rabbitmq.PublishQueueExpiration(rabbitmq.QueueExpirationJob{
		QueueID:   queueEntry.ID,
		LabID:     labID,
		UserID:    userID,
		ExpiresAt: queueEntry.ExpiresAt,
	})

	// Calculate estimated wait time
	estimatedWait := s.calculateEstimatedWait(lab, queueEntry.Position)

	return &dto.JoinQueueResponse{
		QueueID:       queueEntry.ID,
		Position:      queueEntry.Position,
		EstimatedWait: estimatedWait,
		ExpiresAt:     queueEntry.ExpiresAt,
	}, nil
}

// LeaveQueue removes user from lab queue
func (s *LabService) LeaveQueue(labID, userID string) error {
	err := s.repo.LeaveQueue(labID, userID)
	if err != nil {
		return errors.New("not in queue or error leaving queue")
	}
	return nil
}

// GetQueueStatus gets user's queue status
func (s *LabService) GetQueueStatus(labID, userID string) (*dto.QueueStatusResponse, error) {
	queueEntry, err := s.repo.GetQueueEntry(labID, userID)
	if err != nil {
		return nil, errors.New("not in queue for this lab")
	}

	lab, _ := s.repo.FindByIDSimple(labID)
	estimatedWait := s.calculateEstimatedWait(lab, queueEntry.Position)

	return &dto.QueueStatusResponse{
		Position:      queueEntry.Position,
		EstimatedWait: estimatedWait,
		ExpiresAt:     queueEntry.ExpiresAt,
		BidAmount:     queueEntry.BidAmount,
	}, nil
}

// calculateEstimatedWait calculates estimated wait time in seconds
func (s *LabService) calculateEstimatedWait(lab *models.Lab, position int) int {
	if position == 1 && lab.Status == models.LabStatusAvailable {
		return 0
	}

	// Average session duration * position in queue
	avgSessionDuration := lab.MaxSessionDuration * 60 // minutes to seconds

	if lab.Status == models.LabStatusBusy && lab.SessionStartedAt != nil {
		// Add remaining time of current session
		remaining := lab.SessionRemainingSeconds()
		return remaining + ((position - 1) * avgSessionDuration)
	}

	return position * avgSessionDuration
}

// ============================================
// Session Management
// ============================================

// StartSession starts a lab session for a user
func (s *LabService) StartSession(labID, userID string) (*dto.StartSessionResponse, error) {
	// Get lab
	lab, err := s.repo.FindByID(labID)
	if err != nil {
		return nil, errors.New("lab not found")
	}

	// Check if lab is available or user is first in queue
	if lab.Status != models.LabStatusAvailable {
		// Check if user is first in queue
		nextInQueue, _ := s.repo.GetNextInQueue(labID)
		if nextInQueue == nil || nextInQueue.UserID != userID {
			return nil, errors.New("lab is not available, please join the queue")
		}
	}

	// Check if user already has active session
	activeSession, _ := s.repo.FindActiveSession(userID)
	if activeSession != nil {
		return nil, errors.New("you already have an active lab session")
	}

	// Get user info
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Remove from queue if present
	s.repo.LeaveQueue(labID, userID)

	// Create LiveKit session
	sessionID := generateUUID()
	livekitInfo, err := livekit.CreateLabSession(lab.Slug, sessionID, userID, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create video session: %v", err)
	}

	// Create session record
	session := &models.LabSession{
		ID:           sessionID,
		LabID:        labID,
		UserID:       userID,
		Status:       models.SessionStatusActive,
		LivekitToken: livekitInfo.Token,
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}

	// Update lab status
	s.repo.SetCurrentUser(labID, userID)
	s.repo.IncrementTotalSessions(labID)

	// Calculate expiration
	sessionDuration := getSessionDuration()
	expiresAt := time.Now().Add(sessionDuration)

	// Publish session expiration job to RabbitMQ
	rabbitmq.PublishSessionExpiration(rabbitmq.SessionExpirationJob{
		SessionID: sessionID,
		LabID:     labID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	})

	// Notify lab agent via MQTT
	mqtt.PublishLabControl(labID, mqtt.LabControlCommand{
		LabID:     labID,
		SessionID: sessionID,
		Command:   "start_session",
		UserID:    userID,
	})

	// Log hardware event
	s.logHardwareEvent(labID, &sessionID, models.EventTypeSessionStart, map[string]interface{}{
		"user_id":    userID,
		"session_id": sessionID,
	})

	return &dto.StartSessionResponse{
		SessionID:    sessionID,
		LivekitToken: livekitInfo.Token,
		LivekitURL:   livekitInfo.LivekitURL,
		RoomName:     livekitInfo.RoomName,
		ExpiresAt:    expiresAt,
		HardwareConfig: &dto.HardwareConfig{
			Board:      string(lab.Platform),
			SerialPort: "/dev/ttyUSB0",
			BaudRate:   9600,
		},
	}, nil
}

// EndSession ends a lab session
func (s *LabService) EndSession(labID, userID string, req dto.EndSessionRequest) (*dto.EndSessionResponse, error) {
	// Get active session
	session, err := s.repo.FindActiveSessionByLab(labID)
	if err != nil {
		return nil, errors.New("no active session found")
	}

	// Verify user owns the session
	if session.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Calculate XP earned
	xpEarned := s.calculateSessionXP(session, req.Rating)

	// End session
	if err := s.repo.EndSession(session.ID, models.SessionStatusCompleted, req.Feedback, req.Rating, xpEarned); err != nil {
		return nil, err
	}

	// Clear lab current user
	s.repo.ClearCurrentUser(labID)

	// End LiveKit session
	lab, _ := s.repo.FindByIDSimple(labID)
	if lab != nil {
		roomName := livekit.BuildLabRoomName(lab.Slug, session.ID)
		livekit.EndLabSession(roomName)
	}

	// Notify lab agent via MQTT
	mqtt.PublishLabControl(labID, mqtt.LabControlCommand{
		LabID:     labID,
		SessionID: session.ID,
		Command:   "end_session",
	})

	// Award XP via RabbitMQ
	rabbitmq.PublishXPReward(rabbitmq.XPRewardJob{
		UserID:    userID,
		SessionID: session.ID,
		Action:    "lab_session_complete",
		XPAmount:  xpEarned,
	})

	// Log hardware event
	s.logHardwareEvent(labID, &session.ID, models.EventTypeSessionEnd, map[string]interface{}{
		"user_id":   userID,
		"xp_earned": xpEarned,
		"rating":    req.Rating,
	})

	// Check if there's someone in queue and notify them
	s.processNextInQueue(labID)

	duration := int(time.Since(session.StartedAt).Seconds())

	return &dto.EndSessionResponse{
		SessionID:       session.ID,
		DurationSeconds: duration,
		XPEarned:        xpEarned,
		Message:         "Session completed successfully!",
	}, nil
}

// GetActiveSession gets user's active session
func (s *LabService) GetActiveSession(userID string) (*dto.ActiveSessionResponse, error) {
	session, err := s.repo.FindActiveSession(userID)
	if err != nil {
		return nil, errors.New("no active session found")
	}

	lab, _ := s.repo.FindByIDSimple(session.LabID)

	return &dto.ActiveSessionResponse{
		SessionID:        session.ID,
		LabID:            session.LabID,
		LabName:          lab.Name,
		LivekitToken:     session.LivekitToken,
		LivekitURL:       livekit.GetHost(),
		RoomName:         livekit.BuildLabRoomName(lab.Slug, session.ID),
		RemainingSeconds: lab.SessionRemainingSeconds(),
		HardwareConfig: &dto.HardwareConfig{
			Board:      string(lab.Platform),
			SerialPort: "/dev/ttyUSB0",
			BaudRate:   9600,
		},
	}, nil
}

// GetUserSessionHistory gets user's session history
func (s *LabService) GetUserSessionHistory(userID string, page, limit int) ([]dto.SessionResponse, dto.PaginationResponse, error) {
	sessions, total, err := s.repo.GetUserSessionHistory(userID, page, limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		labName := ""
		if session.Lab != nil {
			labName = session.Lab.Name
		}
		responses[i] = dto.SessionResponse{
			ID:                session.ID,
			LabID:             session.LabID,
			LabName:           labName,
			Status:            string(session.Status),
			StartedAt:         session.StartedAt,
			EndedAt:           session.EndedAt,
			DurationSeconds:   session.DurationSeconds,
			CompilationStatus: string(session.CompilationStatus),
			XPEarned:          session.XPEarned,
			Rating:            session.Rating,
		}
	}

	pagination := dto.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	return responses, pagination, nil
}

// calculateSessionXP calculates XP earned from a session
func (s *LabService) calculateSessionXP(session *models.LabSession, rating int) int {
	baseXP := 50 // Base XP for completing a session

	// Bonus for successful compilation
	if session.CompilationStatus == models.CompilationStatusSuccess {
		baseXP += 20
	}

	// Bonus for rating
	if rating == 5 {
		baseXP += 25
	}

	return baseXP
}

// processNextInQueue notifies next user in queue
func (s *LabService) processNextInQueue(labID string) {
	nextInQueue, err := s.repo.GetNextInQueue(labID)
	if err != nil || nextInQueue == nil {
		return
	}

	// Send notification
	rabbitmq.PublishNotification(rabbitmq.NotificationJob{
		UserID:  nextInQueue.UserID,
		Type:    "lab_available",
		Title:   "Lab Available!",
		Message: "It's your turn! The lab is now available for you.",
		Data: map[string]interface{}{
			"lab_id": labID,
		},
	})
}

// ============================================
// Code Execution
// ============================================

// SubmitCode submits code for compilation and upload
func (s *LabService) SubmitCode(labID, userID string, req dto.SubmitCodeRequest) (*dto.SubmitCodeResponse, error) {
	var session *models.LabSession
	var err error

	// If SessionID is provided, use it; otherwise get user's active session
	if req.SessionID != "" {
		session, err = s.repo.FindSessionByID(req.SessionID)
		if err != nil {
			return nil, errors.New("session not found")
		}
	} else {
		// Get user's active session for this lab
		session, err = s.repo.FindActiveSessionByLab(labID)
		if err != nil {
			return nil, errors.New("no active session found for this lab")
		}
		// Use the found session ID
		req.SessionID = session.ID
	}

	if session.UserID != userID {
		return nil, errors.New("access denied")
	}

	if session.Status != models.SessionStatusActive {
		return nil, errors.New("session is not active")
	}

	// Create compilation record
	compilation := &models.CodeCompilation{
		SessionID: req.SessionID,
		LabID:     labID,
		UserID:    userID,
		Code:      req.Code,
		Language:  req.Language,
		Filename:  req.Filename,
		Status:    models.CompilationStatusPending,
	}

	if err := s.repo.CreateCompilation(compilation); err != nil {
		return nil, err
	}

	// Update session with code
	session.CodeSubmitted = req.Code
	session.CompilationStatus = models.CompilationStatusPending
	s.repo.UpdateSession(session)

	// Publish compilation job to RabbitMQ
	rabbitmq.PublishCodeCompilation(rabbitmq.CodeCompilationJob{
		CompilationID: compilation.ID,
		LabID:         labID,
		SessionID:     req.SessionID,
		UserID:        userID,
		Code:          req.Code,
		Language:      req.Language,
		Filename:      req.Filename,
	})

	// Also send via MQTT for immediate processing by lab agent
	mqtt.PublishCodeUpload(labID, mqtt.CodeUploadCommand{
		LabID:         labID,
		SessionID:     req.SessionID,
		CompilationID: compilation.ID,
		Code:          req.Code,
		Language:      req.Language,
		Filename:      req.Filename,
		Timestamp:     time.Now(),
	})

	return &dto.SubmitCodeResponse{
		CompilationID: compilation.ID,
		Status:        "pending",
	}, nil
}

// GetCompilationStatus gets compilation status
func (s *LabService) GetCompilationStatus(labID, compilationID, userID string) (*dto.CompilationStatusResponse, error) {
	compilation, err := s.repo.FindCompilationByID(compilationID)
	if err != nil {
		return nil, errors.New("compilation not found")
	}

	if compilation.UserID != userID {
		return nil, errors.New("access denied")
	}

	return &dto.CompilationStatusResponse{
		Status:     string(compilation.Status),
		Output:     compilation.Output,
		Errors:     compilation.Errors,
		UploadedAt: compilation.UploadedAt,
	}, nil
}

// ============================================
// Sensor/Actuator Control
// ============================================

// GetSensorData gets current sensor data from a lab
func (s *LabService) GetSensorData(labID, userID string) (*dto.SensorDataResponse, error) {
	// Verify user has active session on this lab
	session, err := s.repo.FindActiveSessionByLab(labID)
	if err != nil {
		return nil, errors.New("no active session on this lab")
	}

	if session.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Request sensor data via MQTT (synchronous request)
	// In production, this would use a request/response pattern
	// For now, return cached/mock data
	sensorData := map[string]interface{}{
		"DHT22": map[string]interface{}{
			"temperature": 25.5,
			"humidity":    60.0,
		},
		"LDR": map[string]interface{}{
			"light_level": 512,
		},
		"HC-SR04": map[string]interface{}{
			"distance_cm": 45.2,
		},
	}

	return &dto.SensorDataResponse{
		Timestamp: time.Now(),
		Sensors:   sensorData,
	}, nil
}

// ControlActuator sends actuator control command
func (s *LabService) ControlActuator(labID, userID string, req dto.ControlActuatorRequest) (*dto.ControlActuatorResponse, error) {
	// Verify user has active session on this lab
	session, err := s.repo.FindActiveSessionByLab(labID)
	if err != nil {
		return nil, errors.New("no active session on this lab")
	}

	if session.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Send actuator command via MQTT
	cmd := mqtt.ActuatorCommand{
		LabID:     labID,
		SessionID: session.ID,
		Actuator:  req.Actuator,
		Action:    req.Action,
		Params:    req.Params,
		Timestamp: time.Now(),
	}

	if err := mqtt.PublishActuatorControl(labID, cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	// Log hardware event
	s.logHardwareEvent(labID, &session.ID, models.EventTypeActuatorControl, map[string]interface{}{
		"actuator": req.Actuator,
		"action":   req.Action,
		"params":   req.Params,
	})

	return &dto.ControlActuatorResponse{
		Success:  true,
		Message:  "Command sent successfully",
		Actuator: req.Actuator,
		Action:   req.Action,
	}, nil
}

// SendSerialCommand sends a serial command to the hardware
func (s *LabService) SendSerialCommand(labID, userID string, req dto.SerialCommandRequest) (*dto.SerialCommandResponse, error) {
	// Verify user has active session on this lab
	session, err := s.repo.FindActiveSessionByLab(labID)
	if err != nil {
		return nil, errors.New("no active session on this lab")
	}

	if session.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Send serial command via MQTT
	cmd := mqtt.SerialCommand{
		LabID:     labID,
		SessionID: session.ID,
		Command:   req.Command,
		Timestamp: time.Now(),
	}

	if err := mqtt.PublishSerialCommand(labID, cmd); err != nil {
		return nil, fmt.Errorf("failed to send serial command: %v", err)
	}

	return &dto.SerialCommandResponse{
		Success:  true,
		Response: "Command sent, awaiting response...",
	}, nil
}

// ============================================
// Helper Functions
// ============================================

// logHardwareEvent logs a hardware event
func (s *LabService) logHardwareEvent(labID string, sessionID *string, eventType models.LabEventType, data map[string]interface{}) {
	eventData, _ := json.Marshal(data)
	log := &models.LabHardwareLog{
		LabID:     labID,
		SessionID: sessionID,
		EventType: eventType,
		EventData: eventData,
	}
	s.repo.CreateHardwareLog(log)
}

// toLabResponse converts Lab model to LabResponse DTO
func (s *LabService) toLabResponse(lab *models.Lab) dto.LabResponse {
	response := dto.LabResponse{
		ID:                 lab.ID,
		Name:               lab.Name,
		Slug:               lab.Slug,
		Description:        lab.Description,
		Platform:           string(lab.Platform),
		Status:             string(lab.Status),
		ThumbnailURL:       lab.ThumbnailURL,
		MaxSessionDuration: lab.MaxSessionDuration,
		QueueCount:         lab.QueueCount,
		TotalSessions:      lab.TotalSessions,
		IsOnline:           lab.IsOnline,
		CreatedAt:          lab.CreatedAt,
		UpdatedAt:          lab.UpdatedAt,
	}

	// Parse hardware specs
	if lab.HardwareSpecs != nil {
		var specs map[string]interface{}
		json.Unmarshal(lab.HardwareSpecs, &specs)
		response.HardwareSpecs = specs
	}

	// Add current user info if lab is busy
	if lab.CurrentUser != nil {
		response.CurrentUser = &dto.LabCurrentUserInfo{
			ID:               lab.CurrentUser.ID,
			Name:             lab.CurrentUser.Name,
			AvatarURL:        lab.CurrentUser.AvatarURL,
			SessionRemaining: lab.SessionRemainingSeconds(),
		}
	}

	return response
}

// toLabDetailResponse converts Lab model to LabDetailResponse DTO
func (s *LabService) toLabDetailResponse(lab *models.Lab) *dto.LabDetailResponse {
	base := s.toLabResponse(lab)

	detail := &dto.LabDetailResponse{
		LabResponse: base,
	}

	// Parse detailed hardware specs
	if lab.HardwareSpecs != nil {
		var specs dto.LabHardwareSpecs
		json.Unmarshal(lab.HardwareSpecs, &specs)
		detail.HardwareSpecsDetailed = &specs
	}

	// Add queue info
	queue := make([]dto.LabQueueEntryBrief, 0)
	for _, q := range lab.Queue {
		userName := ""
		if q.User != nil {
			userName = q.User.Name
		}
		queue = append(queue, dto.LabQueueEntryBrief{
			Position:  q.Position,
			UserName:  userName,
			BidAmount: q.BidAmount,
		})
	}
	detail.Queue = queue

	return detail
}

// getSessionDuration returns session duration from env or default
func getSessionDuration() time.Duration {
	durationStr := os.Getenv("LAB_SESSION_DURATION")
	if durationStr == "" {
		return 30 * time.Minute // Default 30 minutes
	}

	seconds, err := strconv.Atoi(durationStr)
	if err != nil {
		return 30 * time.Minute
	}

	return time.Duration(seconds) * time.Second
}

// generateUUID generates a proper UUID v4
func generateUUID() string {
	return uuid.New().String()
}
