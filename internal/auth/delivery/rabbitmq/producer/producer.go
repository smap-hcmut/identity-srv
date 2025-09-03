package producer

import (
	"context"
	"encoding/json"

	rmqDelivery "github.com/nguyentantai21042004/smap-api/internal/auth/delivery/rabbitmq"
	"github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

func (p implProducer) PubSendEmail(ctx context.Context, msg rmqDelivery.SendEmailMsg) error {
	body, err := json.Marshal(msg)
	if err != nil {
		p.l.Errorf(ctx, "auth.delivery.rabbitmq.producer.PubSendEmail.json.Marshal: %v", err)
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
