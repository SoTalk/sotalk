package message

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for message data operations
type Repository interface {
	// Basic message operations
	Create(ctx context.Context, message *Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	// FindByConversationID finds messages. userID is optional - if provided, isPinned only true for messages pinned by that user
	FindByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]*Message, error)
	Update(ctx context.Context, message *Message) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, messageID uuid.UUID, status Status) error
	CountByConversationID(ctx context.Context, conversationID uuid.UUID) (int64, error)
	CountUnreadByConversationID(ctx context.Context, conversationID, userID uuid.UUID) (int64, error)

	// Message Reactions (Day 13)
	AddReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	GetMessageReactions(ctx context.Context, messageID uuid.UUID) ([]MessageReaction, error)

	// Pinned Messages (Day 13)
	PinMessage(ctx context.Context, conversationID, messageID, pinnedBy uuid.UUID) error
	UnpinMessage(ctx context.Context, conversationID, messageID, userID uuid.UUID) error // userID: only unpin for this user
	GetPinnedMessages(ctx context.Context, conversationID uuid.UUID) ([]PinnedMessage, error)
	IsPinned(ctx context.Context, messageID uuid.UUID) (bool, error)

	// Message Forwarding (Day 13)
	RecordForward(ctx context.Context, originalID, newID, forwardedBy, targetConversationID uuid.UUID) error
	GetForwardHistory(ctx context.Context, messageID uuid.UUID) ([]ForwardedMessage, error)

	// Search (Day 13)
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, conversationID *uuid.UUID, limit, offset int) ([]*Message, error)
}
