package usecase

import (
	"context"
	"fmt"
	"smap-api/internal/authentication"
	"time"
)

// AddToken adds a token to the blacklist by JTI
// TTL is set to the remaining token lifetime to automatically expire
func (bm *BlacklistManager) AddToken(ctx context.Context, jti string, expiresAt time.Time) error {
	// Calculate TTL as remaining token lifetime
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	// Store in Redis with key: blacklist:{jti}
	key := fmt.Sprintf("blacklist:%s", jti)
	if err := bm.redis.Set(ctx, key, "1", ttl); err != nil {
		return fmt.Errorf("%w: failed to add token to blacklist: %v", authentication.ErrInternalSystem, err)
	}

	return nil
}

// AddAllUserTokens adds all tokens for a user to the blacklist
// This is used when revoking all sessions for a user
func (bm *BlacklistManager) AddAllUserTokens(ctx context.Context, jtis []string, expiresAt time.Time) error {
	// Calculate TTL as remaining token lifetime
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Tokens already expired, no need to blacklist
		return nil
	}

	// Add each JTI to blacklist
	for _, jti := range jtis {
		key := fmt.Sprintf("blacklist:%s", jti)
		if err := bm.redis.Set(ctx, key, "1", ttl); err != nil {
			return fmt.Errorf("%w: failed to add user token to blacklist: %v", authentication.ErrInternalSystem, err)
		}
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted by JTI
func (bm *BlacklistManager) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", jti)
	exists, err := bm.redis.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check blacklist: %v", authentication.ErrInternalSystem, err)
	}
	return exists, nil
}

// RemoveToken removes a token from the blacklist (rarely used)
func (bm *BlacklistManager) RemoveToken(ctx context.Context, jti string) error {
	key := fmt.Sprintf("blacklist:%s", jti)
	if err := bm.redis.Delete(ctx, key); err != nil {
		return fmt.Errorf("%w: failed to remove token from blacklist: %v", authentication.ErrInternalSystem, err)
	}
	return nil
}
