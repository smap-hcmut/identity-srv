package collector

import (
	"smap-project/config"
	"smap-project/pkg/log"
)

// NewClient creates a new Collector Service client.
func NewClient(cfg config.CollectorConfig, l log.Logger) Client {
	return newHTTPClient(cfg, l)
}
