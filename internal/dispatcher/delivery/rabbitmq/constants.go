package rabbitmq

import pkgRabbit "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"

const (
	InboundExchangeName = "collector.inbound" // upstream ingress (collector consumes)
	TaskExchangeName    = ""                  // default exchange publish trực tiếp queue

	InboundQueueName      = "collector.inbound.queue"
	InboundRoutingPattern = "crawler.#"
)

var (
	InboundExchange = pkgRabbit.ExchangeArgs{
		Name:       InboundExchangeName,
		Type:       pkgRabbit.ExchangeTypeTopic,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	TaskExchange = pkgRabbit.ExchangeArgs{
		Name:       TaskExchangeName,
		Type:       pkgRabbit.ExchangeTypeDirect,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}
)
