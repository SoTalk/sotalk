package notification

import "errors"

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrSettingsNotFound     = errors.New("notification settings not found")
	ErrInvalidNotification  = errors.New("invalid notification")
)
