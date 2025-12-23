package response

import "time"

// NotificationDTO is the notification data in response
type NotificationDTO struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
}

// NotificationsResponse is the HTTP response for notifications list
type NotificationsResponse struct {
	Notifications []NotificationDTO `json:"notifications"`
	Total         int               `json:"total"`
	Unread        int               `json:"unread"`
}

// UnreadCountResponse is the HTTP response for unread count
type UnreadCountResponse struct {
	Count int `json:"count"`
}

// NotificationSettingsResponse is the HTTP response for notification settings
type NotificationSettingsResponse struct {
	MessagesEnabled  bool `json:"messages_enabled"`
	GroupsEnabled    bool `json:"groups_enabled"`
	ChannelsEnabled  bool `json:"channels_enabled"`
	PaymentsEnabled  bool `json:"payments_enabled"`
	MentionsEnabled  bool `json:"mentions_enabled"`
	ReactionsEnabled bool `json:"reactions_enabled"`
	SoundEnabled     bool `json:"sound_enabled"`
	VibrationEnabled bool `json:"vibration_enabled"`
}
