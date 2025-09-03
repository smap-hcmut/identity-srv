package session

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name Repository
type Repository interface {
	Create(ctx context.Context, sc models.Scope, ip CreateSessionOptions) (models.Session, error)
}
