package authentication

import (
	"context"
	"identity-srv/internal/audit"
	"identity-srv/internal/model"
)

// UseCase interface for authentication module
type UseCase interface {
	// User operations
	GetCurrentUser(ctx context.Context, sc model.Scope) (*model.User, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)

	// Session & Token operations
	Logout(ctx context.Context, sc model.Scope) error
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
	RevokeToken(ctx context.Context, jti string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// OAuth flow
	InitiateOAuthLogin(ctx context.Context, input OAuthLoginInput) (*OAuthLoginOutput, error)
	ProcessOAuthCallback(ctx context.Context, input OAuthCallbackInput) (*OAuthCallbackOutput, error)

	// Audit
	PublishAuditEvent(ctx context.Context, event audit.AuditEvent)
}
