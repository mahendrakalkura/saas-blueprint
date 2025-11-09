package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/redis"
)

// Limiter implements rate limiting using Redis
type Limiter struct {
	redis *redis.Client
}

// Config defines rate limit configuration
type Config struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window for the rate limit
}

// Result contains information about a rate limit check
type Result struct {
	Allowed    bool      // Whether the request is allowed
	Limit      int       // The rate limit ceiling
	Remaining  int       // Remaining requests in current window
	RetryAfter time.Time // When to retry if rate limited
}

func NewLimiter(redis *redis.Client) *Limiter {
	return &Limiter{redis: redis}
}

// Allow checks if a request is allowed under the rate limit
// Uses a sliding window algorithm with Redis
func (l *Limiter) Allow(ctx context.Context, key string, config Config) (*Result, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(config.Window.Seconds()))

	// Increment the counter
	count, err := l.redis.Incr(ctx, windowKey)
	if err != nil {
		return nil, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Set expiration on first request in this window
	if count == 1 {
		if err := l.redis.Expire(ctx, windowKey, config.Window+time.Second); err != nil {
			return nil, fmt.Errorf("failed to set expiration: %w", err)
		}
	}

	remaining := config.Requests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	allowed := count <= int64(config.Requests)
	retryAfter := now.Add(config.Window)

	return &Result{
		Allowed:    allowed,
		Limit:      config.Requests,
		Remaining:  remaining,
		RetryAfter: retryAfter,
	}, nil
}

// Reset removes all rate limit entries for a key
func (l *Limiter) Reset(ctx context.Context, key string) error {
	pattern := fmt.Sprintf("ratelimit:%s:*", key)

	// Get keys matching the pattern
	iter := l.redis.Client().Scan(ctx, 0, pattern, 0).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		if err := l.redis.Del(ctx, keys...); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}
