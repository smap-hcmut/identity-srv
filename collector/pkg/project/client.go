package project

import "context"

// Client defines the interface for the Project Service webhook client.
type Client interface {
	// SendDryRunCallback sends a dry-run webhook callback to the Project Service.
	SendDryRunCallback(ctx context.Context, req CallbackRequest) error
}
