package status

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// StatusPrivacy defines who can view a status
type StatusPrivacy string

const (
	StatusPrivacyEveryone StatusPrivacy = "everyone"
	StatusPrivacyContacts StatusPrivacy = "contacts"
	StatusPrivacyPrivate  StatusPrivacy = "private"
)

// Status represents a user's temporary status/story
type Status struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.UUID     `json:"user_id"`
	MediaID   uuid.UUID     `json:"media_id"` // Reference to media table
	Caption   string        `json:"caption"`
	Privacy   StatusPrivacy `json:"privacy"`
	ViewCount int           `json:"view_count"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
}

// StatusView tracks who viewed a status
type StatusView struct {
	ID       uuid.UUID `json:"id"`
	StatusID uuid.UUID `json:"status_id"`
	ViewerID uuid.UUID `json:"viewer_id"`
	ViewedAt time.Time `json:"viewed_at"`
}

// IsExpired checks if the status has expired
func (s *Status) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// CanView checks if a user can view this status based on privacy settings
func (s *Status) CanView(viewerID uuid.UUID, isContact bool) bool {
	// Owner can always view
	if s.UserID == viewerID {
		return true
	}

	switch s.Privacy {
	case StatusPrivacyEveryone:
		return true
	case StatusPrivacyContacts:
		return isContact
	case StatusPrivacyPrivate:
		return false
	default:
		return false
	}
}

// Repository defines the interface for status operations
type Repository interface {
	// Status CRUD
	CreateStatus(ctx context.Context, status *Status) error
	GetStatus(ctx context.Context, statusID uuid.UUID) (*Status, error)
	GetUserStatuses(ctx context.Context, userID uuid.UUID) ([]*Status, error)
	GetStatusFeed(ctx context.Context, viewerID uuid.UUID, limit, offset int) ([]*Status, error)
	DeleteStatus(ctx context.Context, statusID uuid.UUID) error
	DeleteExpiredStatuses(ctx context.Context) error

	// Status Views
	AddView(ctx context.Context, view *StatusView) error
	GetStatusViews(ctx context.Context, statusID uuid.UUID) ([]*StatusView, error)
	HasViewed(ctx context.Context, statusID, viewerID uuid.UUID) (bool, error)
	IncrementViewCount(ctx context.Context, statusID uuid.UUID) error
}

// Errors
var (
	ErrStatusNotFound     = errors.New("status not found")
	ErrStatusExpired      = errors.New("status has expired")
	ErrUnauthorizedView   = errors.New("not authorized to view this status")
	ErrAlreadyViewed      = errors.New("status already viewed by user")
	ErrInvalidPrivacy     = errors.New("invalid privacy setting")
)
