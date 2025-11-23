package consumer

import (
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/rabbitmq"
)

// Consumer represents a consumer

type Consumer struct {
	l    log.Logger
	conn *rabbitmq.Connection
	uc   event.UseCase
}

// NewConsumer creates a new consumer
func NewConsumer(l log.Logger, conn *rabbitmq.Connection, uc event.UseCase) Consumer {
	return Consumer{
		l:    l,
		conn: conn,
		uc:   uc,
	}
}
