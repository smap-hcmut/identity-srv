package usecase

import (
	"errors"

	"smap-collector/internal/dispatcher"
	"smap-collector/internal/dispatcher/delivery/rabbitmq/producer"
	"smap-collector/internal/models"
	"smap-collector/pkg/log"
)

type implUseCase struct {
	l              log.Logger
	prod           producer.Producer
	defaultOptions dispatcher.Options
}

func NewUseCase(l log.Logger, prod producer.Producer, opts dispatcher.Options) (dispatcher.UseCase, error) {
	if l == nil || prod == nil {
		return nil, errors.New("logger and producer are required")
	}

	if opts.DefaultMaxAttempts <= 0 {
		opts.DefaultMaxAttempts = 3
	}
	if opts.SchemaVersion <= 0 {
		opts.SchemaVersion = 1
	}
	if len(opts.PlatformQueues) == 0 {
		opts.PlatformQueues = map[models.Platform]string{
			models.PlatformYouTube: "crawler.youtube.queue",
			models.PlatformTikTok:  "crawler.tiktok.queue",
		}
	}

	return implUseCase{
		l:              l,
		prod:           prod,
		defaultOptions: opts,
	}, nil
}
