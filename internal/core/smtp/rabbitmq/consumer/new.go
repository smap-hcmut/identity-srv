package consumer

import (

	rabbitmqPkg "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/internal/core/smtp"
)

type Consumer struct {
	l    log.Logger
	conn *rabbitmqPkg.Connection
	uc   smtp.UseCase
}

func NewConsumer(l log.Logger, conn *rabbitmqPkg.Connection, uc smtp.UseCase) Consumer {
	return Consumer{
		l:    l,
		conn: conn,
		uc:   uc,
	}
}
