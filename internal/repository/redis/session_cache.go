package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionCache manages user session caching
type SessionCache struct {
	client *redis.Client
}

// NewSessionCache creates a new session cache
func NewSessionCache(client *redis.Client) *SessionCache {
	return &SessionCache{
		client: client,
	}
}

// SessionData represents cached session information
type SessionData struct {
	UserID        uuid.UUID              `json:"user_id"`
	WalletAddress string                 `json:"wallet_address"`
	Username      string                 `json:"username"`
	DisplayName   string                 `json:"display_name"`
	Avatar        string                 `json:"avatar"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	LastActivity  time.Time              `json:"last_activity"`
}

// Set stores a session with TTL
func (s *SessionCache) Set(ctx context.Context, sessionID string, data *SessionData, ttl time.Duration) error {
	key := s.sessionKey(sessionID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	return s.client.Set(ctx, key, jsonData, ttl).Err()
}

// Get retrieves a session
func (s *SessionCache) Get(ctx context.Context, sessionID string) (*SessionData, error) {
	key := s.sessionKey(sessionID)

	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

// Delete removes a session
func (s *SessionCache) Delete(ctx context.Context, sessionID string) error {
	key := s.sessionKey(sessionID)
	return s.client.Del(ctx, key).Err()
}

// DeleteUserSessions removes all sessions for a user
func (s *SessionCache) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	pattern := s.userSessionPattern(userID)

	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// Extend extends the TTL of a session
func (s *SessionCache) Extend(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := s.sessionKey(sessionID)
	return s.client.Expire(ctx, key, ttl).Err()
}

// UpdateLastActivity updates the last activity timestamp
func (s *SessionCache) UpdateLastActivity(ctx context.Context, sessionID string) error {
	key := s.sessionKey(sessionID)

	// Get existing session
	data, err := s.Get(ctx, sessionID)
	if err != nil || data == nil {
		return err
	}

	// Update last activity
	data.LastActivity = time.Now()

	// Get current TTL
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	// Save with same TTL
	return s.Set(ctx, sessionID, data, ttl)
}

// SetUserSession stores a session by user ID (for quick lookup)
func (s *SessionCache) SetUserSession(ctx context.Context, userID uuid.UUID, sessionID string, ttl time.Duration) error {
	key := s.userSessionKey(userID, sessionID)
	return s.client.Set(ctx, key, sessionID, ttl).Err()
}

// GetUserSessions retrieves all session IDs for a user
func (s *SessionCache) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	pattern := s.userSessionPattern(userID)

	var sessionIDs []string
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		sessionID, err := s.client.Get(ctx, iter.Val()).Result()
		if err == nil {
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, iter.Err()
}

// Helper methods for key generation
func (s *SessionCache) sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func (s *SessionCache) userSessionKey(userID uuid.UUID, sessionID string) string {
	return fmt.Sprintf("user_session:%s:%s", userID.String(), sessionID)
}

func (s *SessionCache) userSessionPattern(userID uuid.UUID) string {
	return fmt.Sprintf("user_session:%s:*", userID.String())
}

// GetActiveSessionCount returns the number of active sessions for a user
func (s *SessionCache) GetActiveSessionCount(ctx context.Context, userID uuid.UUID) (int, error) {
	sessions, err := s.GetUserSessions(ctx, userID)
	if err != nil {
		return 0, err
	}
	return len(sessions), nil
}

// InvalidateToken blacklists a token
func (s *SessionCache) InvalidateToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	return s.client.Set(ctx, key, "1", ttl).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *SessionCache) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	val, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}
