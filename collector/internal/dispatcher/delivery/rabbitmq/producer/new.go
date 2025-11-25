package producer

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	pkgRabbit "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

// Producer is a interface that represents a producer.
//
//go:generate mockery --name=Producer
type Producer interface {
	PublishTask(ctx context.Context, task models.CollectorTask) error
	// Run chuẩn bị writer/publisher.
	Run() error
	// Close đóng tài nguyên MQ.
	Close()
}

type implProducer struct {
	l      pkgLog.Logger
	conn   *pkgRabbit.Connection
	writer *pkgRabbit.Channel
}

// New creates a new producer.
func New(l pkgLog.Logger, conn *pkgRabbit.Connection) Producer {
	return &implProducer{
		l:    l,
		conn: conn,
	}
}
