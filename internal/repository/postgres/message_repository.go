package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/message"
	"gorm.io/gorm"
)

// messageRepository implements message.Repository interface
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) message.Repository {
	return &messageRepository{db: db}
}

// Create creates a new message
func (r *messageRepository) Create(ctx context.Context, m *message.Message) error {
	dbMessage := toMessageModel(m)
	result := r.db.WithContext(ctx).Create(dbMessage)
	if result.Error != nil {
		return result.Error
	}

	// Update domain entity with generated values
	m.ID = dbMessage.ID
	m.CreatedAt = dbMessage.CreatedAt
	m.UpdatedAt = dbMessage.UpdatedAt

	return nil
}

// FindByID finds a message by ID
func (r *messageRepository) FindByID(ctx context.Context, id uuid.UUID) (*message.Message, error) {
	var dbMessage Message
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbMessage)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, message.ErrMessageNotFound
		}
		return nil, result.Error
	}

	return toDomainMessage(&dbMessage), nil
}

// FindByConversationID finds messages by conversation ID with pagination
// userID is optional - if provided, only returns messages pinned by that user
func (r *messageRepository) FindByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]*message.Message, error) {
	var dbMessages []Message

	result := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbMessages)

	if result.Error != nil {
		return nil, result.Error
	}

	messages := make([]*message.Message, len(dbMessages))
	for i, dbMsg := range dbMessages {
		domainMsg := toDomainMessage(&dbMsg)

		// Fetch reactions for this message
		reactions, _ := r.getMessageReactions(ctx, dbMsg.ID)
		domainMsg.Reactions = reactions

		// Get list of users who pinned this message
		pinnedBy, _ := r.getMessagePinnedBy(ctx, dbMsg.ID, conversationID)
		domainMsg.PinnedBy = pinnedBy

		// Check if message is pinned (by specific user if userID provided)
		isPinned, _ := r.isMessagePinned(ctx, dbMsg.ID, conversationID, userID)
		domainMsg.IsPinned = isPinned

		messages[i] = domainMsg
	}

	return messages, nil
}

// Update updates a message
func (r *messageRepository) Update(ctx context.Context, m *message.Message) error {
	dbMessage := toMessageModel(m)
	result := r.db.WithContext(ctx).Model(&Message{}).Where("id = ?", m.ID).Updates(dbMessage)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return message.ErrMessageNotFound
	}

	return nil
}

// Delete soft deletes a message
func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Message{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return message.ErrMessageNotFound
	}

	return nil
}

// UpdateStatus updates message status
func (r *messageRepository) UpdateStatus(ctx context.Context, messageID uuid.UUID, status message.Status) error {
	result := r.db.WithContext(ctx).Model(&Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"status":     string(status),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return message.ErrMessageNotFound
	}

	return nil
}

// CountByConversationID counts messages in a conversation
func (r *messageRepository) CountByConversationID(ctx context.Context, conversationID uuid.UUID) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&Message{}).Where("conversation_id = ?", conversationID).Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountUnreadByConversationID counts unread messages in a conversation for a specific user
func (r *messageRepository) CountUnreadByConversationID(ctx context.Context, conversationID, userID uuid.UUID) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&Message{}).
		Where("conversation_id = ? AND sender_id != ? AND status != ?", conversationID, userID, string(message.StatusRead)).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// Message Reactions (Day 13)

func (r *messageRepository) AddReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	// Check if reaction already exists
	var existing MessageReaction
	err := r.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		First(&existing).Error

	if err == nil {
		// Already reacted with same emoji
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := &MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *messageRepository) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	result := r.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&MessageReaction{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("reaction not found")
	}

	return nil
}

func (r *messageRepository) GetMessageReactions(ctx context.Context, messageID uuid.UUID) ([]message.MessageReaction, error) {
	var models []MessageReaction
	if err := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]message.MessageReaction, len(models))
	for i, m := range models {
		result[i] = message.MessageReaction{
			ID:        m.ID,
			MessageID: m.MessageID,
			UserID:    m.UserID,
			Emoji:     m.Emoji,
			CreatedAt: m.CreatedAt,
		}
	}

	return result, nil
}

// Pinned Messages (Day 13)

