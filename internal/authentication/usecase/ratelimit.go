package usecase

import (
	"context"
	"fmt"
	"smap-api/internal/authentication"
	"strconv"
)

// RecordFailedAttempt records a failed login attempt for an IP address
func (rl *RateLimiter) RecordFailedAttempt(ctx context.Context, ip string) error {
	key := fmt.Sprintf("ratelimit:login:%s", ip)
	client := rl.redis.GetClient()

	// Increment attempt counter
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("%w: failed to increment rate limit counter: %v", authentication.ErrInternalSystem, err)
	}

	// Set expiration on first attempt
	if count == 1 {
		client.Expire(ctx, key, rl.windowDuration)
	}

	// Block IP if max attempts exceeded
	if count >= int64(rl.maxAttempts) {
		blockKey := fmt.Sprintf("ratelimit:block:%s", ip)
		client.Set(ctx, blockKey, "1", rl.blockDuration)
	}

	return nil
}

// ClearFailedAttempts clears failed login attempts for an IP (called on successful login)
func (rl *RateLimiter) ClearFailedAttempts(ctx context.Context, ip string) error {
	key := fmt.Sprintf("ratelimit:login:%s", ip)
	return rl.redis.Delete(ctx, key)
}

// IsBlocked checks if an IP is currently blocked
func (rl *RateLimiter) IsBlocked(ctx context.Context, ip string) (bool, error) {
	blockKey := fmt.Sprintf("ratelimit:block:%s", ip)
	return rl.redis.Exists(ctx, blockKey)
}

// GetRemainingAttempts returns the number of remaining login attempts for an IP
func (rl *RateLimiter) GetRemainingAttempts(ctx context.Context, ip string) (int, error) {
	key := fmt.Sprintf("ratelimit:login:%s", ip)

	// Use pkg/redis wrapper — returns ("", redis.Nil) when key not found
	val, err := rl.redis.Get(ctx, key)
	if err != nil {
		// Key not found means no attempts recorded yet
		return rl.maxAttempts, nil
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to parse rate limit counter: %v", authentication.ErrInternalSystem, err)
	}

	remaining := rl.maxAttempts - count
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}
