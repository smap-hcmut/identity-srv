package producer

import (
	"fmt"

	rabb "gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
	rmqPkg "gitlab.com/gma-vietnam/tanca-connect/pkg/rabbitmq"
)

// Run runs the producer
func (p *implProducer) Run() (err error) {
	p.pushNotiWriter, err = p.getWriter(rabb.CreateNotificationExchange)
	if err != nil {
		fmt.Println("Error when getting writer")
		return
	}

	p.updateRequestEventIDWriter, err = p.getWriter(rabb.UpdateRequestEventIDExchange)
	if err != nil {
		fmt.Println("Error when getting writer")
		return
	}

	p.updateTaskEventIDWriter, err = p.getWriter(rabb.UpdateTaskEventIDExchange)
	if err != nil {
		fmt.Println("Error when getting writer")
		return
	}

	return
}

// Close closes the producer
func (p *implProducer) Close() {
	p.pushNotiWriter.Close()
	p.updateRequestEventIDWriter.Close()
	p.updateTaskEventIDWriter.Close()
}

func (p implProducer) getWriter(exchange rmqPkg.ExchangeArgs) (*rmqPkg.Channel, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		fmt.Println("Error when getting channel")
		return nil, err
	}

	err = ch.ExchangeDeclare(exchange)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (p implProducer) getWriterWithQueue(exchange rmqPkg.ExchangeArgs, queue rmqPkg.QueueArgs) (*rmqPkg.Channel, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		fmt.Println("Error when getting channel")
		return nil, err
	}

	err = ch.ExchangeDeclare(exchange)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(queue)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(rmqPkg.QueueBindArgs{
		Queue:    queue.Name,
		Exchange: exchange.Name,
	})
	if err != nil {
		return nil, err
	}

	return ch, nil
}
