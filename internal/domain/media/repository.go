package media

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the media repository interface
type Repository interface {
	// Create creates a new media record
	Create(ctx context.Context, media *Media) error

	// FindByID finds a media by ID
	FindByID(ctx context.Context, id uuid.UUID) (*Media, error)

	// FindByUserID finds all media uploaded by a user
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Media, error)

	// FindByMessageID finds all media attached to a message
	FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]*Media, error)

	// FindByType finds media by type
	FindByType(ctx context.Context, userID uuid.UUID, mediaType Type, limit, offset int) ([]*Media, error)

	// Update updates a media record
	Update(ctx context.Context, media *Media) error

	// Delete deletes a media record
	Delete(ctx context.Context, id uuid.UUID) error

	// GetTotalSize gets total size of media uploaded by user
	GetTotalSize(ctx context.Context, userID uuid.UUID) (int64, error)
}
