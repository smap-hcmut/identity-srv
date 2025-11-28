package rabbitmq

import pkgRabbit "smap-collector/pkg/rabbitmq"

const (
	// Inbound (Ingress)
	ExchangeInbound   = "collector.inbound"
	QueueInbound      = "collector.inbound.queue"
	RoutingKeyInbound = "crawler.#"

	// TikTok
	ExchangeTikTok   = "collector.tiktok"
	QueueTikTok      = "collector.tiktok.queue"
	RoutingKeyTikTok = "tiktok.task"

	// YouTube
	ExchangeYouTube   = "collector.youtube"
	QueueYouTube      = "collector.youtube.queue"
	RoutingKeyYouTube = "youtube.task"
)

var (
	InboundExchangeArgs = pkgRabbit.ExchangeArgs{
		Name:       ExchangeInbound,
		Type:       pkgRabbit.ExchangeTypeTopic,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	TikTokExchangeArgs = pkgRabbit.ExchangeArgs{
		Name:       ExchangeTikTok,
		Type:       pkgRabbit.ExchangeTypeDirect,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	YouTubeExchangeArgs = pkgRabbit.ExchangeArgs{
		Name:       ExchangeYouTube,
		Type:       pkgRabbit.ExchangeTypeDirect,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}
)
