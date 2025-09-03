package user

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name UseCase
type UseCase interface {
	Get(ctx context.Context, sc models.Scope, ip GetInput) (GetUserOutput, error)
	GetOne(ctx context.Context, sc models.Scope, ip GetOneInput) (models.User, error)
	Detail(ctx context.Context, sc models.Scope, ID string) (UserOutput, error)
	DetailMe(ctx context.Context, sc models.Scope) (UserOutput, error)
	Create(ctx context.Context, sc models.Scope, ip CreateInput) (UserOutput, error)
	UpdateVerified(ctx context.Context, sc models.Scope, ip UpdateVerifiedInput) (UserOutput, error)
	UpdateAvatar(ctx context.Context, sc models.Scope, ip UpdateAvatarInput) error
	Delete(ctx context.Context, sc models.Scope, ids []string) error
}
