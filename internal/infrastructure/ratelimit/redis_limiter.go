package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/sotalk/internal/domain/privacy"
)

// RedisLimiter implements rate limiting using Redis
type RedisLimiter struct {
	client *redis.Client
}

// NewRedisLimiter creates a new Redis-based rate limiter
func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
	}
}

// Allow checks if a request is allowed under the rate limit
func (r *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, *privacy.RateLimitInfo, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(window.Seconds()))

	// Increment counter
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, window)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return false, nil, err
	}

	count := int(incrCmd.Val())
	allowed := count <= limit

	info := &privacy.RateLimitInfo{
		Key:        key,
		Count:      count,
		Limit:      limit,
		WindowSecs: int(window.Seconds()),
		ExpiresAt:  now.Add(window),
	}

	return allowed, info, nil
}

// Check retrieves current rate limit info without incrementing
func (r *RedisLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (*privacy.RateLimitInfo, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(window.Seconds()))

	count, err := r.client.Get(ctx, windowKey).Int()
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		return nil, err
	}

	info := &privacy.RateLimitInfo{
		Key:        key,
		Count:      count,
		Limit:      limit,
		WindowSecs: int(window.Seconds()),
		ExpiresAt:  now.Add(window),
	}

	return info, nil
}

// Reset removes rate limit for a specific key
func (r *RedisLimiter) Reset(ctx context.Context, key string) error {
	now := time.Now()
	// Try to delete keys from last few windows
	for i := 0; i < 5; i++ {
		windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()-int64(i*60))
		r.client.Del(ctx, windowKey)
	}
	return nil
}

// Pre-defined rate limit configurations
var (
	// API rate limits
	RateLimitAPIDefault   = RateLimitConfig{Limit: 100, Window: time.Minute}
	RateLimitAPIStrict    = RateLimitConfig{Limit: 10, Window: time.Minute}
	RateLimitAuth         = RateLimitConfig{Limit: 5, Window: time.Minute}

	// Message rate limits
	RateLimitMessage      = RateLimitConfig{Limit: 50, Window: time.Minute}
	RateLimitMessageBurst = RateLimitConfig{Limit: 10, Window: 10 * time.Second}

	// Payment rate limits
	RateLimitPayment      = RateLimitConfig{Limit: 10, Window: time.Minute}

	// Media upload rate limits
	RateLimitMediaUpload  = RateLimitConfig{Limit: 20, Window: time.Minute}
)

// RateLimitConfig defines a rate limit configuration
type RateLimitConfig struct {
	Limit  int
	Window time.Duration
}

// Apply applies the rate limit and returns whether it's allowed
func (c RateLimitConfig) Apply(ctx context.Context, limiter *RedisLimiter, key string) (bool, *privacy.RateLimitInfo, error) {
	return limiter.Allow(ctx, key, c.Limit, c.Window)
}
