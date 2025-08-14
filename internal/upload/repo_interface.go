package upload

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

//go:generate mockery --name Repository
type Repository interface {
	Create(ctx context.Context, sc models.Scope, opts CreateOptions) (models.Upload, error)
	Detail(ctx context.Context, sc models.Scope, ID string) (models.Upload, error)
	Get(ctx context.Context, sc models.Scope, opts GetOptions) ([]models.Upload, paginator.Paginator, error)
}
