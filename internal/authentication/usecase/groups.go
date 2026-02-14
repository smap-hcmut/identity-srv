package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"smap-api/internal/authentication"
)

// GetUserGroups fetches user groups with cache-first lookup
func (gm *GroupsManager) GetUserGroups(ctx context.Context, userEmail string) ([]string, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user_groups:%s", userEmail)
	cachedData, err := gm.redis.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		// Cache hit - deserialize and return
		var groups []string
		if err := json.Unmarshal([]byte(cachedData), &groups); err == nil {
			return groups, nil
		}
		// If unmarshal fails, continue to fetch from API
	}

	// Cache miss or error - fetch from Google Directory API
	if gm.googleClient == nil {
		// Google Client not configured, return empty groups (safe fallback)
		return []string{}, nil
	}
	groups, err := gm.googleClient.GetUserGroups(ctx, userEmail)
	if err != nil {
		// If API call fails, try to return stale cached data if available
		if cachedData != "" {
			var groups []string
			if err := json.Unmarshal([]byte(cachedData), &groups); err == nil {
				// Return stale cache with warning
				return groups, nil
			}
		}
		return nil, fmt.Errorf("%w: failed to fetch user groups: %v", authentication.ErrInternalSystem, err)
	}

	// Store in cache for future requests
	groupsJSON, err := json.Marshal(groups)
	if err == nil {
		// Ignore cache write errors - not critical
		_ = gm.redis.Set(ctx, cacheKey, groupsJSON, gm.cacheTTL)
	}

	return groups, nil
}

// InvalidateUserGroupsCache invalidates the cache for a specific user
func (gm *GroupsManager) InvalidateUserGroupsCache(ctx context.Context, userEmail string) error {
	cacheKey := fmt.Sprintf("user_groups:%s", userEmail)
	return gm.redis.Delete(ctx, cacheKey)
}

// RefreshUserGroups forces a refresh of user groups from API
func (gm *GroupsManager) RefreshUserGroups(ctx context.Context, userEmail string) ([]string, error) {
	// Invalidate cache first
	_ = gm.InvalidateUserGroupsCache(ctx, userEmail)

	// Fetch fresh data from API
	return gm.GetUserGroups(ctx, userEmail)
}
