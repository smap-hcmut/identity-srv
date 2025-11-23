package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	pkgRabbit "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

type Producer struct {
	l     pkgLog.Logger
	conn  *pkgRabbit.Connection
	exchg string
}

type Config struct {
	Exchange string
}

func NewProducer(l pkgLog.Logger, conn *pkgRabbit.Connection, cfg Config) (*Producer, error) {
	if l == nil || conn == nil {
		return nil, fmt.Errorf("logger and connection are required")
	}

	p := &Producer{
		l:     l,
		conn:  conn,
		exchg: cfg.Exchange,
	}

	return p, nil
}

func (p *Producer) PublishTask(ctx context.Context, task models.CollectorTask) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	defer ch.Close()

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	pub := pkgRabbit.PublishArgs{
		Exchange:   p.exchg,
		RoutingKey: task.RoutingKey,
		Msg: pkgRabbit.Publishing{
			Body:         body,
			ContentType:  pkgRabbit.ContentTypeJSON,
			DeliveryMode: 2,
		},
	}

	if err := ch.Publish(ctx, pub); err != nil {
		return fmt.Errorf("publish task: %w", err)
	}

	return nil
}
