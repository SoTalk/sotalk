package dto

import (
	"time"

	"github.com/google/uuid"
)

// SendMessageRequest is the request for sending a message
type SendMessageRequest struct {
	RecipientID uuid.UUID `json:"recipient_id" validate:"required"`
	Content     string    `json:"content" validate:"required"`
	ContentType string    `json:"content_type" validate:"required"`
	Signature   string    `json:"signature"`
	ReplyToID   *uuid.UUID `json:"reply_to_id,omitempty"`
}

// SendMessageResponse is the response for sending a message
type SendMessageResponse struct {
	Message MessageDTO `json:"message"`
}

// GetMessagesRequest is the request for getting messages
type GetMessagesRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
	Limit          int       `json:"limit"`
	Offset         int       `json:"offset"`
}

// GetMessagesResponse is the response for getting messages
type GetMessagesResponse struct {
	Messages []MessageDTO `json:"messages"`
	Total    int64        `json:"total"`
}

// GetConversationsRequest is the request for getting conversations
type GetConversationsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// GetConversationsResponse is the response for getting conversations
type GetConversationsResponse struct {
	Conversations []ConversationDTO `json:"conversations"`
}

// MarkAsReadRequest is the request for marking messages as read
type MarkAsReadRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
}

// CreateConversationRequest is the request for creating a conversation
type CreateConversationRequest struct {
	ParticipantID string `json:"participant_id" validate:"required"` // Accepts UUID or wallet address
	Type          string `json:"type"`                               // direct, group, channel
}

// CreateConversationResponse is the response for creating a conversation
type CreateConversationResponse struct {
	Conversation ConversationDTO `json:"conversation"`
}

// MessageDTO is the data transfer object for message
type MessageDTO struct {
	ID             string              `json:"id"`
	ConversationID string              `json:"conversation_id"`
	SenderID       string              `json:"sender_id"`
	Content        string              `json:"content"`
	ContentType    string              `json:"content_type"`
	Signature      string              `json:"signature,omitempty"`
	ReplyToID      *string             `json:"reply_to_id,omitempty"`
	Status         string              `json:"status"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Sender         *UserDTO            `json:"sender,omitempty"`
	Reactions      map[string][]string `json:"reactions,omitempty"`
	IsPinned       bool                `json:"is_pinned"`           // Deprecated: use PinnedBy instead
	PinnedBy       []string            `json:"pinned_by,omitempty"` // List of user IDs who pinned this message
}

// ConversationDTO is the data transfer object for conversation
type ConversationDTO struct {
	ID            string     `json:"id"`
	Type          string     `json:"type"`
	Participants  []UserDTO  `json:"participants"`
	LastMessage   *MessageDTO `json:"last_message,omitempty"`
	UnreadCount   int        `json:"unread_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
