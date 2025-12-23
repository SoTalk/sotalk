package channel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/channel"
	"github.com/yourusername/sotalk/internal/domain/conversation"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

type service struct {
	channelRepo      channel.Repository
	conversationRepo conversation.Repository
	userRepo         user.Repository
}

// NewService creates a new channel service
func NewService(
	channelRepo channel.Repository,
	conversationRepo conversation.Repository,
	userRepo user.Repository,
) Service {
	return &service{
		channelRepo:      channelRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
	}
}

// CreateChannel creates a new channel
func (s *service) CreateChannel(ctx context.Context, userID uuid.UUID, req *dto.CreateChannelRequest) (*dto.CreateChannelResponse, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("owner not found: %w", err)
	}

	// Create conversation for the channel
	conv := conversation.NewChannelConversation()
	if err := s.conversationRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Create channel
	ch := channel.NewChannel(req.Name, req.Username, req.Description, userID, conv.ID, req.IsPublic)
	if err := s.channelRepo.Create(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Auto-subscribe owner
	subscriber := channel.NewSubscriber(ch.ID, userID)
	s.channelRepo.Subscribe(ctx, subscriber)

	return &dto.CreateChannelResponse{
		Channel: toChannelDTO(ch),
	}, nil
}

// GetChannel gets channel information
func (s *service) GetChannel(ctx context.Context, userID uuid.UUID, username string) (*dto.GetChannelResponse, error) {
	ch, err := s.channelRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Check if user is subscribed
	isSubscribed, _ := s.channelRepo.IsSubscribed(ctx, ch.ID, userID)
	isOwner := ch.OwnerID == userID
	isAdmin, _ := s.channelRepo.IsAdmin(ctx, ch.ID, userID)

	return &dto.GetChannelResponse{
		Channel:      toChannelDTO(ch),
		IsSubscribed: isSubscribed,
		IsOwner:      isOwner,
		IsAdmin:      isAdmin,
	}, nil
}

// GetPublicChannels gets all public channels
func (s *service) GetPublicChannels(ctx context.Context, limit, offset int) (*dto.GetChannelsResponse, error) {
	if limit == 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	channels, err := s.channelRepo.FindPublicChannels(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get public channels: %w", err)
	}

	channelDTOs := make([]dto.ChannelDTO, len(channels))
	for i, ch := range channels {
		channelDTOs[i] = toChannelDTO(ch)
	}

	return &dto.GetChannelsResponse{
		Channels: channelDTOs,
	}, nil
}

// GetUserChannels gets all channels owned by a user
func (s *service) GetUserChannels(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetChannelsResponse, error) {
	if limit == 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	channels, err := s.channelRepo.FindByOwnerID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user channels: %w", err)
	}

	channelDTOs := make([]dto.ChannelDTO, len(channels))
	for i, ch := range channels {
		channelDTOs[i] = toChannelDTO(ch)
	}

	return &dto.GetChannelsResponse{
		Channels: channelDTOs,
	}, nil
}

// GetUserSubscriptions gets all channels a user is subscribed to
func (s *service) GetUserSubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetSubscriptionsResponse, error) {
	if limit == 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	channels, err := s.channelRepo.FindUserSubscriptions(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	channelDTOs := make([]dto.ChannelDTO, len(channels))
	for i, ch := range channels {
		channelDTOs[i] = toChannelDTO(ch)
	}

	return &dto.GetSubscriptionsResponse{
		Channels: channelDTOs,
	}, nil
}

// UpdateChannel updates channel information
func (s *service) UpdateChannel(ctx context.Context, userID, channelID uuid.UUID, req *dto.UpdateChannelRequest) error {
	ch, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return err
	}

	// Only owner can update
	if ch.OwnerID != userID {
		return channel.ErrNotOwner
	}

	ch.UpdateInfo(req.Name, req.Description, req.Avatar)

	if err := s.channelRepo.Update(ctx, ch); err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

// DeleteChannel deletes a channel
func (s *service) DeleteChannel(ctx context.Context, userID, channelID uuid.UUID) error {
	ch, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return err
	}

	// Only owner can delete
	if ch.OwnerID != userID {
		return channel.ErrNotOwner
	}

	if err := s.channelRepo.Delete(ctx, channelID); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

// Subscribe subscribes a user to a channel
func (s *service) Subscribe(ctx context.Context, userID uuid.UUID, username string) error {
	ch, err := s.channelRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}

	// Check if already subscribed
	isSubscribed, err := s.channelRepo.IsSubscribed(ctx, ch.ID, userID)
	if err != nil {
		return err
	}
	if isSubscribed {
		return channel.ErrAlreadySubscribed
	}

	// Subscribe
	subscriber := channel.NewSubscriber(ch.ID, userID)
	if err := s.channelRepo.Subscribe(ctx, subscriber); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Increment subscriber count
	ch.IncrementSubscriberCount()
	s.channelRepo.Update(ctx, ch)

	return nil
}

