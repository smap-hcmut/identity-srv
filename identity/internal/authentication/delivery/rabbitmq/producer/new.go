package producer

import (
	"context"

	rabb "smap-api/internal/authentication/delivery/rabbitmq"
	"smap-api/pkg/log"
	"smap-api/pkg/rabbitmq"
)

//go:generate mockery --name=Producer
type Producer interface {
	PublishSendEmail(ctx context.Context, msg rabb.PublishSendEmailMsg) error
	// Run runs the producer
	Run() error
	// Close closes the producer
	Close()
}

type implProducer struct {
	l               log.Logger
	conn            *rabbitmq.Connection
	sendEmailWriter *rabbitmq.Channel
}

var _ Producer = &implProducer{}

func New(l log.Logger, conn *rabbitmq.Connection) Producer {
	return &implProducer{
		l:    l,
		conn: conn,
	}
}
