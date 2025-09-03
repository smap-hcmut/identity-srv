package role

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name=UseCase
type UseCase interface {
	Create(ctx context.Context, sc models.Scope, input CreateInput) (CreateOutput, error)
	Detail(ctx context.Context, sc models.Scope, id string) (DetailOutput, error)
	Update(ctx context.Context, sc models.Scope, input UpdateInput) (UpdateOutput, error)
	Delete(ctx context.Context, sc models.Scope, ids []string) error
	List(ctx context.Context, sc models.Scope, ip ListInput) (ListOutput, error)
	GetOne(ctx context.Context, sc models.Scope, ip GetOneInput) (GetOneOutput, error)
	Get(ctx context.Context, sc models.Scope, ip GetInput) (GetOutput, error)
}
