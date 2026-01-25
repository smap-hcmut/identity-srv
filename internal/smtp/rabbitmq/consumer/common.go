package consumer

import (
	"context"

	rabbitmqPkg "smap-api/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type WorkerFunc func(msg amqp.Delivery)

func (c Consumer) consume(
	exchange rabbitmqPkg.ExchangeArgs,
	queueName string,
	workerFunc WorkerFunc,
) {
	ctx := context.Background()

	ch, err := c.conn.Channel()
	if err != nil {
		c.l.Errorf(ctx, "Failed to open channel: %v", err)
		panic(err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(exchange)
	if err != nil {
		c.l.Errorf(ctx, "Failed to declare exchange: %v", err)
		panic(err)
	}

	q, err := ch.QueueDeclare(rabbitmqPkg.QueueArgs{
		Name:    queueName,
		Durable: true,
	})
	if err != nil {
		c.l.Errorf(ctx, "Failed to declare queue: %v", err)
		panic(err)
	}

	err = ch.QueueBind(rabbitmqPkg.QueueBindArgs{
		Queue:    q.Name,
		Exchange: exchange.Name,
	})
	if err != nil {
		c.l.Errorf(ctx, "Failed to bind queue: %v", err)
		panic(err)
	}

	msgs, err := ch.Consume(rabbitmqPkg.ConsumeArgs{
		Queue: q.Name,
	})
	if err != nil {
		c.l.Errorf(ctx, "Failed to consume queue: %v", err)
		panic(err)
	}

	c.l.Infof(ctx, "Queue %s is being consumed", q.Name)

	forever := make(chan struct{})
	go func() {
		for msg := range msgs {
			workerFunc(msg)
		}
	}()

	<-forever
}
