package repositories

import (
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// SecurityRepository handles security-related database operations
type SecurityRepository struct {
	*BaseRepository
}

// NewSecurityRepository creates a new SecurityRepository
func NewSecurityRepository(db *gorm.DB) *SecurityRepository {
	return &SecurityRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// ============================================
// Session Management
// ============================================

// CreateSession creates a new session
func (r *SecurityRepository) CreateSession(session *models.UserSession) error {
	return r.DB.Create(session).Error
}

// FindSessionByID finds session by ID
func (r *SecurityRepository) FindSessionByID(id string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.DB.First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindSessionByToken finds session by token hash
func (r *SecurityRepository) FindSessionByToken(tokenHash string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.DB.Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserSessions gets all active sessions for a user
func (r *SecurityRepository) GetUserSessions(userID string) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := r.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("last_active_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// UpdateSessionActivity updates last active time
func (r *SecurityRepository) UpdateSessionActivity(sessionID string) error {
	return r.DB.Model(&models.UserSession{}).
		Where("id = ?", sessionID).
		Update("last_active_at", time.Now()).Error
}

// RevokeSession deletes a session
func (r *SecurityRepository) RevokeSession(sessionID, userID string) error {
	return r.DB.Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&models.UserSession{}).Error
}

// RevokeAllSessionsExcept revokes all sessions except the specified one
func (r *SecurityRepository) RevokeAllSessionsExcept(userID, exceptSessionID string) (int64, error) {
	result := r.DB.Where("user_id = ? AND id != ?", userID, exceptSessionID).
		Delete(&models.UserSession{})
	return result.RowsAffected, result.Error
}

// RevokeAllSessions revokes all sessions for a user
func (r *SecurityRepository) RevokeAllSessions(userID string) error {
	return r.DB.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}

// CountUserSessions counts active sessions for a user
func (r *SecurityRepository) CountUserSessions(userID string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.UserSession{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}

// CleanupExpiredSessions deletes expired sessions
func (r *SecurityRepository) CleanupExpiredSessions() error {
	return r.DB.Where("expires_at < ?", time.Now()).Delete(&models.UserSession{}).Error
}

// ============================================
// Login History
// ============================================

// CreateLoginHistory creates a login history entry
func (r *SecurityRepository) CreateLoginHistory(history *models.LoginHistory) error {
	return r.DB.Create(history).Error
}

// GetLoginHistory gets login history for a user
func (r *SecurityRepository) GetLoginHistory(userID string, page, limit int) ([]models.LoginHistory, int64, error) {
	var history []models.LoginHistory
	var total int64

	query := r.DB.Model(&models.LoginHistory{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Order("created_at DESC").
		Find(&history).Error

	return history, total, err
}

// GetRecentFailedAttempts counts recent failed login attempts
func (r *SecurityRepository) GetRecentFailedAttempts(email string, since time.Time) (int64, error) {
	var count int64
	err := r.DB.Model(&models.LoginHistory{}).
		Where("email = ? AND status = ? AND created_at > ?", email, models.LoginStatusFailed, since).
		Count(&count).Error
	return count, err
}

// GetLastSuccessfulLogin gets the last successful login
func (r *SecurityRepository) GetLastSuccessfulLogin(userID string) (*models.LoginHistory, error) {
	var history models.LoginHistory
	err := r.DB.Where("user_id = ? AND status = ?", userID, models.LoginStatusSuccess).
		Order("created_at DESC").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ============================================
// Two-Factor Authentication
// ============================================

// CreateUser2FA creates 2FA settings for a user
func (r *SecurityRepository) CreateUser2FA(twoFA *models.User2FA) error {
	return r.DB.Create(twoFA).Error
}

// GetUser2FA gets 2FA settings for a user
func (r *SecurityRepository) GetUser2FA(userID string) (*models.User2FA, error) {
	var twoFA models.User2FA
	err := r.DB.Where("user_id = ?", userID).First(&twoFA).Error
	if err != nil {
		return nil, err
	}
	return &twoFA, nil
}

// UpdateUser2FA updates 2FA settings
func (r *SecurityRepository) UpdateUser2FA(twoFA *models.User2FA) error {
	return r.DB.Save(twoFA).Error
}

// DeleteUser2FA deletes 2FA settings
func (r *SecurityRepository) DeleteUser2FA(userID string) error {
	return r.DB.Where("user_id = ?", userID).Delete(&models.User2FA{}).Error
}

// ============================================
// Password History
// ============================================

// CreatePasswordHistory adds password to history
func (r *SecurityRepository) CreatePasswordHistory(history *models.PasswordHistory) error {
	return r.DB.Create(history).Error
}

// GetPasswordHistory gets recent password history for a user
func (r *SecurityRepository) GetPasswordHistory(userID string, limit int) ([]models.PasswordHistory, error) {
	var history []models.PasswordHistory
	err := r.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// ============================================
// Security Logs
// ============================================

// CreateSecurityLog creates a security log entry
func (r *SecurityRepository) CreateSecurityLog(log *models.SecurityLog) error {
	return r.DB.Create(log).Error
}

// GetSecurityLogs gets security logs for a user
func (r *SecurityRepository) GetSecurityLogs(userID string, page, limit int) ([]models.SecurityLog, int64, error) {
	var logs []models.SecurityLog
	var total int64

	query := r.DB.Model(&models.SecurityLog{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, total, err
}
