package repository

import (
	"context"

	"identity-srv/internal/model"
)

//go:generate mockery --name Repository
type Repository interface {
	// OAuth user operations (only supported methods)
	Upsert(ctx context.Context, opts UpsertOptions) (model.User, error)
	Update(ctx context.Context, opts UpdateOptions) error
	Detail(ctx context.Context, opts DetailOptions) (model.User, error)
}
