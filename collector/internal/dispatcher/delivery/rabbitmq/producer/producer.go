package producer

import (
	"context"
	"encoding/json"
	"errors"

	rabb "smap-collector/internal/dispatcher/delivery/rabbitmq"
	"smap-collector/internal/models"
	pkgRabbit "smap-collector/pkg/rabbitmq"
)

func (p implProducer) PublishTask(ctx context.Context, task models.CollectorTask) error {
	if p.writer == nil {
		return errors.New("producer not started")
	}

	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return p.writer.Publish(ctx, pkgRabbit.PublishArgs{
		Exchange:   rabb.TaskExchange.Name,
		RoutingKey: task.RoutingKey,
		Msg: pkgRabbit.Publishing{
			Body:        body,
			ContentType: pkgRabbit.ContentTypeJSON,
		},
	})
}
