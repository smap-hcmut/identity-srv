package usecase

import (
	"context"
	"fmt"
	"smap-api/internal/model"
	"smap-api/internal/user"
	"strings"
	"time"
)

// --- internal helpers (private, not exposed on the interface) ---

// generateState generates a CSRF state token
func (u *implUsecase) generateState() string {
	return fmt.Sprintf("state-%d", u.clock().UnixNano())
}

// isAllowedDomain checks if the email domain is in the allowlist
func (u *implUsecase) isAllowedDomain(email string) bool {
	if len(u.allowedDomains) == 0 {
		return true // No restrictions configured
	}
	domain := extractDomain(email)
	for _, d := range u.allowedDomains {
		if domain == d {
			return true
		}
	}
	return false
}

// isBlockedEmail checks if the email is in the blocklist
func (u *implUsecase) isBlockedEmail(email string) bool {
	for _, blocked := range u.blockedEmails {
		if email == blocked {
			return true
		}
	}
	return false
}

// extractDomain extracts domain from email address
func extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// createOrUpdateUser creates or updates a user via the user UseCase
func (u *implUsecase) createOrUpdateUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	usr, err := u.userUC.Create(ctx, user.CreateInput{
		Email:     email,
		Name:      name,
		AvatarURL: avatarURL,
	})
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.createOrUpdateUser: %v", err)
		return nil, err
	}
	return &usr, nil
}

// updateUserRole updates the user's role
func (u *implUsecase) updateUserRole(ctx context.Context, userID, role string) error {
	return u.userUC.Update(ctx, user.UpdateInput{
		UserID: userID,
		Role:   role,
	})
}

// getUserGroups fetches user groups from the groups manager
func (u *implUsecase) getUserGroups(ctx context.Context, email string) ([]string, error) {
	if u.groupsManager == nil {
		return []string{}, nil
	}
	return u.groupsManager.GetUserGroups(ctx, email)
}

// mapGroupsToRole maps groups to a role using the role mapper
func (u *implUsecase) mapGroupsToRole(groups []string) string {
	if u.roleMapper == nil {
		return "VIEWER"
	}
	return u.roleMapper.MapGroupsToRole(groups)
}

// generateToken generates a JWT and extracts the JTI
func (u *implUsecase) generateToken(ctx context.Context, usr *model.User, role string, groups []string) (string, string, error) {
	if u.jwtManager == nil {
		return "", "", fmt.Errorf("jwt manager not configured")
	}

	token, err := u.jwtManager.GenerateToken(usr.ID, usr.Email, role, groups)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.generateToken: %v", err)
		return "", "", err
	}

	claims, err := u.jwtManager.VerifyToken(token)
	if err != nil {
		return "", "", err
	}

	return token, claims.ID, nil
}

// createSession creates a session in Redis
func (u *implUsecase) createSession(ctx context.Context, userID, jti string, rememberMe bool) error {
	if u.sessionManager == nil {
		return nil
	}
	return u.sessionManager.CreateSession(ctx, userID, jti, rememberMe)
}

// recordFailedAttempt records a failed login attempt for rate limiting
func (u *implUsecase) recordFailedAttempt(ctx context.Context, ipAddress string) {
	if u.rateLimiter == nil {
		return
	}
	if err := u.rateLimiter.RecordFailedAttempt(ctx, ipAddress); err != nil {
		u.l.Warnf(ctx, "authentication.usecase.recordFailedAttempt: %v", err)
	}
}

// clearFailedAttempts clears failed login attempts on successful login
func (u *implUsecase) clearFailedAttempts(ctx context.Context, ipAddress string) {
	if u.rateLimiter == nil {
		return
	}
	if err := u.rateLimiter.ClearFailedAttempts(ctx, ipAddress); err != nil {
		u.l.Warnf(ctx, "authentication.usecase.clearFailedAttempts: %v", err)
	}
}

// revokeAllUserTokensInternal internal helper
func (u *implUsecase) revokeAllUserTokensInternal(ctx context.Context, userID string) error {
	if u.sessionManager == nil || u.blacklistManager == nil {
		return fmt.Errorf("session/blacklist manager not configured")
	}

	jtis, err := u.sessionManager.GetAllUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	expiresAt := u.clock().Add(7 * 24 * time.Hour)
	if err := u.blacklistManager.AddAllUserTokens(ctx, jtis, expiresAt); err != nil {
		return err
	}

	return u.sessionManager.DeleteUserSessions(ctx, userID)
}
