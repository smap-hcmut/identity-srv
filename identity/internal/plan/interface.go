package plan

import (
	"context"

	"smap-api/internal/model"
)

//go:generate mockery --name UseCase
type UseCase interface {
	Detail(ctx context.Context, sc model.Scope, id string) (PlanOutput, error)
	List(ctx context.Context, sc model.Scope, ip ListInput) ([]model.Plan, error)
	Get(ctx context.Context, sc model.Scope, ip GetInput) (GetPlanOutput, error)
	Create(ctx context.Context, sc model.Scope, ip CreateInput) (PlanOutput, error)
	GetOne(ctx context.Context, sc model.Scope, ip GetOneInput) (model.Plan, error)
	Update(ctx context.Context, sc model.Scope, ip UpdateInput) (PlanOutput, error)
	Delete(ctx context.Context, sc model.Scope, id string) error
}
