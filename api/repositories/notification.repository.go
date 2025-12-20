package repositories

import (
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// NotificationRepository handles notification database operations
type NotificationRepository struct {
	*BaseRepository
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// FindByUser finds notifications for a user
func (r *NotificationRepository) FindByUser(userID string, notifType string, unreadOnly bool, page, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.DB.Model(&models.Notification{}).Where("user_id = ?", userID)

	if notifType != "" {
		query = query.Where("type = ?", notifType)
	}

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Order("created_at DESC").
		Find(&notifications).Error

	return notifications, total, err
}

// GetUnreadCount gets unread notification count
func (r *NotificationRepository) GetUnreadCount(userID string) int64 {
	var count int64
	r.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count)
	return count
}

// Create creates a new notification
func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.DB.Create(notification).Error
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(id, userID string) error {
	return r.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(userID string) error {
	return r.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(id, userID string) error {
	return r.DB.Delete(&models.Notification{}, "id = ? AND user_id = ?", id, userID).Error
}

// DeleteOld deletes notifications older than specified days
func (r *NotificationRepository) DeleteOld(userID string, days int) error {
	return r.DB.Exec(`
		DELETE FROM notifications 
		WHERE user_id = ? AND created_at < NOW() - INTERVAL '? days'
	`, userID, days).Error
}
