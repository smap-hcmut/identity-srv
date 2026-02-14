package usecase

import (
	"context"
	"fmt"
	"smap-api/internal/audit"
	"smap-api/internal/authentication"
	"smap-api/internal/model"
)

// GetCurrentUser gets current user from scope
func (u *ImplUsecase) GetCurrentUser(ctx context.Context, sc model.Scope) (*model.User, error) {
	usr, err := u.userUC.Detail(ctx, sc.UserID)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.GetCurrentUser.Detail: %v", err)
		return nil, authentication.ErrUserNotFound
	}
	return &usr, nil
}

// GetUserByID gets a user by ID (for internal service calls)
func (u *ImplUsecase) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	usr, err := u.userUC.Detail(ctx, userID)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.GetUserByID.Detail: %v", err)
		return nil, authentication.ErrUserNotFound
	}
	return &usr, nil
}

// Logout invalidates the current session
func (u *ImplUsecase) Logout(ctx context.Context, sc model.Scope) error {
	if u.sessionManager == nil {
		return nil
	}

	if err := u.sessionManager.DeleteSession(ctx, sc.JTI); err != nil {
		u.l.Errorf(ctx, "authentication.usecase.Logout.DeleteSession: %v", err)
		return err
	}

	u.PublishAuditEvent(ctx, audit.AuditEvent{
		UserID:       sc.UserID,
		Action:       audit.ActionLogout,
		ResourceType: "authentication",
	})

	return nil
}

// ValidateToken verifies a JWT token
func (u *ImplUsecase) ValidateToken(ctx context.Context, token string) (*authentication.TokenValidationResult, error) {
	if u.jwtManager == nil {
		return nil, fmt.Errorf("jwt manager not configured")
	}

	claims, err := u.jwtManager.VerifyToken(token)
	if err != nil {
		return &authentication.TokenValidationResult{Valid: false}, nil
	}

	if u.blacklistManager != nil {
		isBlacklisted, err := u.blacklistManager.IsBlacklisted(ctx, claims.ID)
		if err != nil {
			u.l.Errorf(ctx, "authentication.usecase.ValidateToken.IsBlacklisted: %v", err)
			return nil, err
		}
		if isBlacklisted {
			return &authentication.TokenValidationResult{Valid: false}, nil
		}
	}

	return &authentication.TokenValidationResult{
		Valid:     true,
		UserID:    claims.Subject,
		Email:     claims.Email,
		Role:      claims.Role,
		Groups:    claims.Groups,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// RevokeToken revokes a specific token
func (u *ImplUsecase) RevokeToken(ctx context.Context, jti string) error {
	if u.sessionManager == nil || u.blacklistManager == nil {
		return authentication.ErrConfigurationMissing
	}

	session, err := u.sessionManager.GetSession(ctx, jti)
	if err != nil {
		return err
	}

	if err := u.blacklistManager.AddToken(ctx, jti, session.ExpiresAt); err != nil {
		return err
	}

	return u.sessionManager.DeleteSession(ctx, jti)
}

// RevokeAllUserTokens revokes all tokens for a user
func (u *ImplUsecase) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return u.revokeAllUserTokensInternal(ctx, userID)
}

// PublishAuditEvent publishes an audit event (non-blocking)
func (u *ImplUsecase) PublishAuditEvent(ctx context.Context, event audit.AuditEvent) {
	if u.auditPublisher == nil {
		u.l.Warnf(ctx, "Audit publisher not configured, skipping audit event")
		return
	}

	if err := u.auditPublisher.Publish(ctx, event); err != nil {
		u.l.Errorf(ctx, "Failed to publish audit event: %v", err)
	}
}
