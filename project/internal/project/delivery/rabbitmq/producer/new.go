package producer

import (
	context "context"

	"smap-project/internal/project/delivery/rabbitmq"
	"smap-project/pkg/log"
	rmq "smap-project/pkg/rabbitmq"
)

// Producer is an interface that represents a producer
//
//go:generate mockery --name=Producer
type Producer interface {
	PublishDryRunTask(ctx context.Context, msg rabbitmq.DryRunCrawlRequest) error
	// Run runs the producer
	Run() error
	// Close closes the producer
	Close()
}

type implProducer struct {
	l            log.Logger
	conn         rmq.Connection
	dryRunWriter *rmq.Channel
}

// New creates a new producer
func New(l log.Logger, conn rmq.Connection) Producer {
	return &implProducer{
		l:    l,
		conn: conn,
	}
}

// this is example code for rabbitmq constants rewrite with the project setup

// import (
// 	context "context"

// 	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
// 	"gitlab.com/gma-vietnam/tanca-connect/pkg/rabbitmq"

// 	rabb "gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
// )

// // Producer is a interface that represents a producer
// //
// //go:generate mockery --name=Producer
// type Producer interface {
// 	PublishPushNotiMsg(ctx context.Context, msg rabb.PushNotiMsg) error
// 	PublishUpdateRequestEventIDMsg(ctx context.Context, msg rabb.UpdateRequestEventIDMsg) error
// 	PublishUpdateTaskEventIDMsg(ctx context.Context, msg rabb.UpdateTaskEventIDMsg) error
// 	// Run runs the producer
// 	Run() error
// 	// Close closes the producer
// 	Close()
// }

// type implProducer struct {
// 	l                          log.Logger
// 	conn                       rabbitmq.Connection
// 	pushNotiWriter             *rabbitmq.Channel
// 	updateRequestEventIDWriter *rabbitmq.Channel
// 	updateTaskEventIDWriter    *rabbitmq.Channel
// }

// // New creates a new producer
// func New(l log.Logger, conn rabbitmq.Connection) Producer {
// 	return &implProducer{
// 		l:    l,
// 		conn: conn,
// 	}
// }
