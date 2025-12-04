package producer

import (
	"fmt"

	rabb "smap-project/internal/project/delivery/rabbitmq"
	rmqPkg "smap-project/pkg/rabbitmq"
)

// Run runs the producer
func (p *implProducer) Run() (err error) {
	p.dryRunWriter, err = p.getWriter(rabb.CollectorInboundExchange)
	if err != nil {
		fmt.Println("Error when getting dry-run writer")
		return
	}

	return
}

// Close closes the producer
func (p *implProducer) Close() {
	if p.dryRunWriter != nil {
		p.dryRunWriter.Close()
	}
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
