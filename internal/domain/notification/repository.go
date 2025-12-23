package notification

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for notification data operations
type Repository interface {
	// Notification operations
	Create(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Settings operations
	GetSettings(ctx context.Context, userID uuid.UUID) (*NotificationSettings, error)
	UpdateSettings(ctx context.Context, settings *NotificationSettings) error
	CreateSettings(ctx context.Context, settings *NotificationSettings) error
}
