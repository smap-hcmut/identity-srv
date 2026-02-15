package usecase

import (
	"context"
	"fmt"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"
	"identity-srv/internal/user"
	"strings"
	"time"
)

// --- internal helpers (private, not exposed on the interface) ---

// generateState generates a CSRF state token
func (u *ImplUsecase) generateState() string {
	return fmt.Sprintf("state-%d", u.clock().UnixNano())
}

// isAllowedDomain checks if the email domain is in the allowlist
func (u *ImplUsecase) isAllowedDomain(email string) bool {
	if len(u.allowedDomains) == 0 {
		return true // No restrictions configured
	}
	domain := u.extractDomain(email)
	for _, d := range u.allowedDomains {
		if domain == d {
			return true
		}
	}
	return false
}

// isBlockedEmail checks if the email is in the blocklist
func (u *ImplUsecase) isBlockedEmail(email string) bool {
	for _, blocked := range u.blockedEmails {
		if email == blocked {
			return true
		}
	}
	return false
}

// extractDomain extracts domain from email address
func (u *ImplUsecase) extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// createOrUpdateUser creates or updates a user via the user UseCase
func (u *ImplUsecase) createOrUpdateUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	usr, err := u.userUC.Create(ctx, user.CreateInput{
		Email:     email,
		Name:      name,
		AvatarURL: avatarURL,
	})
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.createOrUpdateUser: %v", err)
		return nil, fmt.Errorf("%w: %v", authentication.ErrUserCreation, err)
	}
	u.l.Infof(ctx, "User created/updated: ID=%s Email=%s", usr.ID, usr.Email)
	return &usr, nil
}

// updateUserRole updates the user's role
func (u *ImplUsecase) updateUserRole(ctx context.Context, userID, role string) error {
	return u.userUC.Update(ctx, user.UpdateInput{
		UserID: userID,
		Role:   role,
	})
}

// mapEmailToRole maps email to a role using the role mapper
func (u *ImplUsecase) mapEmailToRole(email string) string {
	if u.roleMapper == nil {
		return "VIEWER"
	}
	return u.roleMapper.MapEmailToRole(email)
}

// generateToken generates a JWT and extracts the JTI
func (u *ImplUsecase) generateToken(ctx context.Context, usr *model.User, role string, groups []string) (string, string, error) {
	if u.jwtManager == nil {
		return "", "", fmt.Errorf("jwt manager not configured")
	}

	token, err := u.jwtManager.GenerateToken(usr.ID, usr.Email, role, groups)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.generateToken: %v", err)
		return "", "", err
	}
	u.l.Infof(ctx, "Token generated for user %s, verifying...", usr.ID)

	claims, err := u.jwtManager.VerifyToken(token)
	if err != nil {
		return "", "", err
	}

	return token, claims.ID, nil
}

// createSession creates a session in Redis
func (u *ImplUsecase) createSession(ctx context.Context, userID, jti string, rememberMe bool) error {
	if u.sessionManager == nil {
		return nil
	}
	return u.sessionManager.CreateSession(ctx, userID, jti, rememberMe)
}

// revokeAllUserTokensInternal internal helper
func (u *ImplUsecase) revokeAllUserTokensInternal(ctx context.Context, userID string) error {
	if u.sessionManager == nil || u.blacklistManager == nil {
		return authentication.ErrConfigurationMissing
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
