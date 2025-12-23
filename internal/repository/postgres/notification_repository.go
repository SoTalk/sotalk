package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	domainNotification "github.com/yourusername/sotalk/internal/domain/notification"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) domainNotification.Repository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domainNotification.Notification) error {
	model := &Notification{
		UserID:    notification.UserID,
		Type:      string(notification.Type),
		Title:     notification.Title,
		Body:      notification.Body,
		Data:      notification.Data,
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	notification.ID = model.ID
	notification.CreatedAt = model.CreatedAt
	notification.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainNotification.Notification, error) {
	var model Notification
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainNotification.ErrNotificationNotFound
		}
		return nil, err
	}

	return &domainNotification.Notification{
		ID:        model.ID,
		UserID:    model.UserID,
		Type:      domainNotification.NotificationType(model.Type),
		Title:     model.Title,
		Body:      model.Body,
		Data:      model.Data,
		Read:      model.Read,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domainNotification.Notification, error) {
	var models []Notification

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
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

	result := make([]*domainNotification.Notification, len(models))
	for i, m := range models {
		result[i] = &domainNotification.Notification{
			ID:        m.ID,
			UserID:    m.UserID,
			Type:      domainNotification.NotificationType(m.Type),
			Title:     m.Title,
			Body:      m.Body,
			Data:      m.Data,
			Read:      m.Read,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
	}

	return result, nil
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ?", id).
		Update("read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true).Error
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&Notification{}, "id = ?", id).Error
}

func (r *notificationRepository) GetSettings(ctx context.Context, userID uuid.UUID) (*domainNotification.NotificationSettings, error) {
	var model NotificationSettings
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default settings if not found
			return &domainNotification.NotificationSettings{
				UserID:           userID,
				MessagesEnabled:  true,
				GroupsEnabled:    true,
				ChannelsEnabled:  true,
				PaymentsEnabled:  true,
				MentionsEnabled:  true,
				ReactionsEnabled: true,
				SoundEnabled:     true,
				VibrationEnabled: true,
			}, nil
		}
		return nil, err
	}

	return &domainNotification.NotificationSettings{
		UserID:           model.UserID,
		MessagesEnabled:  model.MessagesEnabled,
		GroupsEnabled:    model.GroupsEnabled,
		ChannelsEnabled:  model.ChannelsEnabled,
		PaymentsEnabled:  model.PaymentsEnabled,
		MentionsEnabled:  model.MentionsEnabled,
		ReactionsEnabled: model.ReactionsEnabled,
		SoundEnabled:     model.SoundEnabled,
		VibrationEnabled: model.VibrationEnabled,
		UpdatedAt:        model.UpdatedAt,
	}, nil
}

func (r *notificationRepository) UpdateSettings(ctx context.Context, settings *domainNotification.NotificationSettings) error {
	model := &NotificationSettings{
		UserID:           settings.UserID,
		MessagesEnabled:  settings.MessagesEnabled,
		GroupsEnabled:    settings.GroupsEnabled,
		ChannelsEnabled:  settings.ChannelsEnabled,
		PaymentsEnabled:  settings.PaymentsEnabled,
		MentionsEnabled:  settings.MentionsEnabled,
		ReactionsEnabled: settings.ReactionsEnabled,
		SoundEnabled:     settings.SoundEnabled,
		VibrationEnabled: settings.VibrationEnabled,
	}

	// Use Save to upsert - insert if not exists, update if exists
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *notificationRepository) CreateSettings(ctx context.Context, settings *domainNotification.NotificationSettings) error {
	model := &NotificationSettings{
		UserID:           settings.UserID,
		MessagesEnabled:  settings.MessagesEnabled,
		GroupsEnabled:    settings.GroupsEnabled,
		ChannelsEnabled:  settings.ChannelsEnabled,
		PaymentsEnabled:  settings.PaymentsEnabled,
		MentionsEnabled:  settings.MentionsEnabled,
		ReactionsEnabled: settings.ReactionsEnabled,
		SoundEnabled:     settings.SoundEnabled,
		VibrationEnabled: settings.VibrationEnabled,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	settings.UpdatedAt = model.UpdatedAt
	return nil
}
