package rabbitmq

import pkgRabbit "smap-collector/pkg/rabbitmq"

const (
	// Inbound (Ingress)
	ExchangeInbound   = "collector.inbound"
	QueueInbound      = "collector.inbound.queue"
	RoutingKeyInbound = "crawler.#"

	// TikTok
	ExchangeTikTok   = "tiktok_exchange"
	QueueTikTok      = "tiktok_crawl_queue"
	RoutingKeyTikTok = "tiktok.crawl"

	// YouTube
	ExchangeYouTube   = "youtube_exchange"
	QueueYouTube      = "youtube_crawl_queue"
	RoutingKeyYouTube = "youtube.crawl"
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
