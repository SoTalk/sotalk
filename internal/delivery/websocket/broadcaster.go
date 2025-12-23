package websocket

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// Broadcaster provides methods to broadcast events to WebSocket clients
type Broadcaster struct {
	hub *Hub
}

// NewBroadcaster creates a new broadcaster instance
func NewBroadcaster(hub *Hub) *Broadcaster {
	return &Broadcaster{
		hub: hub,
	}
}

// BroadcastNewMessage broadcasts a new message event to conversation participants
func (b *Broadcaster) BroadcastNewMessage(ctx context.Context, conversationID uuid.UUID, msg dto.MessageDTO) error {
	event, err := NewEvent(EventMessageNew, MessagePayload{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		ContentType:    msg.ContentType,
		Status:         msg.Status,
		CreatedAt:      msg.CreatedAt,
		Reactions:      msg.Reactions,
		IsPinned:       msg.IsPinned,
		ReplyToID:      msg.ReplyToID,
	})
	if err != nil {
		logger.Error("Failed to create new message event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastMessageDeleted broadcasts a message deleted event
func (b *Broadcaster) BroadcastMessageDeleted(ctx context.Context, conversationID uuid.UUID, messageID string) error {
	event, err := NewEvent(EventMessageDeleted, map[string]interface{}{
		"message_id":      messageID,
		"conversation_id": conversationID.String(),
	})
	if err != nil {
		logger.Error("Failed to create deleted message event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastMessageUpdated broadcasts a message updated event
func (b *Broadcaster) BroadcastMessageUpdated(ctx context.Context, conversationID uuid.UUID, msg dto.MessageDTO) error {
	event, err := NewEvent(EventMessageUpdated, MessagePayload{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		ContentType:    msg.ContentType,
		Status:         msg.Status,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
		Reactions:      msg.Reactions,
		IsPinned:       msg.IsPinned,
		PinnedBy:       msg.PinnedBy,
		ReplyToID:      msg.ReplyToID,
	})
	if err != nil {
		logger.Error("Failed to create message updated event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastReactionAdded broadcasts a reaction added event
func (b *Broadcaster) BroadcastReactionAdded(ctx context.Context, conversationID uuid.UUID, messageID, userID, emoji string) error {
	event, err := NewEvent(EventReactionAdded, ReactionPayload{
		MessageID:      messageID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		Emoji:          emoji,
	})
	if err != nil {
		logger.Error("Failed to create reaction added event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastReactionRemoved broadcasts a reaction removed event
func (b *Broadcaster) BroadcastReactionRemoved(ctx context.Context, conversationID uuid.UUID, messageID, userID, emoji string) error {
	event, err := NewEvent(EventReactionRemoved, ReactionPayload{
		MessageID:      messageID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		Emoji:          emoji,
	})
	if err != nil {
		logger.Error("Failed to create reaction removed event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastMessagePinned broadcasts a message pinned event
func (b *Broadcaster) BroadcastMessagePinned(ctx context.Context, conversationID uuid.UUID, messageID, userID string) error {
	event, err := NewEvent(EventMessagePinned, PinPayload{
		MessageID:      messageID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		IsPinned:       true,
	})
	if err != nil {
		logger.Error("Failed to create message pinned event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastMessageUnpinned broadcasts a message unpinned event
func (b *Broadcaster) BroadcastMessageUnpinned(ctx context.Context, conversationID uuid.UUID, messageID, userID string) error {
	event, err := NewEvent(EventMessageUnpinned, PinPayload{
		MessageID:      messageID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		IsPinned:       false,
	})
	if err != nil {
		logger.Error("Failed to create message unpinned event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastPaymentRequest broadcasts a payment request event
func (b *Broadcaster) BroadcastPaymentRequest(ctx context.Context, toUserID uuid.UUID, payment dto.PaymentRequestDTO) error {
	event, err := NewEvent(EventPaymentRequest, PaymentPayload{
		ID:             payment.ID,
		ConversationID: payment.ConversationID,
		FromUserID:     payment.FromUserID,
		ToUserID:       payment.ToUserID,
		Amount:         payment.Amount,
		AmountSOL:      payment.AmountSOL,
		Status:         payment.Status,
		Message:        payment.Message,
		TransactionSig: payment.TransactionSig,
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
	})
	if err != nil {
		logger.Error("Failed to create payment request event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToUser(toUserID, event)
}

// BroadcastPaymentUpdate broadcasts a payment status update event
// eventType parameter is a string for compatibility with payment service interface
func (b *Broadcaster) BroadcastPaymentUpdate(ctx context.Context, userID uuid.UUID, eventTypeStr string, payment dto.PaymentRequestDTO) error {
	// Map string event type to EventType
	var eventType EventType
	switch eventTypeStr {
	case "payment.accepted":
		eventType = EventPaymentAccepted
	case "payment.rejected":
		eventType = EventPaymentRejected
	case "payment.canceled":
		eventType = EventPaymentCanceled
	case "payment.confirmed":
		eventType = EventPaymentConfirmed
	default:
		logger.Warn("Unknown payment event type", zap.String("type", eventTypeStr))
		eventType = EventType(eventTypeStr)
	}

	event, err := NewEvent(eventType, PaymentPayload{
		ID:             payment.ID,
		ConversationID: payment.ConversationID,
		FromUserID:     payment.FromUserID,
		ToUserID:       payment.ToUserID,
		Amount:         payment.Amount,
		AmountSOL:      payment.AmountSOL,
		Status:         payment.Status,
		Message:        payment.Message,
		TransactionSig: payment.TransactionSig,
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
	})
	if err != nil {
		logger.Error("Failed to create payment update event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToUser(userID, event)
}

// BroadcastConversationUpdated broadcasts a conversation update event
func (b *Broadcaster) BroadcastConversationUpdated(ctx context.Context, userIDs []uuid.UUID, conversationID string, unreadCount int, lastMessageAt time.Time) error {
	event, err := NewEvent(EventConversationUpdated, ConversationPayload{
		ID:            conversationID,
		UnreadCount:   unreadCount,
		LastMessageAt: lastMessageAt,
	})
	if err != nil {
		logger.Error("Failed to create conversation updated event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToUsers(userIDs, event)
}

// BroadcastGroupCreated broadcasts a group created event to all members
func (b *Broadcaster) BroadcastGroupCreated(ctx context.Context, conversationID uuid.UUID, group dto.GroupDTO) error {
	event, err := NewEvent(EventGroupCreated, GroupPayload{
		ID:             group.ID,
		ConversationID: group.ConversationID,
		Name:           group.Name,
		Description:    group.Description,
		Avatar:         group.Avatar,
		CreatorID:      group.CreatorID,
		MemberCount:    group.MemberCount,
		Settings: map[string]interface{}{
			"who_can_message":     group.Settings.WhoCanMessage,
			"who_can_add_members": group.Settings.WhoCanAddMembers,
			"who_can_edit_info":   group.Settings.WhoCanEditInfo,
		},
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	})
	if err != nil {
		logger.Error("Failed to create group created event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastGroupUpdated broadcasts a group updated event to all members
func (b *Broadcaster) BroadcastGroupUpdated(ctx context.Context, conversationID uuid.UUID, group dto.GroupDTO) error {
	event, err := NewEvent(EventGroupUpdated, GroupPayload{
		ID:             group.ID,
		ConversationID: group.ConversationID,
		Name:           group.Name,
		Description:    group.Description,
		Avatar:         group.Avatar,
		CreatorID:      group.CreatorID,
		MemberCount:    group.MemberCount,
		Settings: map[string]interface{}{
			"who_can_message":     group.Settings.WhoCanMessage,
			"who_can_add_members": group.Settings.WhoCanAddMembers,
			"who_can_edit_info":   group.Settings.WhoCanEditInfo,
		},
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	})
	if err != nil {
		logger.Error("Failed to create group updated event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastGroupDeleted broadcasts a group deleted event to all members
func (b *Broadcaster) BroadcastGroupDeleted(ctx context.Context, conversationID uuid.UUID, groupID, deletedBy string) error {
	event, err := NewEvent(EventGroupDeleted, map[string]interface{}{
		"group_id":        groupID,
		"conversation_id": conversationID.String(),
		"deleted_by":      deletedBy,
	})
	if err != nil {
		logger.Error("Failed to create group deleted event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastGroupSettingsUpdated broadcasts a group settings updated event
func (b *Broadcaster) BroadcastGroupSettingsUpdated(ctx context.Context, conversationID uuid.UUID, groupID, updatedBy string, settings dto.GroupSettings) error {
	event, err := NewEvent(EventGroupSettingsUpdated, GroupSettingsPayload{
		GroupID:        groupID,
		ConversationID: conversationID.String(),
		Settings: map[string]interface{}{
			"who_can_message":     settings.WhoCanMessage,
			"who_can_add_members": settings.WhoCanAddMembers,
			"who_can_edit_info":   settings.WhoCanEditInfo,
		},
		UpdatedBy: updatedBy,
	})
	if err != nil {
		logger.Error("Failed to create group settings updated event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastGroupMemberJoined broadcasts a member joined event
func (b *Broadcaster) BroadcastGroupMemberJoined(ctx context.Context, conversationID uuid.UUID, groupID, userID, username, role, addedBy string) error {
	event, err := NewEvent(EventGroupMemberJoined, GroupMemberPayload{
		GroupID:        groupID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		Username:       username,
		Role:           role,
		ActionBy:       addedBy,
		JoinedAt:       time.Now(),
	})
	if err != nil {
		logger.Error("Failed to create group member joined event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastGroupMemberLeft broadcasts a member left event
func (b *Broadcaster) BroadcastGroupMemberLeft(ctx context.Context, conversationID uuid.UUID, groupID, userID, username string) error {
	event, err := NewEvent(EventGroupMemberLeft, GroupMemberPayload{
		GroupID:        groupID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		Username:       username,
		ActionBy:       userID, // User removed themselves
	})
	if err != nil {
		logger.Error("Failed to create group member left event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastGroupMemberRemoved broadcasts a member removed event
func (b *Broadcaster) BroadcastGroupMemberRemoved(ctx context.Context, conversationID uuid.UUID, groupID, userID, username, removedBy string) error {
	event, err := NewEvent(EventGroupMemberRemoved, GroupMemberPayload{
		GroupID:        groupID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		Username:       username,
		ActionBy:       removedBy,
	})
	if err != nil {
		logger.Error("Failed to create group member removed event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}

// BroadcastGroupMemberRoleChanged broadcasts a member role changed event
func (b *Broadcaster) BroadcastGroupMemberRoleChanged(ctx context.Context, conversationID uuid.UUID, groupID, userID, username, previousRole, newRole, changedBy string) error {
	event, err := NewEvent(EventGroupMemberRoleChanged, GroupMemberPayload{
		GroupID:        groupID,
		ConversationID: conversationID.String(),
		UserID:         userID,
		Username:       username,
		Role:           newRole,
		PreviousRole:   previousRole,
		ActionBy:       changedBy,
	})
	if err != nil {
		logger.Error("Failed to create group member role changed event", zap.Error(err))
		return err
	}

	return b.hub.BroadcastToConversation(ctx, conversationID, event)
}
