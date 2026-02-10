package user

import (
	"context"

	"smap-api/internal/model"
)

//go:generate mockery --name UseCase
type UseCase interface {
	// OAuth user operations (only supported methods)
	Create(ctx context.Context, ip CreateInput) (model.User, error)
	Update(ctx context.Context, ip UpdateInput) error
	Detail(ctx context.Context, id string) (model.User, error)
}
