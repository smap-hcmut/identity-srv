package collector

import "context"

// Client defines the interface for the Collector Service client.
type Client interface {
	// DryRun performs a dry run of the collector service.
	DryRun(ctx context.Context, keywords []string, limit int) ([]Post, error)
}
