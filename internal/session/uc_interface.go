package session

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/user"
)

//go:generate mockery --name UseCase
type UseCase interface {
	Create(ctx context.Context, sc models.Scope, ip CreateSessionInput) (CreateSessionOutput, error)
	SetUserUseCase(userUC user.UseCase)
}
