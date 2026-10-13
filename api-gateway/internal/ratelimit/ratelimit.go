package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	window time.Duration
}

type Result struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetAt   time.Time
}

func New(client *redis.Client, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		window: window,
	}
}

func (rl *RateLimiter) Allow(
	ctx context.Context,
	key string,
	limit int,
) (*Result, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("rate limit must be positive")
	}
	if rl.window <= 0 {
		return nil, fmt.Errorf("rate limit window must be positive")
	}

	redisKey := "ratelimit:" + key

	count, err := rl.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("increment rate limit counter: %w", err)
	}

	if count == 1 {
		if err := rl.client.Expire(ctx, redisKey, rl.window).Err(); err != nil {
			return nil, fmt.Errorf("set rate limit expiration: %w", err)
		}
	}

	ttl, err := rl.client.TTL(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get rate limit TTL: %w", err)
	}

	if ttl < 0 {
		return nil, fmt.Errorf("rate limit key has no expiration")
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return &Result{
		Allowed:   count <= int64(limit),
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   time.Now().Add(ttl),
	}, nil
}
