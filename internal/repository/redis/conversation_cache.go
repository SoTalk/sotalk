package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ConversationCache manages conversation list caching
type ConversationCache struct {
	client *redis.Client
}

// NewConversationCache creates a new conversation cache
func NewConversationCache(client *redis.Client) *ConversationCache {
	return &ConversationCache{
		client: client,
	}
}

// CachedConversation represents a cached conversation summary
type CachedConversation struct {
	ID              uuid.UUID  `json:"id"`
	Type            string     `json:"type"`
	Name            string     `json:"name,omitempty"`
	Avatar          string     `json:"avatar,omitempty"`
	LastMessageID   *uuid.UUID `json:"last_message_id,omitempty"`
	LastMessageText string     `json:"last_message_text,omitempty"`
	LastMessageAt   time.Time  `json:"last_message_at"`
	UnreadCount     int        `json:"unread_count"`
	IsMuted         bool       `json:"is_muted"`
	IsPinned        bool       `json:"is_pinned"`
	Participants    []uuid.UUID `json:"participants,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// SetUserConversations caches the conversation list for a user
func (c *ConversationCache) SetUserConversations(ctx context.Context, userID uuid.UUID, conversations []*CachedConversation, ttl time.Duration) error {
	key := c.userConversationsKey(userID)

	// Clear existing list
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return err
	}

	if len(conversations) == 0 {
		return nil
	}

	// Add conversations to sorted set (score = last message timestamp)
	pipe := c.client.Pipeline()

	for _, conv := range conversations {
		jsonData, err := json.Marshal(conv)
		if err != nil {
			return fmt.Errorf("failed to marshal conversation: %w", err)
		}

		// Use last message timestamp as score for sorting
		score := float64(conv.LastMessageAt.Unix())
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  score,
			Member: jsonData,
		})
	}

	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// GetUserConversations retrieves cached conversations for a user
func (c *ConversationCache) GetUserConversations(ctx context.Context, userID uuid.UUID, limit int) ([]*CachedConversation, error) {
	key := c.userConversationsKey(userID)

	// Get conversations in descending order (most recent first)
	vals, err := c.client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}

	conversations := make([]*CachedConversation, 0, len(vals))
	for _, val := range vals {
		var conv CachedConversation
		if err := json.Unmarshal([]byte(val), &conv); err != nil {
			continue // Skip malformed conversations
		}
		conversations = append(conversations, &conv)
	}

	return conversations, nil
}

// AddConversation adds or updates a conversation in the user's list
func (c *ConversationCache) AddConversation(ctx context.Context, userID uuid.UUID, conversation *CachedConversation) error {
	key := c.userConversationsKey(userID)

	jsonData, err := json.Marshal(conversation)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	// Add to sorted set with last message timestamp as score
	score := float64(conversation.LastMessageAt.Unix())
	return c.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: jsonData,
	}).Err()
}

// RemoveConversation removes a conversation from the user's cached list
func (c *ConversationCache) RemoveConversation(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) error {
	key := c.userConversationsKey(userID)

	// Find and remove the conversation
	// We need to iterate because we're storing JSON
	vals, err := c.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, val := range vals {
		var conv CachedConversation
		if err := json.Unmarshal([]byte(val), &conv); err != nil {
			continue
		}

		if conv.ID == conversationID {
			return c.client.ZRem(ctx, key, val).Err()
		}
	}

	return nil
}

// UpdateLastMessage updates the last message info for a conversation
func (c *ConversationCache) UpdateLastMessage(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, messageID uuid.UUID, messageText string, timestamp time.Time) error {
	key := c.userConversationsKey(userID)

	// Get all conversations
	vals, err := c.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	// Find and update the specific conversation
	for _, val := range vals {
		var conv CachedConversation
		if err := json.Unmarshal([]byte(val), &conv); err != nil {
			continue
		}

		if conv.ID == conversationID {
			// Remove old entry
			if err := c.client.ZRem(ctx, key, val).Err(); err != nil {
				return err
			}

			// Update conversation
			conv.LastMessageID = &messageID
			conv.LastMessageText = messageText
			conv.LastMessageAt = timestamp
			conv.UpdatedAt = time.Now()

			// Re-add with new score
			return c.AddConversation(ctx, userID, &conv)
		}
	}

	return nil
}

// IncrementUnread increments the unread count for a conversation
func (c *ConversationCache) IncrementUnread(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) error {
	return c.updateUnreadCount(ctx, userID, conversationID, 1)
}

// ResetUnread resets the unread count for a conversation
func (c *ConversationCache) ResetUnread(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) error {
	return c.setUnreadCount(ctx, userID, conversationID, 0)
}

// updateUnreadCount updates the unread count by delta
func (c *ConversationCache) updateUnreadCount(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, delta int) error {
	key := c.userConversationsKey(userID)

	vals, err := c.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, val := range vals {
		var conv CachedConversation
		if err := json.Unmarshal([]byte(val), &conv); err != nil {
			continue
		}

		if conv.ID == conversationID {
			// Remove old entry
			if err := c.client.ZRem(ctx, key, val).Err(); err != nil {
				return err
			}

			// Update unread count
			conv.UnreadCount += delta
			if conv.UnreadCount < 0 {
				conv.UnreadCount = 0
			}
			conv.UpdatedAt = time.Now()

			// Re-add
			return c.AddConversation(ctx, userID, &conv)
		}
	}

	return nil
}

// setUnreadCount sets the exact unread count
func (c *ConversationCache) setUnreadCount(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, count int) error {
	key := c.userConversationsKey(userID)

	vals, err := c.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, val := range vals {
		var conv CachedConversation
		if err := json.Unmarshal([]byte(val), &conv); err != nil {
			continue
		}

		if conv.ID == conversationID {
			// Remove old entry
			if err := c.client.ZRem(ctx, key, val).Err(); err != nil {
				return err
			}

			// Set unread count
			conv.UnreadCount = count
			conv.UpdatedAt = time.Now()

			// Re-add
			return c.AddConversation(ctx, userID, &conv)
		}
	}

	return nil
}

// SetPinned sets the pinned status for a conversation
func (c *ConversationCache) SetPinned(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, pinned bool) error {
	key := c.userConversationsKey(userID)

	vals, err := c.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, val := range vals {
		var conv CachedConversation
		if err := json.Unmarshal([]byte(val), &conv); err != nil {
			continue
		}

		if conv.ID == conversationID {
			// Remove old entry
			if err := c.client.ZRem(ctx, key, val).Err(); err != nil {
				return err
			}

			// Update pinned status
			conv.IsPinned = pinned
			conv.UpdatedAt = time.Now()

			// Re-add
			return c.AddConversation(ctx, userID, &conv)
		}
	}

	return nil
}

// SetMuted sets the muted status for a conversation
func (c *ConversationCache) SetMuted(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, muted bool) error {
	key := c.userConversationsKey(userID)

	vals, err := c.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, val := range vals {
		var conv CachedConversation
		if err := json.Unmarshal([]byte(val), &conv); err != nil {
			continue
		}

		if conv.ID == conversationID {
			// Remove old entry
			if err := c.client.ZRem(ctx, key, val).Err(); err != nil {
				return err
			}

			// Update muted status
			conv.IsMuted = muted
			conv.UpdatedAt = time.Now()

			// Re-add
			return c.AddConversation(ctx, userID, &conv)
		}
	}

	return nil
}

// DeleteUserConversations clears all cached conversations for a user
func (c *ConversationCache) DeleteUserConversations(ctx context.Context, userID uuid.UUID) error {
	key := c.userConversationsKey(userID)
	return c.client.Del(ctx, key).Err()
}

// GetConversationCount returns the number of cached conversations for a user
func (c *ConversationCache) GetConversationCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	key := c.userConversationsKey(userID)
	return c.client.ZCard(ctx, key).Result()
}

// GetTotalUnreadCount returns the total unread count across all conversations
func (c *ConversationCache) GetTotalUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	conversations, err := c.GetUserConversations(ctx, userID, 1000) // Get all
	if err != nil {
		return 0, err
	}

	total := 0
	for _, conv := range conversations {
		total += conv.UnreadCount
	}

	return total, nil
}

// Helper methods for key generation
func (c *ConversationCache) userConversationsKey(userID uuid.UUID) string {
	return fmt.Sprintf("conversations:user:%s", userID.String())
}
