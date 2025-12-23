package user

import (
	"time"

	"github.com/google/uuid"
)

// Preferences represents user preferences
type Preferences struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Language  string
	Theme     string
	NotificationsEnabled bool
	SoundEnabled bool
	EmailNotifications bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPreferences creates new user preferences with defaults
func NewPreferences(userID uuid.UUID) *Preferences {
	now := time.Now()
	return &Preferences{
		ID:                   uuid.New(),
		UserID:               userID,
		Language:             "en",
		Theme:                "dark",
		NotificationsEnabled: true,
		SoundEnabled:         true,
		EmailNotifications:   true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// UpdateLanguage updates the language preference
func (p *Preferences) UpdateLanguage(language string) {
	p.Language = language
	p.UpdatedAt = time.Now()
}

// UpdateTheme updates the theme preference
func (p *Preferences) UpdateTheme(theme string) {
	p.Theme = theme
	p.UpdatedAt = time.Now()
}

// UpdateNotifications updates notification settings
func (p *Preferences) UpdateNotifications(enabled bool) {
	p.NotificationsEnabled = enabled
	p.UpdatedAt = time.Now()
}

// UpdateSound updates sound settings
func (p *Preferences) UpdateSound(enabled bool) {
	p.SoundEnabled = enabled
	p.UpdatedAt = time.Now()
}

// UpdateEmailNotifications updates email notification settings
func (p *Preferences) UpdateEmailNotifications(enabled bool) {
	p.EmailNotifications = enabled
	p.UpdatedAt = time.Now()
}
