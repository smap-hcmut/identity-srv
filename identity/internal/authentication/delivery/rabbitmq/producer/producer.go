package producer

import (
	"context"
	"encoding/json"

	rmqDelivery "smap-api/internal/authentication/delivery/rabbitmq"
	"smap-api/pkg/rabbitmq"
)

func (p implProducer) PublishSendEmail(ctx context.Context, msg rmqDelivery.PublishSendEmailMsg) error {
	body, err := json.Marshal(msg)
	if err != nil {
		p.l.Errorf(ctx, "authentication.delivery.rabbitmq.producer.PublishSendEmail.json.Marshal: %v", err)
		return err
	}

	return p.sendEmailWriter.Publish(ctx, rabbitmq.PublishArgs{
		Exchange: rmqDelivery.SendEmailExc.Name,
		Msg: rabbitmq.Publishing{
			Body:        body,
			ContentType: rabbitmq.ContentTypePlainText,
		},
	})
}
