package privacy

import (
	"time"

	"github.com/google/uuid"
)

// Visibility defines who can see certain information
type Visibility string

const (
	VisibilityEveryone Visibility = "everyone"
	VisibilityContacts Visibility = "contacts"
	VisibilityNobody   Visibility = "nobody"
)

// PrivacySettings represents a user's privacy preferences
type PrivacySettings struct {
	UserID                 uuid.UUID  `json:"user_id"`
	ProfilePhotoVisibility Visibility `json:"profile_photo_visibility"`
	LastSeenVisibility     Visibility `json:"last_seen_visibility"`
	StatusVisibility       Visibility `json:"status_visibility"`
	ReadReceiptsEnabled    bool       `json:"read_receipts_enabled"`
	TypingIndicatorEnabled bool       `json:"typing_indicator_enabled"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// BlockedUser represents a user blocking relationship
type BlockedUser struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`         // User who blocked
	BlockedUserID uuid.UUID `json:"blocked_user_id"` // User who was blocked
	Reason        string    `json:"reason,omitempty"`
	BlockedAt     time.Time `json:"blocked_at"`
}

// DisappearingMessagesConfig represents disappearing message settings for a conversation
type DisappearingMessagesConfig struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	DurationSeconds int       `json:"duration_seconds"` // 0 = disabled
	EnabledBy      uuid.UUID `json:"enabled_by"`
	EnabledAt      time.Time `json:"enabled_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TwoFactorAuth represents 2FA settings for a user
type TwoFactorAuth struct {
	UserID       uuid.UUID `json:"user_id"`
	Enabled      bool      `json:"enabled"`
	Secret       string    `json:"secret"`        // Encrypted TOTP secret
	BackupCodes  []string  `json:"backup_codes"`  // Encrypted backup codes
	EnabledAt    time.Time `json:"enabled_at,omitempty"`
	LastUsedAt   time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	Key        string    `json:"key"`         // e.g., "user:123:messages"
	Count      int       `json:"count"`       // Current count
	Limit      int       `json:"limit"`       // Maximum allowed
	WindowSecs int       `json:"window_secs"` // Time window in seconds
	ExpiresAt  time.Time `json:"expires_at"`
}

// IsBlocked checks if current user has blocked the target user
func (ps *PrivacySettings) CanSeeProfilePhoto(relationship string) bool {
	switch ps.ProfilePhotoVisibility {
	case VisibilityEveryone:
		return true
	case VisibilityContacts:
		return relationship == "contact"
	case VisibilityNobody:
		return false
	default:
		return false
	}
}

// CanSeeLastSeen checks visibility of last seen
func (ps *PrivacySettings) CanSeeLastSeen(relationship string) bool {
	switch ps.LastSeenVisibility {
	case VisibilityEveryone:
		return true
	case VisibilityContacts:
		return relationship == "contact"
	case VisibilityNobody:
		return false
	default:
		return false
	}
}

// CanSeeStatus checks visibility of status
func (ps *PrivacySettings) CanSeeStatus(relationship string) bool {
	switch ps.StatusVisibility {
	case VisibilityEveryone:
		return true
	case VisibilityContacts:
		return relationship == "contact"
	case VisibilityNobody:
		return false
	default:
		return false
	}
}

// IsDisappearing checks if messages should disappear
func (dmc *DisappearingMessagesConfig) IsDisappearing() bool {
	return dmc.DurationSeconds > 0
}

// GetExpiryTime calculates when a message should disappear
func (dmc *DisappearingMessagesConfig) GetExpiryTime(messageTime time.Time) time.Time {
	if !dmc.IsDisappearing() {
		return time.Time{} // Zero time = never expires
	}
	return messageTime.Add(time.Duration(dmc.DurationSeconds) * time.Second)
}
