package response

import "time"

// SendMessageResponse is the HTTP response for sending a message
type SendMessageResponse struct {
	Message MessageDTO `json:"message"`
}

// GetMessagesResponse is the HTTP response for getting messages
type GetMessagesResponse struct {
	Messages []MessageDTO `json:"messages"`
	Total    int64        `json:"total"`
}

// GetConversationsResponse is the HTTP response for getting conversations
type GetConversationsResponse struct {
	Conversations []ConversationDTO `json:"conversations"`
}

// ConversationResponse is the HTTP response for a single conversation
type ConversationResponse struct {
	Conversation ConversationDTO `json:"conversation"`
}

// MessageDTO is the message data in response
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

// ConversationDTO is the conversation data in response
type ConversationDTO struct {
	ID            string        `json:"id"`
	Type          string        `json:"type"`
	Participants  []UserDTO     `json:"participants"`
	LastMessage   *MessageDTO   `json:"last_message,omitempty"`
	UnreadCount   int           `json:"unread_count"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
