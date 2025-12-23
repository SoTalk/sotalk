package dto

import "time"

// NotificationDTO represents a notification
type NotificationDTO struct {
	ID        string
	Type      string
	Title     string
	Body      string
	Data      map[string]interface{}
	Read      bool
	CreatedAt time.Time
}

// GetNotificationsRequest is the request DTO for getting notifications
type GetNotificationsRequest struct {
	UserID string
	Limit  int
	Offset int
}

// GetNotificationsResponse is the response DTO for getting notifications
type GetNotificationsResponse struct {
	Notifications []NotificationDTO
	Total         int
	Unread        int
}

// MarkNotificationAsReadRequest is the request DTO for marking notification as read
type MarkNotificationAsReadRequest struct {
	UserID         string
	NotificationID string
}

// UpdateNotificationSettingsRequest is the request DTO for updating notification settings
type UpdateNotificationSettingsRequest struct {
	UserID           string
	MessagesEnabled  *bool
	GroupsEnabled    *bool
	ChannelsEnabled  *bool
	PaymentsEnabled  *bool
	MentionsEnabled  *bool
	ReactionsEnabled *bool
	SoundEnabled     *bool
	VibrationEnabled *bool
}

// UpdateNotificationSettingsResponse is the response DTO for updating notification settings
type UpdateNotificationSettingsResponse struct {
	MessagesEnabled  bool
	GroupsEnabled    bool
	ChannelsEnabled  bool
	PaymentsEnabled  bool
	MentionsEnabled  bool
	ReactionsEnabled bool
	SoundEnabled     bool
	VibrationEnabled bool
}

// GetNotificationSettingsResponse is the response DTO for getting notification settings
type GetNotificationSettingsResponse struct {
	MessagesEnabled  bool
	GroupsEnabled    bool
	ChannelsEnabled  bool
	PaymentsEnabled  bool
	MentionsEnabled  bool
	ReactionsEnabled bool
	SoundEnabled     bool
	VibrationEnabled bool
}

// DeleteNotificationRequest is the request DTO for deleting notification
type DeleteNotificationRequest struct {
	UserID         string
	NotificationID string
}
