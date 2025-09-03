package producer

import (
	"context"

	rabb "github.com/nguyentantai21042004/smap-api/internal/auth/delivery/rabbitmq"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

//go:generate mockery --name=Producer
type Producer interface {
	PubSendEmail(ctx context.Context, msg rabb.SendEmailMsg) error
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

func NewProducer(l log.Logger, conn *rabbitmq.Connection) Producer {
	return &implProducer{
		l:    l,
		conn: conn,
	}
}
