package usecase

import (
	"context"

	"identity-srv/internal/model"
	"identity-srv/internal/user"
	"identity-srv/internal/user/repository"
)

// Create creates or updates user by email (for OAuth - uses Upsert internally)
func (u *usecase) Create(ctx context.Context, ip user.CreateInput) (model.User, error) {
	return u.repo.Upsert(ctx, repository.UpsertOptions{
		Email:     ip.Email,
		Name:      ip.Name,
		AvatarURL: ip.AvatarURL,
	})
}

// Update updates user role
func (u *usecase) Update(ctx context.Context, ip user.UpdateInput) error {
	return u.repo.Update(ctx, repository.UpdateOptions{
		UserID: ip.UserID,
		Role:   ip.Role,
	})
}

// Detail gets user by ID
func (u *usecase) Detail(ctx context.Context, id string) (model.User, error) {
	return u.repo.Detail(ctx, repository.DetailOptions{
		UserID: id,
	})
}
