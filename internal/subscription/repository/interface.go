package repository

import (
	"context"

	"smap-api/internal/model"
	"smap-api/pkg/paginator"
)

//go:generate mockery --name Repository
type Repository interface {
	Get(ctx context.Context, sc model.Scope, opts GetOptions) ([]model.Subscription, paginator.Paginator, error)
	Detail(ctx context.Context, sc model.Scope, id string) (model.Subscription, error)
	List(ctx context.Context, sc model.Scope, opts ListOptions) ([]model.Subscription, error)
	Create(ctx context.Context, sc model.Scope, opts CreateOptions) (model.Subscription, error)
	Update(ctx context.Context, sc model.Scope, opts UpdateOptions) (model.Subscription, error)
	GetOne(ctx context.Context, sc model.Scope, opts GetOneOptions) (model.Subscription, error)
	Delete(ctx context.Context, sc model.Scope, id string) error
	GetUserSubscription(ctx context.Context, sc model.Scope, userID string) (model.Subscription, error)
}
