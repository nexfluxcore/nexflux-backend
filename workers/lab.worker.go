package workers

import (
	"context"
	"encoding/json"
	"log"
	"nexfi-backend/api/repositories"
	"nexfi-backend/models"
	"nexfi-backend/pkg/mqtt"
	"nexfi-backend/pkg/rabbitmq"
	"time"

	"gorm.io/gorm"
)

// LabWorker handles background job processing for labs
type LabWorker struct {
	db       *gorm.DB
	labRepo  *repositories.LabRepository
	userRepo *repositories.UserRepository
	gamRepo  *repositories.GamificationRepository
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewLabWorker creates a new LabWorker
func NewLabWorker(db *gorm.DB) *LabWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &LabWorker{
		db:       db,
		labRepo:  repositories.NewLabRepository(db),
		userRepo: repositories.NewUserRepository(db),
		gamRepo:  repositories.NewGamificationRepository(db),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts all worker goroutines
func (w *LabWorker) Start() {
	log.Println("🏭 Starting lab workers...")

	// Start queue consumers
	go w.startCodeCompilationWorker()
	go w.startSessionExpirationWorker()
	go w.startQueueExpirationWorker()
	go w.startXPRewardWorker()
	go w.startNotificationWorker()
	go w.startHardwareLogWorker()

	// Start MQTT message handlers
	go w.startMQTTHandlers()

	log.Println("✅ Lab workers started")
}

// Stop stops all workers gracefully
func (w *LabWorker) Stop() {
	log.Println("🛑 Stopping lab workers...")
	w.cancel()
}

// ============================================
// Code Compilation Worker
// ============================================

func (w *LabWorker) startCodeCompilationWorker() {
	err := rabbitmq.ConsumeWithContext(w.ctx, rabbitmq.QueueCodeCompilation, func(body []byte) error {
		var job rabbitmq.CodeCompilationJob
		if err := json.Unmarshal(body, &job); err != nil {
			log.Printf("Failed to parse compilation job: %v", err)
			return err
		}

		log.Printf("📦 Processing code compilation: %s (lab: %s)", job.CompilationID, job.LabID)

		// Update compilation status to compiling
		w.labRepo.UpdateCompilationStatus(job.CompilationID, models.CompilationStatusCompiling, "", "")

		// Send code to lab agent via MQTT
		cmd := mqtt.CodeUploadCommand{
			LabID:         job.LabID,
			SessionID:     job.SessionID,
			CompilationID: job.CompilationID,
			Code:          job.Code,
			Language:      job.Language,
			Filename:      job.Filename,
			Timestamp:     time.Now(),
		}

		if err := mqtt.PublishCodeUpload(job.LabID, cmd); err != nil {
			log.Printf("Failed to send code to lab agent: %v", err)
			w.labRepo.UpdateCompilationStatus(job.CompilationID, models.CompilationStatusFailed, "", "Failed to send code to lab agent")
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("Failed to start compilation worker: %v", err)
	}
}

// ============================================
// Session Expiration Worker
// ============================================

func (w *LabWorker) startSessionExpirationWorker() {
	err := rabbitmq.ConsumeWithContext(w.ctx, rabbitmq.QueueSessionExpiration, func(body []byte) error {
		var job rabbitmq.SessionExpirationJob
		if err := json.Unmarshal(body, &job); err != nil {
			log.Printf("Failed to parse session expiration job: %v", err)
			return err
		}

		log.Printf("⏰ Processing session expiration: %s (lab: %s)", job.SessionID, job.LabID)

		// Check if session is still active
		session, err := w.labRepo.FindSessionByID(job.SessionID)
		if err != nil || session.Status != models.SessionStatusActive {
			// Session already ended, skip
			return nil
		}

		// End the session
		w.labRepo.EndSession(job.SessionID, models.SessionStatusExpired, "", 0, 50) // Base XP for expired session

		// Clear lab current user
		w.labRepo.ClearCurrentUser(job.LabID)

		// Notify user
		rabbitmq.PublishNotification(rabbitmq.NotificationJob{
			UserID:  job.UserID,
			Type:    "session_expired",
			Title:   "Lab Session Expired",
			Message: "Your lab session has expired due to time limit.",
			Data: map[string]interface{}{
				"lab_id":     job.LabID,
				"session_id": job.SessionID,
			},
		})

		// Notify lab agent via MQTT
		mqtt.PublishLabControl(job.LabID, mqtt.LabControlCommand{
			LabID:     job.LabID,
			SessionID: job.SessionID,
			Command:   "end_session",
		})

		// Process next in queue
		w.processNextInQueue(job.LabID)

		return nil
	})

	if err != nil {
		log.Printf("Failed to start session expiration worker: %v", err)
	}
}

// ============================================
// Queue Expiration Worker
// ============================================

func (w *LabWorker) startQueueExpirationWorker() {
	err := rabbitmq.ConsumeWithContext(w.ctx, rabbitmq.QueueQueueExpiration, func(body []byte) error {
		var job rabbitmq.QueueExpirationJob
		if err := json.Unmarshal(body, &job); err != nil {
			log.Printf("Failed to parse queue expiration job: %v", err)
			return err
		}

		log.Printf("⏰ Processing queue expiration: %s (lab: %s)", job.QueueID, job.LabID)

		// Remove from queue
		w.labRepo.RemoveFromQueue(job.QueueID)

		// Notify user
		rabbitmq.PublishNotification(rabbitmq.NotificationJob{
			UserID:  job.UserID,
			Type:    "queue_expired",
			Title:   "Queue Position Expired",
			Message: "Your position in the lab queue has expired.",
			Data: map[string]interface{}{
				"lab_id":   job.LabID,
				"queue_id": job.QueueID,
			},
		})

		return nil
	})

	if err != nil {
		log.Printf("Failed to start queue expiration worker: %v", err)
	}
}

// ============================================
// XP Reward Worker
// ============================================

func (w *LabWorker) startXPRewardWorker() {
	err := rabbitmq.ConsumeWithContext(w.ctx, rabbitmq.QueueXPReward, func(body []byte) error {
		var job rabbitmq.XPRewardJob
		if err := json.Unmarshal(body, &job); err != nil {
			log.Printf("Failed to parse XP reward job: %v", err)
			return err
		}

		log.Printf("🎯 Processing XP reward: %s (%d XP, action: %s)", job.UserID, job.XPAmount, job.Action)

		// Get user
		user, err := w.userRepo.FindByID(job.UserID)
		if err != nil {
			log.Printf("User not found: %v", err)
			return nil // Don't retry
		}

		// Add XP
		user.TotalXP += job.XPAmount
		user.CurrentXP += job.XPAmount

		// Check for level up
		if user.CurrentXP >= user.TargetXP {
			user.Level++
			user.CurrentXP -= user.TargetXP
			user.TargetXP = calculateNextLevelXP(user.Level)

			// Send level up notification
			rabbitmq.PublishNotification(rabbitmq.NotificationJob{
				UserID:  job.UserID,
				Type:    "level_up",
				Title:   "🎉 Level Up!",
				Message: "Congratulations! You've reached level " + string(rune(user.Level+'0')) + "!",
				Data: map[string]interface{}{
					"new_level": user.Level,
				},
			})
		}

		// Save user
		w.userRepo.Update(user)

		// Update streak
		w.gamRepo.UpdateStreak(job.UserID)

		return nil
	})

	if err != nil {
		log.Printf("Failed to start XP reward worker: %v", err)
	}
}

// calculateNextLevelXP calculates XP required for next level
func calculateNextLevelXP(level int) int {
	return 1000 + (level * 500)
}

// ============================================
// Notification Worker
// ============================================

func (w *LabWorker) startNotificationWorker() {
	err := rabbitmq.ConsumeWithContext(w.ctx, rabbitmq.QueueNotification, func(body []byte) error {
		var job rabbitmq.NotificationJob
		if err := json.Unmarshal(body, &job); err != nil {
			log.Printf("Failed to parse notification job: %v", err)
			return err
		}

		log.Printf("📬 Processing notification: %s (type: %s)", job.UserID, job.Type)

		// Create notification in database
		data, _ := json.Marshal(job.Data)
		notification := &models.Notification{
			UserID:  job.UserID,
			Type:    models.NotificationType(job.Type),
			Title:   job.Title,
			Message: job.Message,
			Data:    data,
		}

		w.db.Create(notification)

		// TODO: Send push notification if user has push enabled

		return nil
	})

	if err != nil {
		log.Printf("Failed to start notification worker: %v", err)
	}
}

// ============================================
// Hardware Log Worker
// ============================================

func (w *LabWorker) startHardwareLogWorker() {
	err := rabbitmq.ConsumeWithContext(w.ctx, rabbitmq.QueueHardwareLog, func(body []byte) error {
		var job rabbitmq.HardwareLogJob
		if err := json.Unmarshal(body, &job); err != nil {
			log.Printf("Failed to parse hardware log job: %v", err)
			return err
		}

		// Create hardware log in database
		data, _ := json.Marshal(job.EventData)
		logEntry := &models.LabHardwareLog{
			LabID:     job.LabID,
			EventType: models.LabEventType(job.EventType),
			EventData: data,
		}

		if job.SessionID != "" {
			logEntry.SessionID = &job.SessionID
		}

		w.labRepo.CreateHardwareLog(logEntry)

		return nil
	})

	if err != nil {
		log.Printf("Failed to start hardware log worker: %v", err)
	}
}

// ============================================
// MQTT Message Handlers
// ============================================

func (w *LabWorker) startMQTTHandlers() {
	if !mqtt.IsConnected() {
		log.Println("⚠️ MQTT not connected, skipping MQTT handlers")
		return
	}

	// Handle compilation results from lab agents
	mqtt.Subscribe("lab/+/compilation", 1, func(topic string, payload []byte) {
		var result mqtt.CompilationResult
		if err := json.Unmarshal(payload, &result); err != nil {
			log.Printf("Failed to parse compilation result: %v", err)
			return
		}

		log.Printf("📥 Received compilation result: %s (status: %s)", result.CompilationID, result.Status)

		// Update compilation status
		status := models.CompilationStatus(result.Status)
		w.labRepo.UpdateCompilationStatus(result.CompilationID, status, result.Output, result.Errors)

		// If successful, award XP
		if status == models.CompilationStatusSuccess {
			session, err := w.labRepo.FindSessionByID(result.SessionID)
			if err == nil {
				rabbitmq.PublishXPReward(rabbitmq.XPRewardJob{
					UserID:    session.UserID,
					SessionID: result.SessionID,
					Action:    "code_upload_success",
					XPAmount:  20,
				})
			}
		}
	})

	// Handle heartbeats from lab agents
	mqtt.Subscribe("lab/+/heartbeat", 1, func(topic string, payload []byte) {
		var heartbeat mqtt.HeartbeatMessage
		if err := json.Unmarshal(payload, &heartbeat); err != nil {
			return
		}

		// Update lab heartbeat
		isOnline := heartbeat.Status == "online"
		w.labRepo.UpdateHeartbeat(heartbeat.LabID, isOnline)
	})

	// Handle serial responses
	mqtt.Subscribe("lab/+/serial/response", 1, func(topic string, payload []byte) {
		// Forward to WebSocket clients
		// This is handled by the WebSocket handler
	})

	log.Println("📡 MQTT handlers started")
}

// ============================================
// Helper Functions
// ============================================

func (w *LabWorker) processNextInQueue(labID string) {
	nextInQueue, err := w.labRepo.GetNextInQueue(labID)
	if err != nil || nextInQueue == nil {
		return
	}

	// Send notification to next user
	rabbitmq.PublishNotification(rabbitmq.NotificationJob{
		UserID:  nextInQueue.UserID,
		Type:    "lab_available",
		Title:   "🎉 Lab Available!",
		Message: "It's your turn! The lab is now available for you.",
		Data: map[string]interface{}{
			"lab_id": labID,
		},
	})
}
