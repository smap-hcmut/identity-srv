package postgres

import (
	"context"
	"smap-api/internal/model"
	"smap-api/internal/user/repository"
	"smap-api/pkg/paginator"
)

// Stub implementations for methods not needed in OAuth flow
// These will be properly implemented in later tasks

func (r *implRepository) Detail(ctx context.Context, sc model.Scope, id string) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}

func (r *implRepository) List(ctx context.Context, sc model.Scope, opts repository.ListOptions) ([]model.User, error) {
	return nil, repository.ErrNotFound
}

func (r *implRepository) Get(ctx context.Context, sc model.Scope, opts repository.GetOptions) ([]model.User, paginator.Paginator, error) {
	return nil, paginator.Paginator{}, repository.ErrNotFound
}

func (r *implRepository) Create(ctx context.Context, sc model.Scope, opts repository.CreateOptions) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}

func (r *implRepository) Update(ctx context.Context, sc model.Scope, opts repository.UpdateOptions) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}

func (r *implRepository) GetOne(ctx context.Context, sc model.Scope, opts repository.GetOneOptions) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}

func (r *implRepository) Delete(ctx context.Context, sc model.Scope, id string) error {
	return repository.ErrNotFound
}
