package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// PresenceCache manages user online/offline status
type PresenceCache struct {
	client *redis.Client
}

// NewPresenceCache creates a new presence cache
func NewPresenceCache(client *redis.Client) *PresenceCache {
	return &PresenceCache{
		client: client,
	}
}

// PresenceStatus represents user presence status
type PresenceStatus string

const (
	StatusOnline  PresenceStatus = "online"
	StatusOffline PresenceStatus = "offline"
	StatusAway    PresenceStatus = "away"
)

// SetOnline marks a user as online
func (p *PresenceCache) SetOnline(ctx context.Context, userID uuid.UUID, ttl time.Duration) error {
	key := p.presenceKey(userID)

	pipe := p.client.Pipeline()
	pipe.Set(ctx, key, string(StatusOnline), ttl)
	pipe.Set(ctx, p.lastSeenKey(userID), time.Now().Unix(), 24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}

// SetOffline marks a user as offline
func (p *PresenceCache) SetOffline(ctx context.Context, userID uuid.UUID) error {
	key := p.presenceKey(userID)

	pipe := p.client.Pipeline()
	pipe.Set(ctx, key, string(StatusOffline), 24*time.Hour)
	pipe.Set(ctx, p.lastSeenKey(userID), time.Now().Unix(), 24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}

// SetAway marks a user as away
func (p *PresenceCache) SetAway(ctx context.Context, userID uuid.UUID) error {
	key := p.presenceKey(userID)

	pipe := p.client.Pipeline()
	pipe.Set(ctx, key, string(StatusAway), 24*time.Hour)
	pipe.Set(ctx, p.lastSeenKey(userID), time.Now().Unix(), 24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}

// GetStatus retrieves a user's presence status
func (p *PresenceCache) GetStatus(ctx context.Context, userID uuid.UUID) (PresenceStatus, error) {
	key := p.presenceKey(userID)

	val, err := p.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return StatusOffline, nil
	}
	if err != nil {
		return StatusOffline, err
	}

	return PresenceStatus(val), nil
}

// IsOnline checks if a user is online
func (p *PresenceCache) IsOnline(ctx context.Context, userID uuid.UUID) (bool, error) {
	status, err := p.GetStatus(ctx, userID)
	if err != nil {
		return false, err
	}
	return status == StatusOnline, nil
}

// GetLastSeen retrieves the last seen timestamp for a user
func (p *PresenceCache) GetLastSeen(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	key := p.lastSeenKey(userID)

	timestamp, err := p.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	lastSeen := time.Unix(timestamp, 0)
	return &lastSeen, nil
}

// SetLastSeen updates the last seen timestamp
func (p *PresenceCache) SetLastSeen(ctx context.Context, userID uuid.UUID) error {
	key := p.lastSeenKey(userID)
	return p.client.Set(ctx, key, time.Now().Unix(), 24*time.Hour).Err()
}

// ExtendOnline extends the online status TTL (for heartbeat)
func (p *PresenceCache) ExtendOnline(ctx context.Context, userID uuid.UUID, ttl time.Duration) error {
	// Only extend if currently online
	status, err := p.GetStatus(ctx, userID)
	if err != nil {
		return err
	}

	if status == StatusOnline {
		return p.SetOnline(ctx, userID, ttl)
	}

	return nil
}

// GetOnlineUsers returns all currently online users
func (p *PresenceCache) GetOnlineUsers(ctx context.Context) ([]uuid.UUID, error) {
	pattern := "presence:*"

	var onlineUsers []uuid.UUID
	iter := p.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		// Get status
		status, err := p.client.Get(ctx, key).Result()
		if err != nil || status != string(StatusOnline) {
			continue
		}

		// Extract user ID from key
		var userID uuid.UUID
		if _, err := fmt.Sscanf(key, "presence:%s", &userID); err == nil {
			onlineUsers = append(onlineUsers, userID)
		}
	}

	return onlineUsers, iter.Err()
}

// GetBulkStatus retrieves presence status for multiple users
func (p *PresenceCache) GetBulkStatus(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]PresenceStatus, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]PresenceStatus{}, nil
	}

	pipe := p.client.Pipeline()

	// Create commands for all users
	cmds := make(map[uuid.UUID]*redis.StringCmd)
	for _, userID := range userIDs {
		key := p.presenceKey(userID)
		cmds[userID] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)

	// Collect results
	result := make(map[uuid.UUID]PresenceStatus)
	for userID, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			result[userID] = StatusOffline
		} else if err != nil {
			result[userID] = StatusOffline
		} else {
			result[userID] = PresenceStatus(val)
		}
	}

	return result, err
}

// SetTyping sets typing indicator for a user in a conversation
func (p *PresenceCache) SetTyping(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID, ttl time.Duration) error {
	key := p.typingKey(conversationID, userID)
	return p.client.Set(ctx, key, "1", ttl).Err()
}

// ClearTyping removes typing indicator
func (p *PresenceCache) ClearTyping(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) error {
	key := p.typingKey(conversationID, userID)
	return p.client.Del(ctx, key).Err()
}

// GetTypingUsers returns all users currently typing in a conversation
func (p *PresenceCache) GetTypingUsers(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	pattern := p.typingPattern(conversationID)

	var typingUsers []uuid.UUID
	iter := p.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		// Extract user ID from key
		var userID uuid.UUID
		if _, err := fmt.Sscanf(key, "typing:%s:%s", &conversationID, &userID); err == nil {
			typingUsers = append(typingUsers, userID)
		}
	}

	return typingUsers, iter.Err()
}

// IsTyping checks if a user is typing in a conversation
func (p *PresenceCache) IsTyping(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) (bool, error) {
	key := p.typingKey(conversationID, userID)
	val, err := p.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// GetOnlineCount returns the count of online users
func (p *PresenceCache) GetOnlineCount(ctx context.Context) (int, error) {
	users, err := p.GetOnlineUsers(ctx)
	if err != nil {
		return 0, err
	}
	return len(users), nil
}

// Helper methods for key generation
func (p *PresenceCache) presenceKey(userID uuid.UUID) string {
	return fmt.Sprintf("presence:%s", userID.String())
}

func (p *PresenceCache) lastSeenKey(userID uuid.UUID) string {
	return fmt.Sprintf("lastseen:%s", userID.String())
}

func (p *PresenceCache) typingKey(conversationID uuid.UUID, userID uuid.UUID) string {
	return fmt.Sprintf("typing:%s:%s", conversationID.String(), userID.String())
}

func (p *PresenceCache) typingPattern(conversationID uuid.UUID) string {
	return fmt.Sprintf("typing:%s:*", conversationID.String())
}
