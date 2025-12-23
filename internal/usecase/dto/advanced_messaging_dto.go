package dto

import "time"

// Message Reactions

type AddReactionRequest struct {
	MessageID string `json:"message_id" binding:"required"`
	Emoji     string `json:"emoji" binding:"required"`
}

type ReactionResponse struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// Pinned Messages

type PinMessageRequest struct {
	MessageID      string `json:"message_id" binding:"required"`
	ConversationID string `json:"conversation_id" binding:"required"`
}

type PinnedMessageResponse struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	MessageID      string    `json:"message_id"`
	PinnedBy       string    `json:"pinned_by"`
	PinnedAt       time.Time `json:"pinned_at"`
}

// Message Forwarding

type ForwardMessageRequest struct {
	MessageID        string `json:"message_id" binding:"required"`
	ConversationID   string `json:"conversation_id" binding:"required"`
}

// Status/Stories

type CreateStatusRequest struct {
	MediaID string  `json:"media_id" binding:"required"`
	Caption *string `json:"caption,omitempty"`
	Privacy *string `json:"privacy,omitempty"` // "everyone", "contacts", "private"
}

type StatusResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	MediaID   string    `json:"media_id"`
	Caption   string    `json:"caption"`
	Privacy   string    `json:"privacy"`
	ViewCount int       `json:"view_count"`
	HasViewed bool      `json:"has_viewed"` // If current user has viewed
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type StatusViewResponse struct {
	ViewerID  string    `json:"viewer_id"`
	ViewedAt  time.Time `json:"viewed_at"`
}

// Contacts

type AddContactRequest struct {
	ContactID   string  `json:"contact_id" binding:"required"`
	DisplayName *string `json:"display_name,omitempty"`
}

type ContactResponse struct {
	ID          string    `json:"id"`
	ContactID   string    `json:"contact_id"`
	DisplayName string    `json:"display_name"`
	IsFavorite  bool      `json:"is_favorite"`
	AddedAt     time.Time `json:"added_at"`
}

type SendInviteRequest struct {
	RecipientID string  `json:"recipient_id" binding:"required"`
	Message     *string `json:"message,omitempty"`
}

type InviteResponse struct {
	ID          string     `json:"id"`
	SenderID    string     `json:"sender_id"`
	RecipientID string     `json:"recipient_id"`
	Message     string     `json:"message"`
	Status      string     `json:"status"`
	ExpiresAt   time.Time  `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// Message Search

type SearchMessagesRequest struct {
	Query          string  `json:"query" binding:"required"`
	ConversationID *string `json:"conversation_id,omitempty"`
	Limit          int     `json:"limit,omitempty"`
	Offset         int     `json:"offset,omitempty"`
}

type SearchMessagesResponse struct {
	Messages []*MessageDTO `json:"messages"`
	Total    int           `json:"total"`
}

type MessageResponse struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversation_id"`
	SenderID       string     `json:"sender_id"`
	Content        []byte     `json:"content"`
	ContentType    string     `json:"content_type"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	Sender         *UserDTO   `json:"sender,omitempty"`
}

type MessageReactionsResponse struct {
	MessageID string              `json:"message_id"`
	Reactions []ReactionResponse  `json:"reactions"`
}

type PinnedMessagesResponse struct {
	ConversationID string                  `json:"conversation_id"`
	PinnedMessages []PinnedMessageResponse `json:"pinned_messages"`
}
