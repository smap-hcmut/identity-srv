package authentication

import (
	"context"

	"smap-api/internal/audit"
	"smap-api/internal/model"
)

//go:generate mockery --name UseCase
type UseCase interface {
	// OAuth2 methods
	GetCurrentUser(ctx context.Context, sc model.Scope) (GetCurrentUserOutput, error)
	CreateOrUpdateUser(ctx context.Context, ip CreateOrUpdateUserInput) (model.User, error)
	UpdateUserRole(ctx context.Context, ip UpdateUserRoleInput) error
	Logout(ctx context.Context, sc model.Scope) error

	// Audit logging
	PublishAuditEvent(ctx context.Context, event audit.AuditEvent)
}
