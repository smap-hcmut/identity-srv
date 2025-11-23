package producer

import (
	context "context"
	"encoding/json"

	rabb "gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/rabbitmq"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
)

func (p implProducer) PublishPushNotiMsg(ctx context.Context, msg rabb.PushNotiMsg) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	util.PrintJson(msg)

	return p.pushNotiWriter.Publish(ctx, rabbitmq.PublishArgs{
		Exchange: rabb.CreateNotificationExchange.Name,
		Msg: rabbitmq.Publishing{
			Body:        body,
			ContentType: rabbitmq.ContentTypePlainText,
		},
	})
}

func (p implProducer) PublishUpdateRequestEventIDMsg(ctx context.Context, msg rabb.UpdateRequestEventIDMsg) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.updateRequestEventIDWriter.Publish(ctx, rabbitmq.PublishArgs{
		Exchange: rabb.UpdateRequestEventIDExchange.Name,
		Msg: rabbitmq.Publishing{
			Body:        body,
			ContentType: rabbitmq.ContentTypePlainText,
		},
	})
}

func (p implProducer) PublishUpdateTaskEventIDMsg(ctx context.Context, msg rabb.UpdateTaskEventIDMsg) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.updateTaskEventIDWriter.Publish(ctx, rabbitmq.PublishArgs{
		Exchange: rabb.UpdateTaskEventIDExchange.Name,
		Msg: rabbitmq.Publishing{
			Body:        body,
			ContentType: rabbitmq.ContentTypePlainText,
		},
	})
}
