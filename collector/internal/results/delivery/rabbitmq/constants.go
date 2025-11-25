package rabbitmq

// Exchange/queue cho fan-in kết quả và retry.
const (
	ResultExchangeName = "collector.results"
	// Routing pattern: crawler.<platform>.result
)
