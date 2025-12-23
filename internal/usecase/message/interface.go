package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the messaging use case interface
type Service interface {
	// SendMessage sends a message to a recipient
	SendMessage(ctx context.Context, senderID uuid.UUID, req *dto.SendMessageRequest) (*dto.SendMessageResponse, error)

	// GetMessages gets messages from a conversation
	GetMessages(ctx context.Context, userID uuid.UUID, req *dto.GetMessagesRequest) (*dto.GetMessagesResponse, error)

	// GetConversations gets user's conversations
	GetConversations(ctx context.Context, userID uuid.UUID, req *dto.GetConversationsRequest) (*dto.GetConversationsResponse, error)

	// CreateConversation creates a new conversation
	CreateConversation(ctx context.Context, userID uuid.UUID, req *dto.CreateConversationRequest) (*dto.CreateConversationResponse, error)

	// GetOrCreateDirectConversation gets or creates a direct conversation
	GetOrCreateDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (uuid.UUID, error)

	// MarkAsRead marks messages in a conversation as read
	MarkAsRead(ctx context.Context, userID uuid.UUID, req *dto.MarkAsReadRequest) error

	// DeleteMessage deletes a message
	DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) error

	// EditMessage edits a message content
	EditMessage(ctx context.Context, userID, messageID uuid.UUID, newContent string) (*dto.MessageDTO, error)

	// Message Reactions (Day 13)
	AddReaction(ctx context.Context, userID, messageID uuid.UUID, emoji string) error
	RemoveReaction(ctx context.Context, userID, messageID uuid.UUID, emoji string) error
	GetMessageReactions(ctx context.Context, messageID uuid.UUID) (*dto.MessageReactionsResponse, error)

	// Pinned Messages (Day 13)
	PinMessage(ctx context.Context, userID, conversationID, messageID uuid.UUID) error
	UnpinMessage(ctx context.Context, userID, conversationID, messageID uuid.UUID) error
	GetPinnedMessages(ctx context.Context, conversationID uuid.UUID) (*dto.PinnedMessagesResponse, error)

	// Message Forwarding (Day 13)
	ForwardMessage(ctx context.Context, userID, messageID, targetConversationID uuid.UUID) (*dto.SendMessageResponse, error)

	// Message Search (Day 13)
	SearchMessages(ctx context.Context, userID uuid.UUID, req *dto.SearchMessagesRequest) (*dto.SearchMessagesResponse, error)

	// Conversation Management
	ArchiveConversation(ctx context.Context, userID, conversationID uuid.UUID) error
	UnarchiveConversation(ctx context.Context, userID, conversationID uuid.UUID) error
	DeleteConversation(ctx context.Context, userID, conversationID uuid.UUID) error

	// Helper methods
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*dto.MessageDTO, error)
	GetConversationParticipants(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status string) error
}
