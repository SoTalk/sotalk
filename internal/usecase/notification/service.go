package notification

import (
	"context"

	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the notification service interface
type Service interface {
	// Notifications
	GetNotifications(ctx context.Context, req *dto.GetNotificationsRequest) (*dto.GetNotificationsResponse, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, req *dto.MarkNotificationAsReadRequest) error
	MarkAllAsRead(ctx context.Context, userID string) error
	DeleteNotification(ctx context.Context, req *dto.DeleteNotificationRequest) error

	// Settings
	GetSettings(ctx context.Context, userID string) (*dto.GetNotificationSettingsResponse, error)
	UpdateSettings(ctx context.Context, req *dto.UpdateNotificationSettingsRequest) (*dto.UpdateNotificationSettingsResponse, error)
}
