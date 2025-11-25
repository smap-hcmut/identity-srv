package consumer

import (
	"context"

	dispatcherConsumer "smap-collector/internal/dispatcher/delivery/rabbitmq/consumer"
	dispatcherProducer "smap-collector/internal/dispatcher/delivery/rabbitmq/producer"
	dispatcherUsecase "smap-collector/internal/dispatcher/usecase"
)

func (srv *Server) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	srv.l.Info(ctx, "consumer starting")

	// 1. Init Producers
	prod := dispatcherProducer.New(srv.l, srv.conn)
	if err := prod.Run(); err != nil {
		return err
	}
	// TODO: Add Close() to cleanup if needed, though srv.Close() handles conn.

	// 2. Init UseCases
	uc, err := dispatcherUsecase.NewUseCase(srv.l, prod, srv.cfg.DispatcherOptions)
	if err != nil {
		return err
	}

	// 3. Init Consumers
	dispatchC := dispatcherConsumer.NewConsumer(srv.l, srv.conn, uc)

	// 4. Start Consumers
	dispatchC.Consume()

	// Block until context cancelled (caller should cancel via signal).
	<-ctx.Done()

	return nil
}

// Close releases MQ resources.
func (srv *Server) Close() {
	if srv.conn != nil {
		srv.conn.Close()
	}
}