func (r *messageRepository) PinMessage(ctx context.Context, conversationID, messageID, pinnedBy uuid.UUID) error {
	// Check if already pinned
	var existing PinnedMessage
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND message_id = ?", conversationID, messageID).
		First(&existing).Error

	if err == nil {
		// Already pinned
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := &PinnedMessage{
		ConversationID: conversationID,
		MessageID:      messageID,
		PinnedBy:       pinnedBy,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *messageRepository) UnpinMessage(ctx context.Context, conversationID, messageID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("conversation_id = ? AND message_id = ? AND pinned_by = ?", conversationID, messageID, userID).
		Delete(&PinnedMessage{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("pinned message not found")
	}

	return nil
}

func (r *messageRepository) GetPinnedMessages(ctx context.Context, conversationID uuid.UUID) ([]message.PinnedMessage, error) {
	var models []PinnedMessage
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("pinned_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]message.PinnedMessage, len(models))
	for i, m := range models {
		result[i] = message.PinnedMessage{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			MessageID:      m.MessageID,
			PinnedBy:       m.PinnedBy,
			PinnedAt:       m.PinnedAt,
		}
	}

	return result, nil
}

func (r *messageRepository) IsPinned(ctx context.Context, messageID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&PinnedMessage{}).
		Where("message_id = ?", messageID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Message Forwarding (Day 13)

func (r *messageRepository) RecordForward(ctx context.Context, originalID, newID, forwardedBy, targetConversationID uuid.UUID) error {
	model := &ForwardedMessage{
		OriginalMessageID:    originalID,
		NewMessageID:         newID,
		ForwardedBy:          forwardedBy,
		TargetConversationID: targetConversationID,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *messageRepository) GetForwardHistory(ctx context.Context, messageID uuid.UUID) ([]message.ForwardedMessage, error) {
	var models []ForwardedMessage
	if err := r.db.WithContext(ctx).
		Where("original_message_id = ?", messageID).
		Order("forwarded_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]message.ForwardedMessage, len(models))
	for i, m := range models {
		result[i] = message.ForwardedMessage{
			ID:                   m.ID,
			OriginalMessageID:    m.OriginalMessageID,
			NewMessageID:         m.NewMessageID,
			ForwardedBy:          m.ForwardedBy,
			TargetConversationID: m.TargetConversationID,
			ForwardedAt:          m.ForwardedAt,
		}
	}

	return result, nil
}

// Search Messages (Day 13)

func (r *messageRepository) SearchMessages(ctx context.Context, userID uuid.UUID, query string, conversationID *uuid.UUID, limit, offset int) ([]*message.Message, error) {
	var models []Message

	// Basic text search using ILIKE
	searchQuery := "%" + query + "%"
	q := r.db.WithContext(ctx).
		Table("messages").
		Select("messages.*").
		Joins("INNER JOIN conversation_participants ON messages.conversation_id = conversation_participants.conversation_id").
		Where("conversation_participants.user_id = ?", userID).
		Where("CAST(messages.content AS TEXT) ILIKE ?", searchQuery)

	if conversationID != nil {
		q = q.Where("messages.conversation_id = ?", *conversationID)
	}

	// Only show messages from conversations user is part of
	// Implemented via INNER JOIN with conversation_participants table above

	q = q.Order("messages.created_at DESC")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if offset > 0 {
		q = q.Offset(offset)
	}

	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*message.Message, len(models))
	for i, m := range models {
		result[i] = toDomainMessage(&m)
	}

	return result, nil
}

// getMessageReactions fetches reactions for a message
func (r *messageRepository) getMessageReactions(ctx context.Context, messageID uuid.UUID) (map[string][]string, error) {
	var reactions []struct {
		Emoji  string
		UserID string
	}

	result := r.db.WithContext(ctx).
		Table("message_reactions").
		Select("emoji, user_id::text").
		Where("message_id = ?", messageID).
		Scan(&reactions)

	if result.Error != nil {
		return nil, result.Error
	}

	// Group reactions by emoji
	reactionMap := make(map[string][]string)
	for _, r := range reactions {
		reactionMap[r.Emoji] = append(reactionMap[r.Emoji], r.UserID)
	}

	if len(reactionMap) == 0 {
		return nil, nil
	}

	return reactionMap, nil
}

// getMessagePinnedBy returns list of user IDs who pinned a message
func (r *messageRepository) getMessagePinnedBy(ctx context.Context, messageID, conversationID uuid.UUID) ([]string, error) {
	var userIDs []string

	err := r.db.WithContext(ctx).
		Table("pinned_messages").
		Where("message_id = ? AND conversation_id = ?", messageID, conversationID).
		Pluck("pinned_by::text", &userIDs).Error

	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

// isMessagePinned checks if a message is pinned by a specific user in a conversation
func (r *messageRepository) isMessagePinned(ctx context.Context, messageID, conversationID uuid.UUID, userID *uuid.UUID) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Table("pinned_messages").
		Where("message_id = ? AND conversation_id = ?", messageID, conversationID)

	// If userID is provided, check if this specific user pinned it
	if userID != nil {
		query = query.Where("pinned_by = ?", *userID)
	}

	result := query.Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// Mapper functions

// toMessageModel converts domain Message to GORM Message model
func toMessageModel(m *message.Message) *Message {
	return &Message{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Content:        m.Content,
		ContentType:    string(m.ContentType),
		Signature:      m.Signature,
		ReplyToID:      m.ReplyToID,
		Status:         string(m.Status),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// toDomainMessage converts GORM Message model to domain Message
func toDomainMessage(m *Message) *message.Message {
	return &message.Message{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Content:        m.Content,
		ContentType:    message.ContentType(m.ContentType),
		Signature:      m.Signature,
		ReplyToID:      m.ReplyToID,
		Status:         message.Status(m.Status),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
