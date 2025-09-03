package user

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

//go:generate mockery --name Repository
type Repository interface {
	Get(ctx context.Context, sc models.Scope, opts GetOptions) ([]models.User, paginator.Paginator, error)
	List(ctx context.Context, sc models.Scope, opts ListOptions) ([]models.User, error)
	Detail(ctx context.Context, sc models.Scope, ID string) (models.User, error)
	GetOne(ctx context.Context, sc models.Scope, opts GetOneOptions) (models.User, error)
	Create(ctx context.Context, sc models.Scope, opts CreateOptions) (models.User, error)
	UpdateVerified(ctx context.Context, sc models.Scope, opts UpdateVerifiedOptions) (models.User, error)
	UpdateAvatar(ctx context.Context, sc models.Scope, opts UpdateAvatarOptions) (models.User, error)
	Delete(ctx context.Context, sc models.Scope, ids []string) error
}
