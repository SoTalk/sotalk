package channel

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the channel use case interface
type Service interface {
	// CreateChannel creates a new channel
	CreateChannel(ctx context.Context, userID uuid.UUID, req *dto.CreateChannelRequest) (*dto.CreateChannelResponse, error)

	// GetChannel gets channel information
	GetChannel(ctx context.Context, userID uuid.UUID, username string) (*dto.GetChannelResponse, error)

	// GetPublicChannels gets all public channels
	GetPublicChannels(ctx context.Context, limit, offset int) (*dto.GetChannelsResponse, error)

	// GetUserChannels gets all channels owned by a user
	GetUserChannels(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetChannelsResponse, error)

	// GetUserSubscriptions gets all channels a user is subscribed to
	GetUserSubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetSubscriptionsResponse, error)

	// UpdateChannel updates channel information
	UpdateChannel(ctx context.Context, userID, channelID uuid.UUID, req *dto.UpdateChannelRequest) error

	// DeleteChannel deletes a channel
	DeleteChannel(ctx context.Context, userID, channelID uuid.UUID) error

	// Subscribe subscribes a user to a channel
	Subscribe(ctx context.Context, userID uuid.UUID, username string) error

	// Unsubscribe unsubscribes a user from a channel
	Unsubscribe(ctx context.Context, userID uuid.UUID, username string) error

	// ToggleNotifications toggles notifications for a channel subscription
	ToggleNotifications(ctx context.Context, userID, channelID uuid.UUID) error

	// AddAdmin adds an admin to a channel
	AddAdmin(ctx context.Context, userID, channelID uuid.UUID, req *dto.AddAdminRequest) error

	// RemoveAdmin removes an admin from a channel
	RemoveAdmin(ctx context.Context, userID, channelID, targetUserID uuid.UUID) error

	// UpdateAdminPermissions updates an admin's permissions
	UpdateAdminPermissions(ctx context.Context, userID, channelID, targetUserID uuid.UUID, req *dto.UpdateAdminPermissionsRequest) error
}
