package conversation

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a conversation between users
type Conversation struct {
	ID            uuid.UUID
	Type          Type
	LastMessageID *uuid.UUID
	LastMessageAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Type represents the type of conversation
type Type string

const (
	TypeDirect  Type = "direct"  // 1-to-1
	TypeGroup   Type = "group"   // Multiple users
	TypeChannel Type = "channel" // Broadcast channel
)

// Participant represents a user in a conversation
type Participant struct {
	ConversationID uuid.UUID
	UserID         uuid.UUID
	Role           Role
	JoinedAt       time.Time
	LastReadAt     *time.Time
	ArchivedAt     *time.Time // Timestamp when user archived this conversation
	IsOnline       bool       // Real-time online status (not persisted in DB)
}

// Role represents the role of a participant
type Role string

const (
	RoleMember    Role = "member"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleOwner     Role = "owner"
)

// NewDirectConversation creates a new direct (1-to-1) conversation
func NewDirectConversation() *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        uuid.New(),
		Type:      TypeDirect,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewGroupConversation creates a new group conversation
func NewGroupConversation() *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        uuid.New(),
		Type:      TypeGroup,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewChannelConversation creates a new channel conversation
func NewChannelConversation() *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        uuid.New(),
		Type:      TypeChannel,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateLastMessage updates the last message info
func (c *Conversation) UpdateLastMessage(messageID uuid.UUID) {
	now := time.Now()
	c.LastMessageID = &messageID
	c.LastMessageAt = &now
	c.UpdatedAt = now
}

// NewParticipant creates a new participant
func NewParticipant(conversationID, userID uuid.UUID, role Role) *Participant {
	return &Participant{
		ConversationID: conversationID,
		UserID:         userID,
		Role:           role,
		JoinedAt:       time.Now(),
	}
}

// UpdateLastRead updates the last read timestamp
func (p *Participant) UpdateLastRead() {
	now := time.Now()
	p.LastReadAt = &now
}
