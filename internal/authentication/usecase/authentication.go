package usecase

import (
	"context"

	"smap-api/internal/authentication"
	"smap-api/internal/model"
	"smap-api/internal/user"
)

// GetCurrentUser - Extract user info from JWT claims
func (u *implUsecase) GetCurrentUser(ctx context.Context, sc model.Scope) (authentication.GetCurrentUserOutput, error) {
	// Get user from user usecase
	user, err := u.userUC.Detail(ctx, sc.UserID)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.GetCurrentUser: %v", err)
		return authentication.GetCurrentUserOutput{}, err
	}

	return authentication.GetCurrentUserOutput{
		User: user,
	}, nil
}

// CreateOrUpdateUser creates a new user or updates existing user on OAuth2 login
func (u *implUsecase) CreateOrUpdateUser(ctx context.Context, ip authentication.CreateOrUpdateUserInput) (model.User, error) {
	return u.userUC.Create(ctx, user.CreateInput{
		Email:     ip.Email,
		Name:      ip.Name,
		AvatarURL: ip.AvatarURL,
	})
}

// UpdateUserRole updates the user's role in the database
func (u *implUsecase) UpdateUserRole(ctx context.Context, ip authentication.UpdateUserRoleInput) error {
	return u.userUC.Update(ctx, user.UpdateInput{
		UserID: ip.UserID,
		Role:   ip.Role,
	})
}

// Logout - Delete session from Redis
func (u *implUsecase) Logout(ctx context.Context, sc model.Scope) error {
	u.l.Infof(ctx, "User %s logged out", sc.UserID)
	return nil
}
