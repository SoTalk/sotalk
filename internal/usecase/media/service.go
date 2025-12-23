package media

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/media"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/infrastructure/storage"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

const (
	// MaxFileSize is the maximum file size (50MB)
	MaxFileSize = 50 * 1024 * 1024

	// MaxStoragePerUser is the maximum storage per user (1GB)
	MaxStoragePerUser = 1 * 1024 * 1024 * 1024
)

type service struct {
	mediaRepo      media.Repository
	userRepo       user.Repository
	storageService storage.Service
}

// NewService creates a new media service
func NewService(
	mediaRepo media.Repository,
	userRepo user.Repository,
	storageService storage.Service,
) Service {
	return &service{
		mediaRepo:      mediaRepo,
		userRepo:       userRepo,
		storageService: storageService,
	}
}

// UploadMedia uploads a media file
func (s *service) UploadMedia(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.UploadMediaResponse, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate file size
	if file.Size > MaxFileSize {
		return nil, media.ErrFileTooLarge
	}

	// Check user storage quota
	totalSize, err := s.mediaRepo.GetTotalSize(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage size: %w", err)
	}

	if totalSize+file.Size > MaxStoragePerUser {
		return nil, media.ErrStorageQuotaExceeded
	}

	// Determine media type from mime type
	contentType := file.Header.Get("Content-Type")
	mediaType := getMediaType(contentType)
	if mediaType == "" {
		return nil, media.ErrInvalidFileFormat
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate unique filename with user ID prefix
	fileName := fmt.Sprintf("%s/%s-%s", userID.String(), uuid.New().String(), file.Filename)

	// Upload to storage service (Azure Blob or local)
	fileURL, err := s.storageService.Upload(ctx, fileName, contentType, src)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create media entity
	m := media.NewMedia(
		userID,
		media.Type(mediaType),
		file.Filename,
		contentType,
		fileURL,
		file.Size,
	)

	// Save to database
	if err := s.mediaRepo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	return &dto.UploadMediaResponse{
		Media: toMediaDTO(m),
	}, nil
}

// GetMedia gets media by ID
func (s *service) GetMedia(ctx context.Context, userID, mediaID uuid.UUID) (*dto.GetMediaResponse, error) {
	m, err := s.mediaRepo.FindByID(ctx, mediaID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if m.UserID != userID {
		return nil, media.ErrNotAuthorized
	}

	return &dto.GetMediaResponse{
		Media: toMediaDTO(m),
	}, nil
}

// GetUserMedia gets all media uploaded by a user
func (s *service) GetUserMedia(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetMediaListResponse, error) {
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	mediaList, err := s.mediaRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user media: %w", err)
	}

	mediaDTOs := make([]dto.MediaDTO, len(mediaList))
	for i, m := range mediaList {
		mediaDTOs[i] = toMediaDTO(m)
	}

	return &dto.GetMediaListResponse{
		Media: mediaDTOs,
		Total: len(mediaDTOs),
	}, nil
}

// GetMessageMedia gets all media attached to a message
func (s *service) GetMessageMedia(ctx context.Context, userID, messageID uuid.UUID) (*dto.GetMediaListResponse, error) {
	// Get media attached to the message
	mediaList, err := s.mediaRepo.FindByMessageID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message media: %w", err)
	}

	// Verify user has access by checking if any media belongs to user
	// In a full implementation, we would check if user is participant in the conversation
	if len(mediaList) > 0 {
		hasAccess := false
		for _, m := range mediaList {
			if m.UserID == userID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			return nil, media.ErrNotAuthorized
		}
	}

	mediaDTOs := make([]dto.MediaDTO, len(mediaList))
	for i, m := range mediaList {
		mediaDTOs[i] = toMediaDTO(m)
	}

	return &dto.GetMediaListResponse{
		Media: mediaDTOs,
		Total: len(mediaDTOs),
	}, nil
}

// DeleteMedia deletes a media
func (s *service) DeleteMedia(ctx context.Context, userID, mediaID uuid.UUID) error {
	m, err := s.mediaRepo.FindByID(ctx, mediaID)
	if err != nil {
		return err
	}

	// Check authorization
	if m.UserID != userID {
		return media.ErrNotAuthorized
	}

	// Delete actual file from storage (Azure Blob or local)
	if m.URL != "" {
		if err := s.storageService.Delete(ctx, m.URL); err != nil {
			// Log error but continue with database deletion
			fmt.Printf("Warning: failed to delete file from storage: %v\n", err)
		}
	}

	// Delete from database
	if err := s.mediaRepo.Delete(ctx, mediaID); err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	return nil
}

// GetUserStorageSize gets total storage used by user
func (s *service) GetUserStorageSize(ctx context.Context, userID uuid.UUID) (int64, error) {
	totalSize, err := s.mediaRepo.GetTotalSize(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get storage size: %w", err)
	}
	return totalSize, nil
}

// Helper functions

func getMediaType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case strings.HasPrefix(mimeType, "application/pdf"),
		strings.HasPrefix(mimeType, "application/msword"),
		strings.HasPrefix(mimeType, "application/vnd.openxmlformats"):
		return "document"
	default:
		return ""
	}
}

func toMediaDTO(m *media.Media) dto.MediaDTO {
	var metadata *dto.MediaMetadata
	if m.Metadata != nil {
		metadata = &dto.MediaMetadata{
			Width:      m.Metadata.Width,
			Height:     m.Metadata.Height,
			Duration:   m.Metadata.Duration,
			Waveform:   m.Metadata.Waveform,
			Format:     m.Metadata.Format,
			Resolution: m.Metadata.Resolution,
		}
	}

	var messageID *string
	if m.MessageID != nil {
		id := m.MessageID.String()
		messageID = &id
	}

	return dto.MediaDTO{
		ID:           m.ID.String(),
		UserID:       m.UserID.String(),
		MessageID:    messageID,
		Type:         string(m.Type),
		FileName:     m.FileName,
		FileSize:     m.FileSize,
		FileSizeMB:   m.GetFileSizeInMB(),
		MimeType:     m.MimeType,
		URL:          m.URL,
		ThumbnailURL: m.ThumbnailURL,
		Metadata:     metadata,
		CreatedAt:    m.CreatedAt,
	}
}
