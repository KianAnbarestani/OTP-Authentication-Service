package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	max    int
	window time.Duration
	prefix string
}

func NewRateLimiter(client *redis.Client, max int, window time.Duration) *RateLimiter {
	return &RateLimiter{client: client, max: max, window: window, prefix: "rl:"}
}

func (r *RateLimiter) Allow(ctx context.Context, phone string) (bool, int, error) {
	key := r.prefix + phone
	// INCR + EXPIRE atomically using pipeline
	cmd := r.client.Incr(ctx, key)
	if cmd.Err() != nil {
		return false, 0, cmd.Err()
	}
	cnt := int(cmd.Val())
	if cnt == 1 {
		// first request: set expiry
		if err := r.client.Expire(ctx, key, r.window).Err(); err != nil {
			return false, cnt, err
		}
	}
	allowed := cnt <= r.max
	return allowed, r.max - cnt, nil
}

func (r *RateLimiter) Reset(ctx context.Context, phone string) error {
	return r.client.Del(ctx, r.prefix+phone).Err()
}
