package consumer

import (
	"errors"

	"smap-collector/internal/dispatcher"
	"smap-collector/pkg/discord"
	pkgLog "smap-collector/pkg/log"
	"smap-collector/pkg/rabbitmq"
)

type Server struct {
	l    pkgLog.Logger
	conn *rabbitmq.Connection
	cfg  Config
	disc *discord.DiscordWebhook
}

type Config struct {
	Logger            pkgLog.Logger
	AMQPConn          *rabbitmq.Connection
	Discord           *discord.DiscordWebhook
	DispatcherOptions dispatcher.Options
}

func New(cfg Config) (*Server, error) {
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.AMQPConn == nil {
		return nil, errors.New("amqp connection is required")
	}

	return &Server{
		l:    cfg.Logger,
		conn: cfg.AMQPConn,
		cfg:  cfg,
		disc: cfg.Discord,
	}, nil
}
