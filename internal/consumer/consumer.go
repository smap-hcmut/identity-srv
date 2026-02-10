package consumer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	pkgKafka "smap-api/pkg/kafka"
	pkgLog "smap-api/pkg/log"

	"github.com/IBM/sarama"
)

// ModuleConsumer represents a consumer for a specific module
type ModuleConsumer interface {
	GetTopics() []string
	GetGroupID() string
	CreateHandler() sarama.ConsumerGroupHandler
}

// Consumer represents the consumer service that manages multiple module consumers
type Consumer struct {
	logger          pkgLog.Logger
	postgresDB      *sql.DB
	kafkaBrokers    []string
	kafkaConsumers  map[string]sarama.ConsumerGroup // groupID -> consumer
	moduleConsumers []ModuleConsumer
	cancelFuncs     []context.CancelFunc
	wg              sync.WaitGroup
}

// Config holds consumer service configuration
type Config struct {
	Logger       pkgLog.Logger
	PostgresDB   *sql.DB
	KafkaBrokers []string
}

// New creates a new consumer service
func New(logger pkgLog.Logger, cfg Config) (*Consumer, error) {
	srv := &Consumer{
		logger:          logger,
		postgresDB:      cfg.PostgresDB,
		kafkaBrokers:    cfg.KafkaBrokers,
		kafkaConsumers:  make(map[string]sarama.ConsumerGroup),
		moduleConsumers: make([]ModuleConsumer, 0),
		cancelFuncs:     make([]context.CancelFunc, 0),
	}

	if err := srv.validate(); err != nil {
		return nil, err
	}

	return srv, nil
}

// validate validates that all required dependencies are provided
func (srv *Consumer) validate() error {
	if srv.logger == nil {
		return errors.New("logger is required")
	}
	if srv.postgresDB == nil {
		return errors.New("postgresDB is required")
	}
	if len(srv.kafkaBrokers) == 0 {
		return errors.New("kafkaBrokers is required")
	}
	return nil
}

// registerConsumers registers all module consumers (similar to mapHandlers in httpserver)
func (srv *Consumer) registerConsumers() error {
	ctx := context.Background()

	// Import and register module consumers here
	// This is where we initialize each module's consumer
	// Similar to how httpserver/handler.go initializes handlers

	if err := srv.registerAuditConsumer(ctx); err != nil {
		return fmt.Errorf("failed to register audit consumer: %w", err)
	}

	// TODO: Register other module consumers here
	// if err := srv.registerNotificationConsumer(ctx); err != nil {
	//     return fmt.Errorf("failed to register notification consumer: %w", err)
	// }

	return nil
}

// Start starts all registered consumers
func (srv *Consumer) Start(ctx context.Context) error {
	// Register all consumers first
	if err := srv.registerConsumers(); err != nil {
		return err
	}

	if len(srv.moduleConsumers) == 0 {
		return fmt.Errorf("no consumers registered")
	}

	srv.logger.Infof(ctx, "Starting %d consumer(s)...", len(srv.moduleConsumers))

	// Start each module consumer in its own goroutine
	for _, moduleConsumer := range srv.moduleConsumers {
		if err := srv.startModuleConsumer(ctx, moduleConsumer); err != nil {
			return fmt.Errorf("failed to start consumer %s: %w", moduleConsumer.GetGroupID(), err)
		}
	}

	srv.logger.Info(ctx, "All consumers started successfully")

	return nil
}

// register is an internal method to register a module consumer
func (srv *Consumer) register(moduleConsumer ModuleConsumer) error {
	ctx := context.Background()

	groupID := moduleConsumer.GetGroupID()
	topics := moduleConsumer.GetTopics()

	srv.logger.Infof(ctx, "Registering consumer: group=%s, topics=%v", groupID, topics)

	srv.moduleConsumers = append(srv.moduleConsumers, moduleConsumer)

	return nil
}

// startModuleConsumer starts a single module consumer
func (srv *Consumer) startModuleConsumer(parentCtx context.Context, moduleConsumer ModuleConsumer) error {
	groupID := moduleConsumer.GetGroupID()
	topics := moduleConsumer.GetTopics()

	// Create Kafka consumer group
	kafkaConsumer, err := pkgKafka.NewConsumerGroup(pkgKafka.ConsumerConfig{
		Brokers: srv.kafkaBrokers,
		GroupID: groupID,
	})
	if err != nil {
		return fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	srv.kafkaConsumers[groupID] = kafkaConsumer
	srv.logger.Infof(parentCtx, "Kafka consumer created: group=%s, brokers=%v", groupID, srv.kafkaBrokers)

	// Create context for this consumer
	ctx, cancel := context.WithCancel(parentCtx)
	srv.cancelFuncs = append(srv.cancelFuncs, cancel)

	// Start consuming in goroutine
	srv.wg.Add(1)
	go func() {
		defer srv.wg.Done()

		handler := moduleConsumer.CreateHandler()

		for {
			if err := kafkaConsumer.Consume(ctx, topics, handler); err != nil {
				srv.logger.Errorf(ctx, "Consumer error (group=%s): %v", groupID, err)
			}

			// Check if context was cancelled
			if ctx.Err() != nil {
				srv.logger.Infof(ctx, "Consumer stopped: group=%s", groupID)
				return
			}
		}
	}()

	srv.logger.Infof(parentCtx, "Consumer started: group=%s, topics=%v", groupID, topics)

	return nil
}

// Close closes all consumers
func (srv *Consumer) Close() error {
	ctx := context.Background()
	srv.logger.Info(ctx, "Closing all consumers...")

	// Cancel all contexts
	for _, cancel := range srv.cancelFuncs {
		cancel()
	}

	// Wait for all goroutines to finish
	srv.wg.Wait()

	// Close all Kafka consumers
	for groupID, kafkaConsumer := range srv.kafkaConsumers {
		if err := kafkaConsumer.Close(); err != nil {
			srv.logger.Errorf(ctx, "Failed to close consumer %s: %v", groupID, err)
		}
	}

	srv.logger.Info(ctx, "All consumers closed")

	return nil
}