// Unsubscribe unsubscribes a user from a channel
func (s *service) Unsubscribe(ctx context.Context, userID uuid.UUID, username string) error {
	ch, err := s.channelRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}

	// Cannot unsubscribe if owner
	if ch.OwnerID == userID {
		return fmt.Errorf("owner cannot unsubscribe from their own channel")
	}

	if err := s.channelRepo.Unsubscribe(ctx, ch.ID, userID); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	// Decrement subscriber count
	ch.DecrementSubscriberCount()
	s.channelRepo.Update(ctx, ch)

	return nil
}

// ToggleNotifications toggles notifications for a channel subscription
func (s *service) ToggleNotifications(ctx context.Context, userID, channelID uuid.UUID) error {
	subscriber, err := s.channelRepo.FindSubscriber(ctx, channelID, userID)
	if err != nil {
		return err
	}

	subscriber.ToggleNotifications()

	if err := s.channelRepo.UpdateSubscriberNotifications(ctx, channelID, userID, subscriber.NotificationsEnabled); err != nil {
		return fmt.Errorf("failed to update notifications: %w", err)
	}

	return nil
}

// AddAdmin adds an admin to a channel
func (s *service) AddAdmin(ctx context.Context, userID, channelID uuid.UUID, req *dto.AddAdminRequest) error {
	ch, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return err
	}

	// Only owner can add admins
	if ch.OwnerID != userID {
		return channel.ErrNotOwner
	}

	// Parse target user ID
	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if target user exists
	if _, err := s.userRepo.FindByID(ctx, targetUserID); err != nil {
		return fmt.Errorf("target user not found: %w", err)
	}

	// Add admin
	admin := channel.NewAdmin(channelID, targetUserID)
	if err := s.channelRepo.AddAdmin(ctx, admin); err != nil {
		return fmt.Errorf("failed to add admin: %w", err)
	}

	return nil
}

// RemoveAdmin removes an admin from a channel
func (s *service) RemoveAdmin(ctx context.Context, userID, channelID, targetUserID uuid.UUID) error {
	ch, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return err
	}

	// Only owner can remove admins
	if ch.OwnerID != userID {
		return channel.ErrNotOwner
	}

	// Cannot remove owner
	if ch.OwnerID == targetUserID {
		return channel.ErrCannotRemoveOwner
	}

	if err := s.channelRepo.RemoveAdmin(ctx, channelID, targetUserID); err != nil {
		return fmt.Errorf("failed to remove admin: %w", err)
	}

	return nil
}

// UpdateAdminPermissions updates an admin's permissions
func (s *service) UpdateAdminPermissions(ctx context.Context, userID, channelID, targetUserID uuid.UUID, req *dto.UpdateAdminPermissionsRequest) error {
	ch, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return err
	}

	// Only owner can update admin permissions
	if ch.OwnerID != userID {
		return channel.ErrNotOwner
	}

	permissions := &channel.AdminPermissions{
		CanPostMessages:   req.CanPostMessages,
		CanEditMessages:   req.CanEditMessages,
		CanDeleteMessages: req.CanDeleteMessages,
		CanManageAdmins:   req.CanManageAdmins,
	}

	if err := s.channelRepo.UpdateAdminPermissions(ctx, channelID, targetUserID, permissions); err != nil {
		return fmt.Errorf("failed to update permissions: %w", err)
	}

	return nil
}

// Helper function to convert channel to DTO
func toChannelDTO(ch *channel.Channel) dto.ChannelDTO {
	var settings *dto.ChannelSettings
	if ch.Settings != nil {
		settings = &dto.ChannelSettings{
			AdminsCanPost:     ch.Settings.AdminsCanPost,
			LinkPreview:       ch.Settings.LinkPreview,
			ForwardingAllowed: ch.Settings.ForwardingAllowed,
		}
	}

	return dto.ChannelDTO{
		ID:              ch.ID.String(),
		ConversationID:  ch.ConversationID.String(),
		Name:            ch.Name,
		Username:        ch.Username,
		Description:     ch.Description,
		Avatar:          ch.Avatar,
		OwnerID:         ch.OwnerID.String(),
		IsPublic:        ch.IsPublic,
		SubscriberCount: ch.SubscriberCount,
		Settings:        settings,
		CreatedAt:       ch.CreatedAt,
		UpdatedAt:       ch.UpdatedAt,
	}
}
