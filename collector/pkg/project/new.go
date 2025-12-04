package project

import (
	"smap-collector/config"
	"smap-collector/pkg/log"
)

// NewClient creates a new Project Service webhook client.
func NewClient(cfg config.ProjectConfig, l log.Logger) Client {
	return newHTTPClient(cfg, l)
}
