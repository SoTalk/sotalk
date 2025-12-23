package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/media"
	"gorm.io/gorm"
)

type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *gorm.DB) media.Repository {
	return &mediaRepository{
		db: db,
	}
}

// Create creates a new media record
func (r *mediaRepository) Create(ctx context.Context, m *media.Media) error {
	dbMedia := toDBMedia(m)
	result := r.db.WithContext(ctx).Create(dbMedia)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID finds a media by ID
func (r *mediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*media.Media, error) {
	var dbMedia Media
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbMedia)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, media.ErrMediaNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainMedia(&dbMedia), nil
}

// FindByUserID finds all media uploaded by a user
func (r *mediaRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*media.Media, error) {
	var dbMediaList []Media
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbMediaList)

	if result.Error != nil {
		return nil, result.Error
	}

	mediaList := make([]*media.Media, len(dbMediaList))
	for i, dbMedia := range dbMediaList {
		mediaList[i] = toDomainMedia(&dbMedia)
	}

	return mediaList, nil
}

// FindByMessageID finds all media attached to a message
func (r *mediaRepository) FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]*media.Media, error) {
	var dbMediaList []Media
	result := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Order("created_at ASC").
		Find(&dbMediaList)

	if result.Error != nil {
		return nil, result.Error
	}

	mediaList := make([]*media.Media, len(dbMediaList))
	for i, dbMedia := range dbMediaList {
		mediaList[i] = toDomainMedia(&dbMedia)
	}

	return mediaList, nil
}

// FindByType finds media by type
func (r *mediaRepository) FindByType(ctx context.Context, userID uuid.UUID, mediaType media.Type, limit, offset int) ([]*media.Media, error) {
	var dbMediaList []Media
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, string(mediaType)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbMediaList)

	if result.Error != nil {
		return nil, result.Error
	}

	mediaList := make([]*media.Media, len(dbMediaList))
	for i, dbMedia := range dbMediaList {
		mediaList[i] = toDomainMedia(&dbMedia)
	}

	return mediaList, nil
}

// Update updates a media record
func (r *mediaRepository) Update(ctx context.Context, m *media.Media) error {
	dbMedia := toDBMedia(m)
	result := r.db.WithContext(ctx).
		Model(&Media{}).
		Where("id = ?", m.ID).
		Updates(dbMedia)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return media.ErrMediaNotFound
	}
	return nil
}

// Delete deletes a media record
func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Media{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return media.ErrMediaNotFound
	}
	return nil
}

// GetTotalSize gets total size of media uploaded by user
func (r *mediaRepository) GetTotalSize(ctx context.Context, userID uuid.UUID) (int64, error) {
	var totalSize int64
	result := r.db.WithContext(ctx).
		Model(&Media{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize)

	if result.Error != nil {
		return 0, result.Error
	}
	return totalSize, nil
}

// Mapper functions

func toDBMedia(m *media.Media) *Media {
	var metadataJSON string
	if m.Metadata != nil {
		data, _ := json.Marshal(m.Metadata)
		metadataJSON = string(data)
	}

	return &Media{
		ID:           m.ID,
		UserID:       m.UserID,
		MessageID:    m.MessageID,
		Type:         string(m.Type),
		FileName:     m.FileName,
		FileSize:     m.FileSize,
		MimeType:     m.MimeType,
		URL:          m.URL,
		ThumbnailURL: m.ThumbnailURL,
		Metadata:     metadataJSON,
		CreatedAt:    m.CreatedAt,
	}
}

func toDomainMedia(dbMedia *Media) *media.Media {
	var metadata media.Metadata
	if dbMedia.Metadata != "" {
		json.Unmarshal([]byte(dbMedia.Metadata), &metadata)
	}

	return &media.Media{
		ID:           dbMedia.ID,
		UserID:       dbMedia.UserID,
		MessageID:    dbMedia.MessageID,
		Type:         media.Type(dbMedia.Type),
		FileName:     dbMedia.FileName,
		FileSize:     dbMedia.FileSize,
		MimeType:     dbMedia.MimeType,
		URL:          dbMedia.URL,
		ThumbnailURL: dbMedia.ThumbnailURL,
		Metadata:     &metadata,
		CreatedAt:    dbMedia.CreatedAt,
	}
}
