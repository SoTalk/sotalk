package conversation

import (
	"context"

	"github.com/google/uuid"
)

// PresenceChecker defines the interface for checking user online status
type PresenceChecker interface {
	IsUserOnline(userID uuid.UUID) bool
}

// Repository defines the interface for conversation data operations
type Repository interface {
	Create(ctx context.Context, conversation *Conversation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Conversation, error)
	FindDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*Conversation, error)
	Update(ctx context.Context, conversation *Conversation) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Participant operations
	AddParticipant(ctx context.Context, participant *Participant) error
	RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error
	FindParticipants(ctx context.Context, conversationID uuid.UUID) ([]*Participant, error)
	IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	UpdateParticipantLastRead(ctx context.Context, conversationID, userID uuid.UUID) error

	// Archive operations
	ArchiveConversation(ctx context.Context, conversationID, userID uuid.UUID) error
	UnarchiveConversation(ctx context.Context, conversationID, userID uuid.UUID) error
}
