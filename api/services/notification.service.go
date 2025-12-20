package services

import (
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"

	"gorm.io/gorm"
)

// NotificationService handles notification business logic
type NotificationService struct {
	repo *repositories.NotificationRepository
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{
		repo: repositories.NewNotificationRepository(db),
	}
}

// ListNotifications lists user's notifications
func (s *NotificationService) ListNotifications(userID string, req dto.NotificationListRequest) (*dto.NotificationListResponse, error) {
	notifications, total, err := s.repo.FindByUser(userID, req.Type, req.UnreadOnly, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = dto.NotificationResponse{
			ID:        n.ID,
			Type:      string(n.Type),
			Title:     n.Title,
			Message:   n.Message,
			Data:      n.Data,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		}
	}

	unreadCount := s.repo.GetUnreadCount(userID)

	return &dto.NotificationListResponse{
		Notifications: responses,
		UnreadCount:   int(unreadCount),
		Pagination: dto.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
		},
	}, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID, userID string) error {
	return s.repo.MarkAsRead(notificationID, userID)
}

// MarkAllAsRead marks all notifications as read
func (s *NotificationService) MarkAllAsRead(userID string) error {
	return s.repo.MarkAllAsRead(userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID, userID string) error {
	return s.repo.Delete(notificationID, userID)
}

// GetUnreadCount gets unread notification count
func (s *NotificationService) GetUnreadCount(userID string) int {
	return int(s.repo.GetUnreadCount(userID))
}
