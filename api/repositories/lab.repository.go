package repositories

import (
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// LabRepository handles lab database operations
type LabRepository struct {
	*BaseRepository
}

// NewLabRepository creates a new LabRepository
func NewLabRepository(db *gorm.DB) *LabRepository {
	return &LabRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// LabFilter defines filter options for labs
type LabFilter struct {
	Platform string
	Status   string
	Search   string
	IsOnline *bool
}

// ============================================
// Lab Operations
// ============================================

// FindAll finds all labs with filters and pagination
func (r *LabRepository) FindAll(filter LabFilter, page, limit int) ([]models.Lab, int64, error) {
	var labs []models.Lab
	var total int64

	query := r.DB.Model(&models.Lab{})

	if filter.Platform != "" {
		query = query.Where("platform = ?", filter.Platform)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.IsOnline != nil {
		query = query.Where("is_online = ?", *filter.IsOnline)
	}

	// Count total
	query.Count(&total)

	// Apply pagination and get results
	err := query.Scopes(Paginate(page, limit)).
		Order("status ASC, name ASC").
		Preload("CurrentUser").
		Find(&labs).Error

	return labs, total, err
}

// FindByID finds lab by ID with relations
func (r *LabRepository) FindByID(id string) (*models.Lab, error) {
	var lab models.Lab
	err := r.DB.Preload("CurrentUser").
		Preload("Queue", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC").Preload("User")
		}).
		First(&lab, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lab, nil
}

// FindBySlug finds lab by slug
func (r *LabRepository) FindBySlug(slug string) (*models.Lab, error) {
	var lab models.Lab
	err := r.DB.Preload("CurrentUser").
		Preload("Queue", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC").Preload("User")
		}).
		First(&lab, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &lab, nil
}

// FindByIDSimple finds lab by ID without preloading
func (r *LabRepository) FindByIDSimple(id string) (*models.Lab, error) {
	var lab models.Lab
	err := r.DB.First(&lab, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lab, nil
}

// Create creates a new lab
func (r *LabRepository) Create(lab *models.Lab) error {
	return r.DB.Create(lab).Error
}

// Update updates a lab
func (r *LabRepository) Update(lab *models.Lab) error {
	return r.DB.Save(lab).Error
}

// UpdateStatus updates lab status
func (r *LabRepository) UpdateStatus(labID string, status models.LabStatus) error {
	return r.DB.Model(&models.Lab{}).Where("id = ?", labID).
		Update("status", status).Error
}

// SetCurrentUser sets the current user and session start time
func (r *LabRepository) SetCurrentUser(labID, userID string) error {
	now := time.Now()
	return r.DB.Model(&models.Lab{}).Where("id = ?", labID).
		Updates(map[string]interface{}{
			"current_user_id":    userID,
			"session_started_at": now,
			"status":             models.LabStatusBusy,
		}).Error
}

// ClearCurrentUser clears the current user
func (r *LabRepository) ClearCurrentUser(labID string) error {
	return r.DB.Model(&models.Lab{}).Where("id = ?", labID).
		Updates(map[string]interface{}{
			"current_user_id":    nil,
			"session_started_at": nil,
			"status":             models.LabStatusAvailable,
		}).Error
}

// IncrementTotalSessions increments total sessions count
func (r *LabRepository) IncrementTotalSessions(labID string) error {
	return r.DB.Model(&models.Lab{}).Where("id = ?", labID).
		UpdateColumn("total_sessions", gorm.Expr("total_sessions + 1")).Error
}

// UpdateQueueCount updates queue count
func (r *LabRepository) UpdateQueueCount(labID string) error {
	var count int64
	r.DB.Model(&models.LabQueue{}).Where("lab_id = ?", labID).Count(&count)
	return r.DB.Model(&models.Lab{}).Where("id = ?", labID).
		Update("queue_count", count).Error
}

// UpdateHeartbeat updates lab heartbeat
func (r *LabRepository) UpdateHeartbeat(labID string, isOnline bool) error {
	now := time.Now()
	return r.DB.Model(&models.Lab{}).Where("id = ?", labID).
		Updates(map[string]interface{}{
			"last_heartbeat": now,
			"is_online":      isOnline,
		}).Error
}

// ============================================
// Session Operations
// ============================================

// CreateSession creates a new lab session
func (r *LabRepository) CreateSession(session *models.LabSession) error {
	return r.DB.Create(session).Error
}

// FindSessionByID finds session by ID
func (r *LabRepository) FindSessionByID(id string) (*models.LabSession, error) {
	var session models.LabSession
	err := r.DB.Preload("Lab").Preload("User").First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindActiveSession finds active session for user
func (r *LabRepository) FindActiveSession(userID string) (*models.LabSession, error) {
	var session models.LabSession
	err := r.DB.Preload("Lab").
		Where("user_id = ? AND status = ?", userID, models.SessionStatusActive).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindActiveSessionByLab finds active session for a lab
func (r *LabRepository) FindActiveSessionByLab(labID string) (*models.LabSession, error) {
	var session models.LabSession
	err := r.DB.Preload("User").
		Where("lab_id = ? AND status = ?", labID, models.SessionStatusActive).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession updates a session
func (r *LabRepository) UpdateSession(session *models.LabSession) error {
	return r.DB.Save(session).Error
}

// EndSession ends a session
func (r *LabRepository) EndSession(sessionID string, status models.SessionStatus, feedback string, rating int, xpEarned int) error {
	now := time.Now()
	var session models.LabSession
	if err := r.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}

	duration := int(now.Sub(session.StartedAt).Seconds())

	return r.DB.Model(&models.LabSession{}).Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":           status,
			"ended_at":         now,
			"duration_seconds": duration,
			"feedback":         feedback,
			"rating":           rating,
			"xp_earned":        xpEarned,
		}).Error
}

// GetUserSessionHistory gets user's session history
func (r *LabRepository) GetUserSessionHistory(userID string, page, limit int) ([]models.LabSession, int64, error) {
	var sessions []models.LabSession
	var total int64

	query := r.DB.Model(&models.LabSession{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Order("created_at DESC").
		Preload("Lab").
		Find(&sessions).Error

	return sessions, total, err
}

// ============================================
// Queue Operations
// ============================================

// JoinQueue adds user to lab queue
func (r *LabRepository) JoinQueue(labID, userID string, bidAmount int) (*models.LabQueue, error) {
	// Check if already in queue
	var existing models.LabQueue
	if err := r.DB.Where("lab_id = ? AND user_id = ?", labID, userID).First(&existing).Error; err == nil {
		return nil, gorm.ErrDuplicatedKey
	}

	// Get current max position
	var maxPosition int
	r.DB.Model(&models.LabQueue{}).Where("lab_id = ?", labID).
		Select("COALESCE(MAX(position), 0)").Scan(&maxPosition)

	queue := &models.LabQueue{
		LabID:     labID,
		UserID:    userID,
		Position:  maxPosition + 1,
		BidAmount: bidAmount,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if err := r.DB.Create(queue).Error; err != nil {
		return nil, err
	}

	// Reorder queue based on bids
	r.reorderQueue(labID)

	// Update queue count
	r.UpdateQueueCount(labID)

	// Refresh queue entry
	r.DB.First(queue, "id = ?", queue.ID)

	return queue, nil
}

// reorderQueue reorders queue based on bid amounts
func (r *LabRepository) reorderQueue(labID string) {
	// Get all queue entries ordered by bid (desc) and join time (asc)
	var entries []models.LabQueue
	r.DB.Where("lab_id = ?", labID).
		Order("bid_amount DESC, joined_at ASC").
		Find(&entries)

	// Update positions
	for i, entry := range entries {
		r.DB.Model(&models.LabQueue{}).Where("id = ?", entry.ID).
			Update("position", i+1)
	}
}

// LeaveQueue removes user from queue
func (r *LabRepository) LeaveQueue(labID, userID string) error {
	err := r.DB.Where("lab_id = ? AND user_id = ?", labID, userID).
		Delete(&models.LabQueue{}).Error
	if err != nil {
		return err
	}

	// Reorder remaining queue
	r.reorderQueue(labID)
	r.UpdateQueueCount(labID)

	return nil
}

// GetQueueEntry gets user's queue entry
func (r *LabRepository) GetQueueEntry(labID, userID string) (*models.LabQueue, error) {
	var queue models.LabQueue
	err := r.DB.Where("lab_id = ? AND user_id = ?", labID, userID).First(&queue).Error
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

// GetLabQueue gets all queue entries for a lab
func (r *LabRepository) GetLabQueue(labID string) ([]models.LabQueue, error) {
	var entries []models.LabQueue
	err := r.DB.Where("lab_id = ?", labID).
		Order("position ASC").
		Preload("User").
		Find(&entries).Error
	return entries, err
}

// GetNextInQueue gets the next user in queue
func (r *LabRepository) GetNextInQueue(labID string) (*models.LabQueue, error) {
	var queue models.LabQueue
	err := r.DB.Where("lab_id = ?", labID).
		Order("position ASC").
		Preload("User").
		First(&queue).Error
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

// RemoveFromQueue removes a specific queue entry
func (r *LabRepository) RemoveFromQueue(queueID string) error {
	var queue models.LabQueue
	if err := r.DB.First(&queue, "id = ?", queueID).Error; err != nil {
		return err
	}

	if err := r.DB.Delete(&queue).Error; err != nil {
		return err
	}

	r.reorderQueue(queue.LabID)
	r.UpdateQueueCount(queue.LabID)

	return nil
}

// CleanExpiredQueues removes expired queue entries
func (r *LabRepository) CleanExpiredQueues() error {
	var expired []models.LabQueue
	r.DB.Where("expires_at < ?", time.Now()).Find(&expired)

	labIDs := make(map[string]bool)
	for _, q := range expired {
		labIDs[q.LabID] = true
	}

	if err := r.DB.Where("expires_at < ?", time.Now()).Delete(&models.LabQueue{}).Error; err != nil {
		return err
	}

	// Reorder affected queues
	for labID := range labIDs {
		r.reorderQueue(labID)
		r.UpdateQueueCount(labID)
	}

	return nil
}

// ============================================
// Booking Operations
// ============================================

// CreateBooking creates a new booking
func (r *LabRepository) CreateBooking(booking *models.LabBooking) error {
	return r.DB.Create(booking).Error
}

// FindBookingByID finds booking by ID
func (r *LabRepository) FindBookingByID(id string) (*models.LabBooking, error) {
	var booking models.LabBooking
	err := r.DB.Preload("Lab").Preload("User").First(&booking, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// GetUserBookings gets user's bookings
func (r *LabRepository) GetUserBookings(userID string, page, limit int) ([]models.LabBooking, int64, error) {
	var bookings []models.LabBooking
	var total int64

	query := r.DB.Model(&models.LabBooking{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Order("scheduled_at DESC").
		Preload("Lab").
		Find(&bookings).Error

	return bookings, total, err
}

// UpdateBooking updates a booking
func (r *LabRepository) UpdateBooking(booking *models.LabBooking) error {
	return r.DB.Save(booking).Error
}

// CancelBooking cancels a booking
func (r *LabRepository) CancelBooking(bookingID string) error {
	return r.DB.Model(&models.LabBooking{}).Where("id = ?", bookingID).
		Update("status", models.BookingStatusCancelled).Error
}

// ============================================
// Code Compilation Operations
// ============================================

// CreateCompilation creates a new compilation job
func (r *LabRepository) CreateCompilation(compilation *models.CodeCompilation) error {
	return r.DB.Create(compilation).Error
}

// FindCompilationByID finds compilation by ID
func (r *LabRepository) FindCompilationByID(id string) (*models.CodeCompilation, error) {
	var compilation models.CodeCompilation
	err := r.DB.First(&compilation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &compilation, nil
}

// UpdateCompilation updates a compilation
func (r *LabRepository) UpdateCompilation(compilation *models.CodeCompilation) error {
	return r.DB.Save(compilation).Error
}

// UpdateCompilationStatus updates compilation status
func (r *LabRepository) UpdateCompilationStatus(id string, status models.CompilationStatus, output, errors string) error {
	updates := map[string]interface{}{
		"status": status,
		"output": output,
		"errors": errors,
	}

	if status == models.CompilationStatusCompiling {
		now := time.Now()
		updates["started_at"] = now
	}

	if status == models.CompilationStatusSuccess || status == models.CompilationStatusFailed {
		now := time.Now()
		updates["completed_at"] = now
	}

	if status == models.CompilationStatusSuccess {
		now := time.Now()
		updates["uploaded_at"] = now
	}

	return r.DB.Model(&models.CodeCompilation{}).Where("id = ?", id).Updates(updates).Error
}

// ============================================
// Hardware Log Operations
// ============================================

// CreateHardwareLog creates a new hardware log
func (r *LabRepository) CreateHardwareLog(log *models.LabHardwareLog) error {
	return r.DB.Create(log).Error
}

// GetSessionHardwareLogs gets hardware logs for a session
func (r *LabRepository) GetSessionHardwareLogs(sessionID string, limit int) ([]models.LabHardwareLog, error) {
	var logs []models.LabHardwareLog
	err := r.DB.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetLabHardwareLogs gets hardware logs for a lab
func (r *LabRepository) GetLabHardwareLogs(labID string, limit int) ([]models.LabHardwareLog, error) {
	var logs []models.LabHardwareLog
	err := r.DB.Where("lab_id = ?", labID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
