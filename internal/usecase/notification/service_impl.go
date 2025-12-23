package notification

import (
	"context"

	"github.com/google/uuid"
	domainNotification "github.com/yourusername/sotalk/internal/domain/notification"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

type service struct {
	notificationRepo domainNotification.Repository
}

// NewService creates a new notification service
func NewService(notificationRepo domainNotification.Repository) Service {
	return &service{
		notificationRepo: notificationRepo,
	}
}

func (s *service) GetNotifications(ctx context.Context, req *dto.GetNotificationsRequest) (*dto.GetNotificationsResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	// Get notifications with pagination
	notifications, err := s.notificationRepo.GetByUserID(ctx, userID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	// Get unread count
	unreadCount, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	notificationDTOs := make([]dto.NotificationDTO, len(notifications))
	for i, n := range notifications {
		notificationDTOs[i] = dto.NotificationDTO{
			ID:        n.ID.String(),
			Type:      string(n.Type),
			Title:     n.Title,
			Body:      n.Body,
			Data:      n.Data,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
		}
	}

	// For total count, we need to count all notifications (can be optimized)
	// For now, we'll use len of current page + offset as minimum
	total := len(notifications) + req.Offset

	return &dto.GetNotificationsResponse{
		Notifications: notificationDTOs,
		Total:         total,
		Unread:        unreadCount,
	}, nil
}

func (s *service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return 0, err
	}

	return s.notificationRepo.GetUnreadCount(ctx, uid)
}

func (s *service) MarkAsRead(ctx context.Context, req *dto.MarkNotificationAsReadRequest) error {
	notifID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return err
	}

	return s.notificationRepo.MarkAsRead(ctx, notifID)
}

func (s *service) MarkAllAsRead(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return s.notificationRepo.MarkAllAsRead(ctx, uid)
}

func (s *service) DeleteNotification(ctx context.Context, req *dto.DeleteNotificationRequest) error {
	notifID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return err
	}

	return s.notificationRepo.Delete(ctx, notifID)
}

func (s *service) GetSettings(ctx context.Context, userID string) (*dto.GetNotificationSettingsResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	settings, err := s.notificationRepo.GetSettings(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &dto.GetNotificationSettingsResponse{
		MessagesEnabled:  settings.MessagesEnabled,
		GroupsEnabled:    settings.GroupsEnabled,
		ChannelsEnabled:  settings.ChannelsEnabled,
		PaymentsEnabled:  settings.PaymentsEnabled,
		MentionsEnabled:  settings.MentionsEnabled,
		ReactionsEnabled: settings.ReactionsEnabled,
		SoundEnabled:     settings.SoundEnabled,
		VibrationEnabled: settings.VibrationEnabled,
	}, nil
}

func (s *service) UpdateSettings(ctx context.Context, req *dto.UpdateNotificationSettingsRequest) (*dto.UpdateNotificationSettingsResponse, error) {
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	// Get existing settings first
	existingSettings, err := s.notificationRepo.GetSettings(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if req.MessagesEnabled != nil {
		existingSettings.MessagesEnabled = *req.MessagesEnabled
	}
	if req.GroupsEnabled != nil {
		existingSettings.GroupsEnabled = *req.GroupsEnabled
	}
	if req.ChannelsEnabled != nil {
		existingSettings.ChannelsEnabled = *req.ChannelsEnabled
	}
	if req.PaymentsEnabled != nil {
		existingSettings.PaymentsEnabled = *req.PaymentsEnabled
	}
	if req.MentionsEnabled != nil {
		existingSettings.MentionsEnabled = *req.MentionsEnabled
	}
	if req.ReactionsEnabled != nil {
		existingSettings.ReactionsEnabled = *req.ReactionsEnabled
	}
	if req.SoundEnabled != nil {
		existingSettings.SoundEnabled = *req.SoundEnabled
	}
	if req.VibrationEnabled != nil {
		existingSettings.VibrationEnabled = *req.VibrationEnabled
	}

	// Check if settings exist, create or update
	existingInDB, err := s.notificationRepo.GetSettings(ctx, uid)
	if err != nil || existingInDB.UserID == uuid.Nil {
		// Create new settings
		if err := s.notificationRepo.CreateSettings(ctx, existingSettings); err != nil {
			return nil, err
		}
	} else {
		// Update existing settings
		if err := s.notificationRepo.UpdateSettings(ctx, existingSettings); err != nil {
			return nil, err
		}
	}

	return &dto.UpdateNotificationSettingsResponse{
		MessagesEnabled:  existingSettings.MessagesEnabled,
		GroupsEnabled:    existingSettings.GroupsEnabled,
		ChannelsEnabled:  existingSettings.ChannelsEnabled,
		PaymentsEnabled:  existingSettings.PaymentsEnabled,
		MentionsEnabled:  existingSettings.MentionsEnabled,
		ReactionsEnabled: existingSettings.ReactionsEnabled,
		SoundEnabled:     existingSettings.SoundEnabled,
		VibrationEnabled: existingSettings.VibrationEnabled,
	}, nil
}
