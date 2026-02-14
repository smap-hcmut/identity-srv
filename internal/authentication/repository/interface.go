package repository

import (
	"context"
	"smap-api/internal/model"
)

// Repository interface for authentication module
type Repository interface {
	// JWT Keys
	SaveKey(ctx context.Context, key *model.JWTKey) error
	GetActiveKey(ctx context.Context) (*model.JWTKey, error)
	GetActiveAndRotatingKeys(ctx context.Context) ([]*model.JWTKey, error)
	UpdateKeyStatus(ctx context.Context, kid, status string) error
	GetRotatingKeys(ctx context.Context) ([]*model.JWTKey, error)

	// User related (if needed, currently mostly handled by internal/user)
}
