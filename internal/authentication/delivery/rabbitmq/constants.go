package rabbitmq

import (
	"smap-api/pkg/rabbitmq"
)

const (
	SEND_EMAIL_EXCHANGE_NAME = "smtp_send_email_exc"
	SEND_EMAIL_QUEUE_NAME    = "smtp_send_email"
)

var (
	SendEmailExc = rabbitmq.ExchangeArgs{
		Name:       SEND_EMAIL_EXCHANGE_NAME,
		Type:       rabbitmq.ExchangeTypeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}
)
