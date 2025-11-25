package producer

import (
	"fmt"

	rabb "github.com/nguyentantai21042004/smap-api/internal/dispatcher/delivery/rabbitmq"
	pkgRabbit "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

// Run prepares writer channel.
func (p *implProducer) Run() (err error) {
	p.writer, err = p.getWriter(rabb.TaskExchange)
	if err != nil {
		return
	}
	return
}

// Close closes the producer.
func (p *implProducer) Close() {
	if p.writer != nil {
		p.writer.Close()
	}
}

func (p implProducer) getWriter(exchange pkgRabbit.ExchangeArgs) (*pkgRabbit.Channel, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		fmt.Println("Error when getting channel")
		return nil, err
	}

	if exchange.Name != "" {
		err = ch.ExchangeDeclare(exchange)
		if err != nil {
			return nil, err
		}
	}

	return ch, nil
}
