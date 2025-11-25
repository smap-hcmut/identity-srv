package consumer

import (
	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	pkgRabbit "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

type Consumer struct {
	l    pkgLog.Logger
	conn *pkgRabbit.Connection
	uc   dispatcher.UseCase
}

// NewConsumer creates a new consumer.
func NewConsumer(l pkgLog.Logger, conn *pkgRabbit.Connection, uc dispatcher.UseCase) Consumer {
	return Consumer{
		l:    l,
		conn: conn,
		uc:   uc,
	}
}
