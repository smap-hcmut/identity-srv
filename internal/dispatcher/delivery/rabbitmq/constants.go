package rabbitmq

// Exchange và queue cho ingress dispatcher.
const (
	InboundExchangeName = "collector.inbound"

	YouTubeQueueName = "crawler.youtube.queue"
	TikTokQueueName  = "crawler.tiktok.queue"
	// Mặc định routing pattern: crawler.<platform>.<task_type>
)
