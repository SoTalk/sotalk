package message

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message in the system
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Content        string // Plain text content
	ContentType    ContentType
	Signature      string // Solana signature
	ReplyToID      *uuid.UUID
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Reactions      map[string][]string // emoji -> list of user IDs
	IsPinned       bool                // Deprecated: use PinnedBy instead. True if current user pinned it.
	PinnedBy       []string            // List of user IDs who pinned this message
}

// ContentType represents the type of message content
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeVideo ContentType = "video"
	ContentTypeAudio ContentType = "audio"
	ContentTypeFile  ContentType = "file"
)

// Status represents message delivery status
type Status string

const (
	StatusSending   Status = "sending"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
	StatusFailed    Status = "failed"
)

// NewMessage creates a new message
func NewMessage(conversationID, senderID uuid.UUID, content string, contentType ContentType) *Message {
	return &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		ContentType:    contentType,
		Status:         StatusSending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// SetSignature sets the Solana signature
func (m *Message) SetSignature(signature string) {
	m.Signature = signature
}

// SetReplyTo sets the message this is replying to
func (m *Message) SetReplyTo(messageID uuid.UUID) {
	m.ReplyToID = &messageID
}

// MarkAsSent marks the message as sent
func (m *Message) MarkAsSent() {
	m.Status = StatusSent
	m.UpdatedAt = time.Now()
}

// MarkAsDelivered marks the message as delivered
func (m *Message) MarkAsDelivered() {
	m.Status = StatusDelivered
	m.UpdatedAt = time.Now()
}

// MarkAsRead marks the message as read
func (m *Message) MarkAsRead() {
	m.Status = StatusRead
	m.UpdatedAt = time.Now()
}

// MarkAsFailed marks the message as failed
func (m *Message) MarkAsFailed() {
	m.Status = StatusFailed
	m.UpdatedAt = time.Now()
}

// MessageReaction represents a reaction to a message (Day 13)
type MessageReaction struct {
	ID        uuid.UUID
	MessageID uuid.UUID
	UserID    uuid.UUID
	Emoji     string
	CreatedAt time.Time
}

// PinnedMessage represents a pinned message in a conversation (Day 13)
type PinnedMessage struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	MessageID      uuid.UUID
	PinnedBy       uuid.UUID
	PinnedAt       time.Time
}

// ForwardedMessage tracks message forwarding (Day 13)
type ForwardedMessage struct {
	ID                   uuid.UUID
	OriginalMessageID    uuid.UUID
	NewMessageID         uuid.UUID
	ForwardedBy          uuid.UUID
	TargetConversationID uuid.UUID
	ForwardedAt          time.Time
}
