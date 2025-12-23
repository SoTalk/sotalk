package websocket

import (
	"encoding/json"
	"time"

)

// EventType represents the type of WebSocket event
type EventType string

const (
	// Message events
	EventMessageNew       EventType = "message.new"
	EventMessageDelivered EventType = "message.delivered"
	EventMessageRead      EventType = "message.read"
	EventMessageDeleted   EventType = "message.deleted"
	EventMessageUpdated   EventType = "message.updated"

	// Reaction events
	EventReactionAdded   EventType = "reaction.added"
	EventReactionRemoved EventType = "reaction.removed"

	// Pin events
	EventMessagePinned   EventType = "message.pinned"
	EventMessageUnpinned EventType = "message.unpinned"

	// Typing events
	EventTypingStart EventType = "typing.start"
	EventTypingStop  EventType = "typing.stop"

	// Conversation events
	EventConversationUpdated EventType = "conversation.updated"

	// Group events
	EventGroupCreated            EventType = "group.created"
	EventGroupUpdated            EventType = "group.updated"
	EventGroupDeleted            EventType = "group.deleted"
	EventGroupSettingsUpdated    EventType = "group.settings_updated"
	EventGroupMemberJoined       EventType = "group.member_joined"
	EventGroupMemberLeft         EventType = "group.member_left"
	EventGroupMemberRemoved      EventType = "group.member_removed"
	EventGroupMemberRoleChanged  EventType = "group.member_role_changed"

	// Payment events
	EventPaymentRequest  EventType = "payment.request"
	EventPaymentAccepted EventType = "payment.accepted"
	EventPaymentRejected EventType = "payment.rejected"
	EventPaymentCanceled EventType = "payment.canceled"
	EventPaymentConfirmed EventType = "payment.confirmed"

	// Presence events
	EventUserOnline  EventType = "user.online"
	EventUserOffline EventType = "user.offline"

	// System events
	EventError EventType = "error"
	EventPing  EventType = "ping"
	EventPong  EventType = "pong"
)

// Event represents a WebSocket event
type Event struct {
	Type      EventType       `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewEvent creates a new event with the given type and payload
func NewEvent(eventType EventType, payload interface{}) (*Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Event{
		Type:      eventType,
		Payload:   data,
		Timestamp: time.Now(),
	}, nil
}

// Payloads for different event types

// MessagePayload for message events
type MessagePayload struct {
	ID             string                       `json:"id"`
	ConversationID string                       `json:"conversation_id"`
	SenderID       string                       `json:"sender_id"`
	Content        string                       `json:"content"`
	ContentType    string                       `json:"content_type"`
	Status         string                       `json:"status"`
	CreatedAt      time.Time                    `json:"created_at"`
	UpdatedAt      time.Time                    `json:"updated_at"`
	Reactions      map[string][]string          `json:"reactions,omitempty"`
	IsPinned       bool                         `json:"is_pinned"`
	PinnedBy       []string                     `json:"pinned_by,omitempty"`
	ReplyToID      *string                      `json:"reply_to_id,omitempty"`
}

// MessageStatusPayload for delivery/read receipts
type MessageStatusPayload struct {
	MessageID      string    `json:"message_id"`
	ConversationID string    `json:"conversation_id"`
	UserID         string    `json:"user_id"`
	Status         string    `json:"status"` // "delivered" or "read"
	Timestamp      time.Time `json:"timestamp"`
}

// ReactionPayload for reaction events
type ReactionPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Emoji          string `json:"emoji"`
}

// PinPayload for pin/unpin events
type PinPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	IsPinned       bool   `json:"is_pinned"`
}

// TypingPayload for typing indicator events
type TypingPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Username       string `json:"username,omitempty"`
	IsTyping       bool   `json:"is_typing"`
}

// ConversationPayload for conversation update events
type ConversationPayload struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	UnreadCount    int       `json:"unread_count"`
	LastMessageAt  time.Time `json:"last_message_at"`
}

// PaymentPayload for payment events
type PaymentPayload struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	FromUserID     string    `json:"from_user_id"`
	ToUserID       string    `json:"to_user_id"`
	Amount         uint64    `json:"amount"`
	AmountSOL      float64   `json:"amount_sol"`
	Status         string    `json:"status"`
	Message        string    `json:"message,omitempty"`
	TransactionSig *string   `json:"transaction_sig,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ErrorPayload for error events
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ClientMessage represents messages sent from client to server
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Client message types
const (
	ClientMessageTyping         = "typing"
	ClientMessageStopTyping     = "stop_typing"
	ClientMessageDelivered      = "message_delivered"
	ClientMessageRead           = "message_read"
	ClientMessagePing           = "ping"
)

// ClientTypingPayload for client typing events
type ClientTypingPayload struct {
	ConversationID string `json:"conversation_id"`
}

// ClientMessageStatusPayload for client delivery/read receipts
type ClientMessageStatusPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
}

// GroupPayload for group events
type GroupPayload struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Avatar         string                 `json:"avatar,omitempty"`
	CreatorID      string                 `json:"creator_id"`
	MemberCount    int                    `json:"member_count"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// GroupMemberPayload for group member events
type GroupMemberPayload struct {
	GroupID        string                 `json:"group_id"`
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	Username       string                 `json:"username,omitempty"`
	Role           string                 `json:"role,omitempty"`
	PreviousRole   string                 `json:"previous_role,omitempty"`
	Permissions    map[string]interface{} `json:"permissions,omitempty"`
	ActionBy       string                 `json:"action_by,omitempty"` // Who performed the action
	JoinedAt       time.Time              `json:"joined_at,omitempty"`
}

// GroupSettingsPayload for group settings update events
type GroupSettingsPayload struct {
	GroupID        string                 `json:"group_id"`
	ConversationID string                 `json:"conversation_id"`
	Settings       map[string]interface{} `json:"settings"`
	UpdatedBy      string                 `json:"updated_by"`
}
