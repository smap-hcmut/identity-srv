package consumer

import (
	"smap-api/internal/smtp"
	pkgLog "smap-api/pkg/log"
	rabbitmqPkg "smap-api/pkg/rabbitmq"
)

type Consumer struct {
	l    pkgLog.Logger
	conn *rabbitmqPkg.Connection
	uc   smtp.UseCase
}

func NewConsumer(l pkgLog.Logger, conn *rabbitmqPkg.Connection, uc smtp.UseCase) Consumer {
	return Consumer{
		l:    l,
		conn: conn,
		uc:   uc,
	}
}
