package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"identity-srv/internal/authentication"
	"time"
)

// CreateSession creates a new session in Redis
func (sm *SessionManager) CreateSession(ctx context.Context, userID, jti string, rememberMe bool) error {
	// Calculate TTL based on remember me flag
	ttl := sm.ttl
	if rememberMe {
		ttl = 7 * 24 * time.Hour // 7 days
	}

	now := time.Now()
	sessionData := SessionData{
		UserID:    userID,
		JTI:       jti,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	// Serialize session data
	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal session data: %v", authentication.ErrInternalSystem, err)
	}

	// Store in Redis with key: session:{jti}
	key := fmt.Sprintf("session:%s", jti)
	if err := sm.redis.Set(ctx, key, data, ttl); err != nil {
		return fmt.Errorf("%w: failed to store session: %v", authentication.ErrInternalSystem, err)
	}

	// Also store user-to-session mapping for logout all functionality
	// Key: user_sessions:{userID}, Value: JSON array of JTIs
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)

	// Get existing JTIs
	existingJTIs := []string{}
	existingData, err := sm.redis.Get(ctx, userSessionsKey)
	if err == nil && existingData != "" {
		// Parse existing JTIs
		if err := json.Unmarshal([]byte(existingData), &existingJTIs); err == nil {
			// Filter out expired/invalid JTIs by checking if session still exists
			validJTIs := []string{}
			for _, existingJTI := range existingJTIs {
				sessionKey := fmt.Sprintf("session:%s", existingJTI)
				exists, _ := sm.redis.Exists(ctx, sessionKey)
				if exists {
					validJTIs = append(validJTIs, existingJTI)
				}
			}
			existingJTIs = validJTIs
		}
	}

	// Add new JTI
	existingJTIs = append(existingJTIs, jti)

	// Store updated JTIs list
	jtisData, err := json.Marshal(existingJTIs)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal JTIs: %v", authentication.ErrInternalSystem, err)
	}

	// Use longest TTL for the mapping (7 days for remember me)
	mappingTTL := 7 * 24 * time.Hour
	if err := sm.redis.Set(ctx, userSessionsKey, jtisData, mappingTTL); err != nil {
		return fmt.Errorf("%w: failed to store user session mapping: %v", authentication.ErrInternalSystem, err)
	}

	return nil
}

// GetSession retrieves session data by JTI
func (sm *SessionManager) GetSession(ctx context.Context, jti string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", jti)
	data, err := sm.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: session not found: %v", authentication.ErrInternalSystem, err)
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal session data: %v", authentication.ErrInternalSystem, err)
	}

	return &sessionData, nil
}

// DeleteSession deletes a session by JTI
func (sm *SessionManager) DeleteSession(ctx context.Context, jti string) error {
	key := fmt.Sprintf("session:%s", jti)
	if err := sm.redis.Delete(ctx, key); err != nil {
		return fmt.Errorf("%w: failed to delete session: %v", authentication.ErrInternalSystem, err)
	}
	return nil
}

// GetAllUserSessions retrieves all JTIs for a user
func (sm *SessionManager) GetAllUserSessions(ctx context.Context, userID string) ([]string, error) {
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	data, err := sm.redis.Get(ctx, userSessionsKey)
	if err != nil {
		// No sessions found, not an error
		return []string{}, nil
	}

	var jtis []string
	if err := json.Unmarshal([]byte(data), &jtis); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal JTIs: %v", authentication.ErrInternalSystem, err)
	}

	return jtis, nil
}

// DeleteUserSessions deletes all sessions for a user
func (sm *SessionManager) DeleteUserSessions(ctx context.Context, userID string) error {
	// Get all JTIs for the user
	jtis, err := sm.GetAllUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: failed to get user sessions: %v", authentication.ErrInternalSystem, err)
	}

	// Delete each session
	for _, jti := range jtis {
		sessionKey := fmt.Sprintf("session:%s", jti)
		if err := sm.redis.Delete(ctx, sessionKey); err != nil {
			// Log error but continue deleting other sessions
			continue
		}
	}

	// Delete user sessions mapping
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	if err := sm.redis.Delete(ctx, userSessionsKey); err != nil {
		return fmt.Errorf("%w: failed to delete user sessions mapping: %v", authentication.ErrInternalSystem, err)
	}

	return nil
}

// SessionExists checks if a session exists
func (sm *SessionManager) SessionExists(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("session:%s", jti)
	return sm.redis.Exists(ctx, key)
}
