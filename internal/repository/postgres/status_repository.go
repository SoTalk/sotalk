package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	domainStatus "github.com/yourusername/sotalk/internal/domain/status"
	"gorm.io/gorm"
)

type statusRepository struct {
	db *gorm.DB
}

// NewStatusRepository creates a new status repository
func NewStatusRepository(db *gorm.DB) domainStatus.Repository {
	return &statusRepository{db: db}
}

func (r *statusRepository) CreateStatus(ctx context.Context, status *domainStatus.Status) error {
	model := &Status{
		UserID:    status.UserID,
		MediaID:   status.MediaID,
		Caption:   status.Caption,
		Privacy:   string(status.Privacy),
		ViewCount: 0,
		ExpiresAt: status.ExpiresAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	status.ID = model.ID
	status.CreatedAt = model.CreatedAt
	return nil
}

func (r *statusRepository) GetStatus(ctx context.Context, statusID uuid.UUID) (*domainStatus.Status, error) {
	var model Status
	if err := r.db.WithContext(ctx).Where("id = ?", statusID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainStatus.ErrStatusNotFound
		}
		return nil, err
	}

	return &domainStatus.Status{
		ID:        model.ID,
		UserID:    model.UserID,
		MediaID:   model.MediaID,
		Caption:   model.Caption,
		Privacy:   domainStatus.StatusPrivacy(model.Privacy),
		ViewCount: model.ViewCount,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
	}, nil
}

func (r *statusRepository) GetUserStatuses(ctx context.Context, userID uuid.UUID) ([]*domainStatus.Status, error) {
	var models []Status
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainStatus.Status, len(models))
	for i, m := range models {
		result[i] = &domainStatus.Status{
			ID:        m.ID,
			UserID:    m.UserID,
			MediaID:   m.MediaID,
			Caption:   m.Caption,
			Privacy:   domainStatus.StatusPrivacy(m.Privacy),
			ViewCount: m.ViewCount,
			ExpiresAt: m.ExpiresAt,
			CreatedAt: m.CreatedAt,
		}
	}

	return result, nil
}

func (r *statusRepository) GetStatusFeed(ctx context.Context, viewerID uuid.UUID, limit, offset int) ([]*domainStatus.Status, error) {
	// Get statuses from contacts and public statuses
	// For now, simple implementation - return recent statuses
	var models []Status
	query := r.db.WithContext(ctx).
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainStatus.Status, len(models))
	for i, m := range models {
		result[i] = &domainStatus.Status{
			ID:        m.ID,
			UserID:    m.UserID,
			MediaID:   m.MediaID,
			Caption:   m.Caption,
			Privacy:   domainStatus.StatusPrivacy(m.Privacy),
			ViewCount: m.ViewCount,
			ExpiresAt: m.ExpiresAt,
			CreatedAt: m.CreatedAt,
		}
	}

	return result, nil
}

func (r *statusRepository) DeleteStatus(ctx context.Context, statusID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", statusID).Delete(&Status{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainStatus.ErrStatusNotFound
	}

	return nil
}

func (r *statusRepository) DeleteExpiredStatuses(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&Status{}).Error
}

// Status Views

func (r *statusRepository) AddView(ctx context.Context, view *domainStatus.StatusView) error {
	// Check if already viewed
	var existing StatusView
	err := r.db.WithContext(ctx).
		Where("status_id = ? AND viewer_id = ?", view.StatusID, view.ViewerID).
		First(&existing).Error

	if err == nil {
		// Already viewed
		return domainStatus.ErrAlreadyViewed
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := &StatusView{
		StatusID: view.StatusID,
		ViewerID: view.ViewerID,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	view.ID = model.ID
	view.ViewedAt = model.ViewedAt
	return nil
}

func (r *statusRepository) GetStatusViews(ctx context.Context, statusID uuid.UUID) ([]*domainStatus.StatusView, error) {
	var models []StatusView
	if err := r.db.WithContext(ctx).
		Where("status_id = ?", statusID).
		Order("viewed_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainStatus.StatusView, len(models))
	for i, m := range models {
		result[i] = &domainStatus.StatusView{
			ID:       m.ID,
			StatusID: m.StatusID,
			ViewerID: m.ViewerID,
			ViewedAt: m.ViewedAt,
		}
	}

	return result, nil
}

func (r *statusRepository) HasViewed(ctx context.Context, statusID, viewerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&StatusView{}).
		Where("status_id = ? AND viewer_id = ?", statusID, viewerID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *statusRepository) IncrementViewCount(ctx context.Context, statusID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Status{}).
		Where("id = ?", statusID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}
