package media

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the media use case interface
type Service interface {
	// UploadMedia uploads a media file
	UploadMedia(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.UploadMediaResponse, error)

	// GetMedia gets media by ID
	GetMedia(ctx context.Context, userID, mediaID uuid.UUID) (*dto.GetMediaResponse, error)

	// GetUserMedia gets all media uploaded by a user
	GetUserMedia(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetMediaListResponse, error)

	// GetMessageMedia gets all media attached to a message
	GetMessageMedia(ctx context.Context, userID, messageID uuid.UUID) (*dto.GetMediaListResponse, error)

	// DeleteMedia deletes a media
	DeleteMedia(ctx context.Context, userID, mediaID uuid.UUID) error

	// GetUserStorageSize gets total storage used by user
	GetUserStorageSize(ctx context.Context, userID uuid.UUID) (int64, error)
}
