package repository

import (
	"context"

	"smap-api/internal/model"
	"smap-api/pkg/paginator"
)

//go:generate mockery --name Repository
type Repository interface {
	Get(ctx context.Context, sc model.Scope, opts GetOptions) ([]model.Plan, paginator.Paginator, error)
	Detail(ctx context.Context, sc model.Scope, id string) (model.Plan, error)
	List(ctx context.Context, sc model.Scope, opts ListOptions) ([]model.Plan, error)
	Create(ctx context.Context, sc model.Scope, opts CreateOptions) (model.Plan, error)
	Update(ctx context.Context, sc model.Scope, opts UpdateOptions) (model.Plan, error)
	GetOne(ctx context.Context, sc model.Scope, opts GetOneOptions) (model.Plan, error)
	Delete(ctx context.Context, sc model.Scope, id string) error
}
