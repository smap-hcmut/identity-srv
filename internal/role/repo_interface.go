package role

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

//go:generate mockery --name=Repository
type Repository interface {
	Create(ctx context.Context, sc models.Scope, opt CreateOptions) (models.Role, error)
	Detail(ctx context.Context, sc models.Scope, id string) (models.Role, error)
	Update(ctx context.Context, sc models.Scope, opt UpdateOptions) (models.Role, error)
	Delete(ctx context.Context, sc models.Scope, ids []string) error
	List(ctx context.Context, sc models.Scope, opt ListOptions) ([]models.Role, error)
	GetOne(ctx context.Context, sc models.Scope, opt GetOneOptions) (models.Role, error)
	Get(ctx context.Context, sc models.Scope, opt GetOptions) ([]models.Role, paginator.Paginator, error)
}
