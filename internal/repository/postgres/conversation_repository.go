package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/conversation"
	"gorm.io/gorm"
)

// conversationRepository implements conversation.Repository interface
type conversationRepository struct {
	db              *gorm.DB
	presenceChecker conversation.PresenceChecker
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB, presenceChecker conversation.PresenceChecker) conversation.Repository {
	return &conversationRepository{
		db:              db,
		presenceChecker: presenceChecker,
	}
}

// Create creates a new conversation
func (r *conversationRepository) Create(ctx context.Context, c *conversation.Conversation) error {
	dbConversation := toConversationModel(c)
	result := r.db.WithContext(ctx).Create(dbConversation)
	if result.Error != nil {
		return result.Error
	}

	// Update domain entity with generated values
	c.ID = dbConversation.ID
	c.CreatedAt = dbConversation.CreatedAt
	c.UpdatedAt = dbConversation.UpdatedAt

	return nil
}

// FindByID finds a conversation by ID
func (r *conversationRepository) FindByID(ctx context.Context, id uuid.UUID) (*conversation.Conversation, error) {
	var dbConversation Conversation
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbConversation)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, conversation.ErrConversationNotFound
		}
		return nil, result.Error
	}

	return toDomainConversation(&dbConversation), nil
}

// FindByUserID finds conversations for a user with pagination (excludes archived)
func (r *conversationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*conversation.Conversation, error) {
	var dbConversations []Conversation

	result := r.db.WithContext(ctx).
		Select("DISTINCT conversations.*").
		Joins("INNER JOIN conversation_participants ON conversation_participants.conversation_id = conversations.id").
		Where("conversation_participants.user_id = ?", userID).
		Where("conversation_participants.archived_at IS NULL"). // Exclude archived conversations
		Order("conversations.last_message_at DESC NULLS LAST, conversations.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbConversations)

	if result.Error != nil {
		return nil, result.Error
	}

	conversations := make([]*conversation.Conversation, len(dbConversations))
	for i, dbConv := range dbConversations {
		conversations[i] = toDomainConversation(&dbConv)
	}

	return conversations, nil
}

// FindDirectConversation finds a direct conversation between two users
func (r *conversationRepository) FindDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*conversation.Conversation, error) {
	var dbConversation Conversation

	// Find conversation where both users are participants and type is 'direct'
	result := r.db.WithContext(ctx).
		Select("conversations.*").
		Joins("INNER JOIN conversation_participants cp1 ON cp1.conversation_id = conversations.id").
		Joins("INNER JOIN conversation_participants cp2 ON cp2.conversation_id = conversations.id").
		Where("conversations.type = ?", "direct").
		Where("cp1.user_id = ? AND cp2.user_id = ?", user1ID, user2ID).
		First(&dbConversation)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, conversation.ErrConversationNotFound
		}
		return nil, result.Error
	}

	return toDomainConversation(&dbConversation), nil
}

// Update updates a conversation
func (r *conversationRepository) Update(ctx context.Context, c *conversation.Conversation) error {
	dbConversation := toConversationModel(c)
	result := r.db.WithContext(ctx).Model(&Conversation{}).Where("id = ?", c.ID).Updates(dbConversation)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return conversation.ErrConversationNotFound
	}

	return nil
}

// Delete soft deletes a conversation
func (r *conversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Conversation{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return conversation.ErrConversationNotFound
	}

	return nil
}

// AddParticipant adds a participant to a conversation
func (r *conversationRepository) AddParticipant(ctx context.Context, p *conversation.Participant) error {
	dbParticipant := toParticipantModel(p)
	result := r.db.WithContext(ctx).Create(dbParticipant)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return conversation.ErrAlreadyParticipant
		}
		return result.Error
	}

	return nil
}

// RemoveParticipant removes a participant from a conversation
func (r *conversationRepository) RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Delete(&ConversationParticipant{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return conversation.ErrNotParticipant
	}

	return nil
}

// FindParticipants finds all participants of a conversation
func (r *conversationRepository) FindParticipants(ctx context.Context, conversationID uuid.UUID) ([]*conversation.Participant, error) {
	var dbParticipants []ConversationParticipant

	result := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Find(&dbParticipants)

	if result.Error != nil {
		return nil, result.Error
	}

	participants := make([]*conversation.Participant, len(dbParticipants))
	for i, dbPart := range dbParticipants {
		participant := toDomainParticipant(&dbPart)

		// Check online status from Hub
		if r.presenceChecker != nil {
			participant.IsOnline = r.presenceChecker.IsUserOnline(participant.UserID)
		}

		participants[i] = participant
	}

	return participants, nil
}

// IsParticipant checks if a user is a participant of a conversation
func (r *conversationRepository) IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// UpdateParticipantLastRead updates when a participant last read messages
func (r *conversationRepository) UpdateParticipantLastRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_at", gorm.Expr("CURRENT_TIMESTAMP"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return conversation.ErrNotParticipant
	}

	return nil
}

// ArchiveConversation archives a conversation for a specific user
func (r *conversationRepository) ArchiveConversation(ctx context.Context, conversationID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("archived_at", gorm.Expr("CURRENT_TIMESTAMP"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return conversation.ErrNotParticipant
	}

	return nil
}

// UnarchiveConversation unarchives a conversation for a specific user
func (r *conversationRepository) UnarchiveConversation(ctx context.Context, conversationID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("archived_at", nil)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return conversation.ErrNotParticipant
	}

	return nil
}

// Mapper functions

// toConversationModel converts domain Conversation to GORM Conversation model
func toConversationModel(c *conversation.Conversation) *Conversation {
	return &Conversation{
		ID:            c.ID,
		Type:          string(c.Type),
		LastMessageID: c.LastMessageID,
		LastMessageAt: c.LastMessageAt,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// toDomainConversation converts GORM Conversation model to domain Conversation
func toDomainConversation(c *Conversation) *conversation.Conversation {
	return &conversation.Conversation{
		ID:            c.ID,
		Type:          conversation.Type(c.Type),
		LastMessageID: c.LastMessageID,
		LastMessageAt: c.LastMessageAt,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// toParticipantModel converts domain Participant to GORM ConversationParticipant model
func toParticipantModel(p *conversation.Participant) *ConversationParticipant {
	return &ConversationParticipant{
		ConversationID: p.ConversationID,
		UserID:         p.UserID,
		Role:           string(p.Role),
		JoinedAt:       p.JoinedAt,
		LastReadAt:     p.LastReadAt,
		ArchivedAt:     p.ArchivedAt,
	}
}

// toDomainParticipant converts GORM ConversationParticipant model to domain Participant
func toDomainParticipant(p *ConversationParticipant) *conversation.Participant {
	return &conversation.Participant{
		ConversationID: p.ConversationID,
		UserID:         p.UserID,
		Role:           conversation.Role(p.Role),
		JoinedAt:       p.JoinedAt,
		LastReadAt:     p.LastReadAt,
		ArchivedAt:     p.ArchivedAt,
	}
}
