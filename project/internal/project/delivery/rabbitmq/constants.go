package rabbitmq

import "smap-project/pkg/rabbitmq"

const (
	CollectorInboundExchangeName = "collector.inbound"
	DryRunKeywordRoutingKey      = "crawler.dryrun_keyword"
)

var (
	CollectorInboundExchange = rabbitmq.ExchangeArgs{
		Name:       CollectorInboundExchangeName,
		Type:       "topic",
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}
)
