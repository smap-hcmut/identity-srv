package subscription

import (
	"context"

	"smap-api/internal/model"
)

//go:generate mockery --name UseCase
type UseCase interface {
	Detail(ctx context.Context, sc model.Scope, id string) (SubscriptionOutput, error)
	List(ctx context.Context, sc model.Scope, ip ListInput) ([]model.Subscription, error)
	Get(ctx context.Context, sc model.Scope, ip GetInput) (GetSubscriptionOutput, error)
	Create(ctx context.Context, sc model.Scope, ip CreateInput) (SubscriptionOutput, error)
	GetOne(ctx context.Context, sc model.Scope, ip GetOneInput) (model.Subscription, error)
	Update(ctx context.Context, sc model.Scope, ip UpdateInput) (SubscriptionOutput, error)
	Delete(ctx context.Context, sc model.Scope, id string) error
	GetActiveSubscription(ctx context.Context, sc model.Scope, userID string) (model.Subscription, error)
	Cancel(ctx context.Context, sc model.Scope, id string) (SubscriptionOutput, error)
}

