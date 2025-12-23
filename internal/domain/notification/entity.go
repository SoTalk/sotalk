package notification

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeMessage  NotificationType = "message"
	NotificationTypeGroup    NotificationType = "group"
	NotificationTypeChannel  NotificationType = "channel"
	NotificationTypePayment  NotificationType = "payment"
	NotificationTypeMention  NotificationType = "mention"
	NotificationTypeReaction NotificationType = "reaction"
	NotificationTypeSystem   NotificationType = "system"
)

// Notification represents a user notification
type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      NotificationType
	Title     string
	Body      string
	Data      map[string]interface{}
	Read      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NotificationSettings represents user notification preferences
type NotificationSettings struct {
	UserID           uuid.UUID
	MessagesEnabled  bool
	GroupsEnabled    bool
	ChannelsEnabled  bool
	PaymentsEnabled  bool
	MentionsEnabled  bool
	ReactionsEnabled bool
	SoundEnabled     bool
	VibrationEnabled bool
	UpdatedAt        time.Time
}
