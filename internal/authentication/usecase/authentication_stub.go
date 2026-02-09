package usecase

import (
	"context"
	"fmt"
	"smap-api/internal/authentication"
	"smap-api/internal/model"
)

// Login - Stub implementation (will be replaced with OAuth2 in Task 1.3)
func (u *implUsecase) Login(ctx context.Context, sc model.Scope, ip authentication.LoginInput) (authentication.LoginOutput, error) {
	// TODO: Implement OAuth2 login flow in Task 1.3
	u.l.Warnf(ctx, "Login stub called - OAuth2 implementation pending (Task 1.3)")
	return authentication.LoginOutput{}, authentication.ErrNotImplemented
}

// Logout - Delete session from Redis
func (u *implUsecase) Logout(ctx context.Context, sc model.Scope) error {
	// TODO: Extract JTI from JWT claims in scope
	// For now, just log the logout
	u.l.Infof(ctx, "User %s logged out", sc.UserID)

	// TODO: Delete session from Redis using SessionManager (Task 1.8)
	// sessionManager.DeleteSession(ctx, jti)

	return nil
}

// GetCurrentUser - Extract user info from JWT claims
func (u *implUsecase) GetCurrentUser(ctx context.Context, sc model.Scope) (authentication.GetCurrentUserOutput, error) {
	if u.db == nil {
		u.l.Errorf(ctx, "Database connection not set in authentication usecase")
		return authentication.GetCurrentUserOutput{}, fmt.Errorf("database connection not available")
	}

	// Get user from database using UserID from scope (JWT claims)
	user, err := u.GetUserByIDDirect(ctx, u.db, sc.UserID)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.GetCurrentUser.GetUserByIDDirect: %v", err)
		return authentication.GetCurrentUserOutput{}, err
	}

	return authentication.GetCurrentUserOutput{
		User: *user,
	}, nil
}

// CreateOrUpdateUser creates a new user or updates existing user on OAuth2 login
func (u *implUsecase) CreateOrUpdateUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	if u.db == nil {
		u.l.Errorf(ctx, "Database connection not set in authentication usecase")
		return nil, fmt.Errorf("database connection not available")
	}
	return u.CreateOrUpdateUserDirect(ctx, u.db, email, name, avatarURL)
}

// UpdateUserRole updates the user's role in the database
func (u *implUsecase) UpdateUserRole(ctx context.Context, userID, role string) error {
	if u.db == nil {
		u.l.Errorf(ctx, "Database connection not set in authentication usecase")
		return fmt.Errorf("database connection not available")
	}
	return u.UpdateUserRoleDirect(ctx, u.db, userID, role)
}
