package request

// UpdateNotificationSettingsRequest is the HTTP request for updating notification settings
type UpdateNotificationSettingsRequest struct {
	MessagesEnabled  *bool `json:"messages_enabled"`
	GroupsEnabled    *bool `json:"groups_enabled"`
	ChannelsEnabled  *bool `json:"channels_enabled"`
	PaymentsEnabled  *bool `json:"payments_enabled"`
	MentionsEnabled  *bool `json:"mentions_enabled"`
	ReactionsEnabled *bool `json:"reactions_enabled"`
	SoundEnabled     *bool `json:"sound_enabled"`
	VibrationEnabled *bool `json:"vibration_enabled"`
}
