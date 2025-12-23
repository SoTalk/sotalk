package request

import "github.com/google/uuid"

// SendMessageRequest is the HTTP request for sending a message
type SendMessageRequest struct {
	RecipientID string `json:"recipient_id" binding:"required"`
	Content     string `json:"content" binding:"required"` // Accept as string, convert to []byte in handler
	ContentType string `json:"content_type" binding:"required"`
	Signature   string `json:"signature"`
	ReplyToID   *string `json:"reply_to_id,omitempty"`
}

// GetMessagesRequest is the HTTP request for getting messages
type GetMessagesRequest struct {
	ConversationID string `form:"conversation_id" binding:"required"`
	Limit          int    `form:"limit"`
	Offset         int    `form:"offset"`
}

// GetConversationsRequest is the HTTP request for getting conversations
type GetConversationsRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// MarkAsReadRequest is the HTTP request for marking messages as read
type MarkAsReadRequest struct {
	ConversationID string `json:"conversation_id" binding:"required"`
}

// EditMessageRequest is the HTTP request for editing a message
type EditMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// ParseRecipientID parses recipient ID from string to UUID
func (r *SendMessageRequest) ParseRecipientID() (uuid.UUID, error) {
	return uuid.Parse(r.RecipientID)
}

// ParseReplyToID parses reply_to_id from string to UUID
func (r *SendMessageRequest) ParseReplyToID() (*uuid.UUID, error) {
	if r.ReplyToID == nil {
		return nil, nil
	}
	id, err := uuid.Parse(*r.ReplyToID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// ParseConversationID parses conversation_id from string to UUID
func (r *GetMessagesRequest) ParseConversationID() (uuid.UUID, error) {
	return uuid.Parse(r.ConversationID)
}

// ParseConversationID parses conversation_id from string to UUID
func (r *MarkAsReadRequest) ParseConversationID() (uuid.UUID, error) {
	return uuid.Parse(r.ConversationID)
}

// CreateConversationRequest is the HTTP request for creating a conversation
type CreateConversationRequest struct {
	ParticipantID string  `json:"participant_id" binding:"required"`
	Type          *string `json:"type,omitempty"` // direct, group, channel
}

// ParseParticipantID parses participant_id from string to UUID
func (r *CreateConversationRequest) ParseParticipantID() (uuid.UUID, error) {
	return uuid.Parse(r.ParticipantID)
}
