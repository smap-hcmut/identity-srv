package google

import (
	"context"
	"fmt"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// Client wraps Google Directory API client
type Client struct {
	service    *directory.Service
	adminEmail string
	domain     string
}

// Config holds configuration for Google Directory API client
type Config struct {
	ServiceAccountKey string
	AdminEmail        string
	Domain            string
}

// New creates a new Google Directory API client
func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.ServiceAccountKey == "" {
		return nil, fmt.Errorf("service account key is required")
	}
	if cfg.AdminEmail == "" {
		return nil, fmt.Errorf("admin email is required")
	}
	if cfg.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	// Create service with service account credentials
	service, err := directory.NewService(ctx,
		option.WithCredentialsFile(cfg.ServiceAccountKey),
		option.WithScopes(directory.AdminDirectoryGroupReadonlyScope),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory service: %w", err)
	}

	return &Client{
		service:    service,
		adminEmail: cfg.AdminEmail,
		domain:     cfg.Domain,
	}, nil
}

// GetUserGroups fetches all groups for a user by email
func (c *Client) GetUserGroups(ctx context.Context, userEmail string) ([]string, error) {
	// Call Directory API to get user's groups
	groups, err := c.service.Groups.List().
		UserKey(userEmail).
		Domain(c.domain).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user groups: %w", err)
	}

	// Extract group emails
	groupEmails := make([]string, 0, len(groups.Groups))
	for _, group := range groups.Groups {
		groupEmails = append(groupEmails, group.Email)
	}

	return groupEmails, nil
}

// HealthCheck verifies connection to Google Directory API
func (c *Client) HealthCheck(ctx context.Context) error {
	// Try to list groups with limit 1 to verify connection
	_, err := c.service.Groups.List().
		Domain(c.domain).
		MaxResults(1).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("directory API health check failed: %w", err)
	}
	return nil
}
