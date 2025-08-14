package consumer

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	productionMode = "production"
)

var (
	ginMode = gin.DebugMode
)

func (srv Consumer) Run() error {
	err := srv.mapHandlers()
	if err != nil {
		srv.l.Fatalf(context.Background(), "Failed to map handlers: %v", err)
		return err
	}

	return nil
}
