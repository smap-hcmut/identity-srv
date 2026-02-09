package usecase

import (
	"context"
	"smap-api/internal/model"
	"smap-api/internal/user"
)

// Stub implementations for methods not needed in OAuth flow
// These will be properly implemented in later tasks

func (uc *usecase) Detail(ctx context.Context, sc model.Scope, id string) (user.UserOutput, error) {
	return user.UserOutput{}, user.ErrNotImplemented
}

func (uc *usecase) DetailMe(ctx context.Context, sc model.Scope) (user.UserOutput, error) {
	return user.UserOutput{}, user.ErrNotImplemented
}

func (uc *usecase) List(ctx context.Context, sc model.Scope, ip user.ListInput) ([]model.User, error) {
	return nil, user.ErrNotImplemented
}

func (uc *usecase) Get(ctx context.Context, sc model.Scope, ip user.GetInput) (user.GetUserOutput, error) {
	return user.GetUserOutput{}, user.ErrNotImplemented
}

func (uc *usecase) UpdateProfile(ctx context.Context, sc model.Scope, ip user.UpdateProfileInput) (user.UserOutput, error) {
	return user.UserOutput{}, user.ErrNotImplemented
}

func (uc *usecase) ChangePassword(ctx context.Context, sc model.Scope, ip user.ChangePasswordInput) error {
	return user.ErrNotImplemented
}

func (uc *usecase) Create(ctx context.Context, sc model.Scope, ip user.CreateInput) (user.UserOutput, error) {
	return user.UserOutput{}, user.ErrNotImplemented
}

func (uc *usecase) GetOne(ctx context.Context, sc model.Scope, ip user.GetOneInput) (model.User, error) {
	return model.User{}, user.ErrNotImplemented
}

func (uc *usecase) Update(ctx context.Context, sc model.Scope, ip user.UpdateInput) (user.UserOutput, error) {
	return user.UserOutput{}, user.ErrNotImplemented
}

func (uc *usecase) Delete(ctx context.Context, sc model.Scope, id string) error {
	return user.ErrNotImplemented
}

func (uc *usecase) Dashboard(ctx context.Context, sc model.Scope, ip user.DashboardInput) (user.UsersDashboardOutput, error) {
	return user.UsersDashboardOutput{}, user.ErrNotImplemented
}
