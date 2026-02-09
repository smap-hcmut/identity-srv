package usecase

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter implements login rate limiting to prevent brute force attacks
type RateLimiter struct {
	redis          *redis.Client
	maxAttempts    int
	windowDuration time.Duration
	blockDuration  time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, maxAttempts int, windowDuration, blockDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:          redisClient,
		maxAttempts:    maxAttempts,
		windowDuration: windowDuration,
		blockDuration:  blockDuration,
	}
}

// LoginRateLimit middleware tracks failed login attempts by IP address
// Blocks login attempts after maxAttempts failures within windowDuration
func (rl *RateLimiter) LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()

		// Check if IP is blocked
		blocked, err := rl.isBlocked(ctx, ip)
		if err != nil {
			// Log error but don't block request on Redis failure
			c.Next()
			return
		}

		if blocked {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "too_many_requests",
				"message": "Too many failed login attempts. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RecordFailedAttempt records a failed login attempt for an IP address
func (rl *RateLimiter) RecordFailedAttempt(ctx context.Context, ip string) error {
	key := fmt.Sprintf("ratelimit:login:%s", ip)

	// Increment attempt counter
	count, err := rl.redis.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment rate limit counter: %w", err)
	}

	// Set expiration on first attempt
	if count == 1 {
		rl.redis.Expire(ctx, key, rl.windowDuration)
	}

	// Block IP if max attempts exceeded
	if count >= int64(rl.maxAttempts) {
		blockKey := fmt.Sprintf("ratelimit:block:%s", ip)
		rl.redis.Set(ctx, blockKey, "1", rl.blockDuration)
	}

	return nil
}

// ClearFailedAttempts clears failed login attempts for an IP (called on successful login)
func (rl *RateLimiter) ClearFailedAttempts(ctx context.Context, ip string) error {
	key := fmt.Sprintf("ratelimit:login:%s", ip)
	return rl.redis.Del(ctx, key).Err()
}

// isBlocked checks if an IP is currently blocked
func (rl *RateLimiter) isBlocked(ctx context.Context, ip string) (bool, error) {
	blockKey := fmt.Sprintf("ratelimit:block:%s", ip)
	exists, err := rl.redis.Exists(ctx, blockKey).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// IsBlocked checks if an IP is currently blocked (public method)
func (rl *RateLimiter) IsBlocked(ctx context.Context, ip string) (bool, error) {
	return rl.isBlocked(ctx, ip)
}

// GetRemainingAttempts returns the number of remaining login attempts for an IP
func (rl *RateLimiter) GetRemainingAttempts(ctx context.Context, ip string) (int, error) {
	key := fmt.Sprintf("ratelimit:login:%s", ip)
	count, err := rl.redis.Get(ctx, key).Int()
	if err == redis.Nil {
		return rl.maxAttempts, nil
	}
	if err != nil {
		return 0, err
	}

	remaining := rl.maxAttempts - count
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}
