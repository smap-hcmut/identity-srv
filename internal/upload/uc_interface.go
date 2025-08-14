package upload

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name UseCase
type UseCase interface {
	Create(ctx context.Context, sc models.Scope, ip CreateInput) (UploadOutput, error)
	Detail(ctx context.Context, sc models.Scope, ID string) (UploadOutput, error)
	Get(ctx context.Context, sc models.Scope, ip GetInput) (GetOutput, error)
}
