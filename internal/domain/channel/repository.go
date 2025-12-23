package channel

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the channel repository interface
type Repository interface {
	// Channel operations
	Create(ctx context.Context, channel *Channel) error
	FindByID(ctx context.Context, id uuid.UUID) (*Channel, error)
	FindByUsername(ctx context.Context, username string) (*Channel, error)
	FindByConversationID(ctx context.Context, conversationID uuid.UUID) (*Channel, error)
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Channel, error)
	FindPublicChannels(ctx context.Context, limit, offset int) ([]*Channel, error)
	Update(ctx context.Context, channel *Channel) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Subscriber operations
	Subscribe(ctx context.Context, subscriber *Subscriber) error
	Unsubscribe(ctx context.Context, channelID, userID uuid.UUID) error
	FindSubscriber(ctx context.Context, channelID, userID uuid.UUID) (*Subscriber, error)
	IsSubscribed(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	UpdateSubscriberNotifications(ctx context.Context, channelID, userID uuid.UUID, enabled bool) error
	CountSubscribers(ctx context.Context, channelID uuid.UUID) (int, error)
	FindUserSubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Channel, error)

	// Admin operations
	AddAdmin(ctx context.Context, admin *Admin) error
	RemoveAdmin(ctx context.Context, channelID, userID uuid.UUID) error
	FindAdmin(ctx context.Context, channelID, userID uuid.UUID) (*Admin, error)
	FindAdmins(ctx context.Context, channelID uuid.UUID) ([]*Admin, error)
	IsAdmin(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	UpdateAdminPermissions(ctx context.Context, channelID, userID uuid.UUID, permissions *AdminPermissions) error
}
