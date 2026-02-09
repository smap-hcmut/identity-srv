package authentication

import (
	"context"

	"smap-api/internal/audit"
	"smap-api/internal/model"
)

//go:generate mockery --name UseCase
type UseCase interface {
	// Legacy methods (will be removed)
	Login(ctx context.Context, sc model.Scope, ip LoginInput) (LoginOutput, error)
	Logout(ctx context.Context, sc model.Scope) error
	GetCurrentUser(ctx context.Context, sc model.Scope) (GetCurrentUserOutput, error)

	// New OAuth2 methods
	CreateOrUpdateUser(ctx context.Context, email, name, avatarURL string) (*model.User, error)
	UpdateUserRole(ctx context.Context, userID, role string) error

	// Audit logging
	PublishAuditEvent(ctx context.Context, event audit.AuditEvent)
}
