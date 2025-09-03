package producer

import (
	rmqDelivery "github.com/nguyentantai21042004/smap-api/internal/auth/delivery/rabbitmq"
	rmqPkg "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
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
