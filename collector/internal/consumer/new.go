package consumer

import (
	"errors"

	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

// Server gom toàn bộ wiring của service: RabbitMQ connection, dispatcher producer/usecase và inbound consumer.
type Server struct {
	l    pkgLog.Logger
	conn *rabbitmq.Connection
	cfg  Config
}

type Config struct {
	Logger            pkgLog.Logger
	AMQPConn          *rabbitmq.Connection
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
	}, nil
}
