package usecase

import (
	"context"
	"errors"

	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
)

type taskProducer interface {
	PublishTask(ctx context.Context, task models.CollectorTask) error
}

type implUseCase struct {
	l              log.Logger
	prod           taskProducer
	defaultOptions dispatcher.Options
}

func NewUseCase(l log.Logger, prod taskProducer, opts dispatcher.Options) (dispatcher.UseCase, error) {
	if l == nil || prod == nil {
		return nil, errors.New("logger and producer are required")
	}

	if opts.DefaultMaxAttempts <= 0 {
		opts.DefaultMaxAttempts = 3
	}
	if opts.SchemaVersion <= 0 {
		opts.SchemaVersion = 1
	}

	return implUseCase{
		l:              l,
		prod:           prod,
		defaultOptions: opts,
	}, nil
}
