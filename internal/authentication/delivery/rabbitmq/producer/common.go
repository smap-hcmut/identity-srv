package producer

import (
	rmqDelivery "smap-api/internal/authentication/delivery/rabbitmq"
	rmqPkg "smap-api/pkg/rabbitmq"
)

func (p *implProducer) Run() (err error) {
	if p.sendEmailWriter, err = p.getWriter(rmqDelivery.SendEmailExc); err != nil {
		return
	}

	return nil
}

// Close closes the producer
func (p *implProducer) Close() {

}

func (p implProducer) getWriter(exchange rmqPkg.ExchangeArgs) (*rmqPkg.Channel, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(exchange)
	if err != nil {
		return nil, err
	}

	return ch, nil
}
