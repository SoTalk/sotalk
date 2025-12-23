package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// MessageCache manages message caching
type MessageCache struct {
	client *redis.Client
}

// NewMessageCache creates a new message cache
func NewMessageCache(client *redis.Client) *MessageCache {
	return &MessageCache{
		client: client,
	}
}

// CachedMessage represents a cached message
type CachedMessage struct {
	ID              uuid.UUID `json:"id"`
	ConversationID  uuid.UUID `json:"conversation_id"`
	SenderID        uuid.UUID `json:"sender_id"`
	Content         []byte    `json:"content"`
	ContentType     string    `json:"content_type"`
	Status          string    `json:"status"`
	ReplyToID       *uuid.UUID `json:"reply_to_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// SetMessage caches a single message
func (m *MessageCache) SetMessage(ctx context.Context, msg *CachedMessage, ttl time.Duration) error {
	key := m.messageKey(msg.ID)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return m.client.Set(ctx, key, jsonData, ttl).Err()
}

// GetMessage retrieves a cached message
func (m *MessageCache) GetMessage(ctx context.Context, messageID uuid.UUID) (*CachedMessage, error) {
	key := m.messageKey(messageID)

	val, err := m.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	var msg CachedMessage
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// SetConversationMessages caches recent messages for a conversation (ordered list)
func (m *MessageCache) SetConversationMessages(ctx context.Context, conversationID uuid.UUID, messages []*CachedMessage, ttl time.Duration) error {
	key := m.conversationMessagesKey(conversationID)

	// Clear existing list
	if err := m.client.Del(ctx, key).Err(); err != nil {
		return err
	}

	// Add messages to sorted set (score = timestamp for ordering)
	pipe := m.client.Pipeline()

	for _, msg := range messages {
		jsonData, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		// Use Unix timestamp as score for sorting
		score := float64(msg.CreatedAt.Unix())
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  score,
			Member: jsonData,
		})
	}

	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// GetConversationMessages retrieves recent messages for a conversation
func (m *MessageCache) GetConversationMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*CachedMessage, error) {
	key := m.conversationMessagesKey(conversationID)

	// Get messages in descending order (most recent first)
	vals, err := m.client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	messages := make([]*CachedMessage, 0, len(vals))
	for _, val := range vals {
		var msg CachedMessage
		if err := json.Unmarshal([]byte(val), &msg); err != nil {
			continue // Skip malformed messages
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// AddMessageToConversation adds a new message to the conversation cache
func (m *MessageCache) AddMessageToConversation(ctx context.Context, conversationID uuid.UUID, msg *CachedMessage) error {
	key := m.conversationMessagesKey(conversationID)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Add to sorted set
	score := float64(msg.CreatedAt.Unix())
	if err := m.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: jsonData,
	}).Err(); err != nil {
		return err
	}

	// Keep only last 100 messages per conversation
	return m.client.ZRemRangeByRank(ctx, key, 0, -101).Err()
}

// DeleteMessage removes a message from cache
func (m *MessageCache) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	key := m.messageKey(messageID)
	return m.client.Del(ctx, key).Err()
}

// DeleteConversationMessages clears all cached messages for a conversation
func (m *MessageCache) DeleteConversationMessages(ctx context.Context, conversationID uuid.UUID) error {
	key := m.conversationMessagesKey(conversationID)
	return m.client.Del(ctx, key).Err()
}

// UpdateMessageStatus updates the status of a cached message
func (m *MessageCache) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status string) error {
	msg, err := m.GetMessage(ctx, messageID)
	if err != nil || msg == nil {
		return err
	}

	msg.Status = status
	msg.UpdatedAt = time.Now()

	return m.SetMessage(ctx, msg, 15*time.Minute)
}

// GetUnreadCount gets the count of unread messages for a conversation (from cache)
func (m *MessageCache) GetUnreadCount(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) (int, error) {
	key := m.unreadCountKey(conversationID, userID)

	count, err := m.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// SetUnreadCount sets the unread message count
func (m *MessageCache) SetUnreadCount(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID, count int) error {
	key := m.unreadCountKey(conversationID, userID)
	return m.client.Set(ctx, key, count, 24*time.Hour).Err()
}

// IncrementUnreadCount increments the unread message count
func (m *MessageCache) IncrementUnreadCount(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) error {
	key := m.unreadCountKey(conversationID, userID)
	return m.client.Incr(ctx, key).Err()
}

// ResetUnreadCount resets the unread message count to 0
func (m *MessageCache) ResetUnreadCount(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) error {
	key := m.unreadCountKey(conversationID, userID)
	return m.client.Del(ctx, key).Err()
}

// Helper methods for key generation
func (m *MessageCache) messageKey(messageID uuid.UUID) string {
	return fmt.Sprintf("message:%s", messageID.String())
}

func (m *MessageCache) conversationMessagesKey(conversationID uuid.UUID) string {
	return fmt.Sprintf("conversation:messages:%s", conversationID.String())
}

func (m *MessageCache) unreadCountKey(conversationID uuid.UUID, userID uuid.UUID) string {
	return fmt.Sprintf("unread:%s:%s", conversationID.String(), userID.String())
}

// GetConversationMessageCount returns the number of cached messages for a conversation
func (m *MessageCache) GetConversationMessageCount(ctx context.Context, conversationID uuid.UUID) (int64, error) {
	key := m.conversationMessagesKey(conversationID)
	return m.client.ZCard(ctx, key).Result()
}
