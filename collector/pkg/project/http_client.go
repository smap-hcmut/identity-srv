package project

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"smap-collector/config"
	"smap-collector/pkg/log"
)

const (
	dryRunCallbackEndpoint = "/internal/dryrun/callback"
)

type httpClient struct {
	client       *http.Client
	l            log.Logger
	cfg          config.ProjectConfig
	baseURL      string
	internalKey  string
	maxRetries   int
	initialDelay time.Duration
	maxDelay     time.Duration
}

func newHTTPClient(cfg config.ProjectConfig, l log.Logger) *httpClient {
	return &httpClient{
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		l:            l,
		cfg:          cfg,
		baseURL:      cfg.BaseURL,
		internalKey:  cfg.InternalKey,
		maxRetries:   cfg.WebhookRetryAttempts,
		initialDelay: time.Duration(cfg.WebhookRetryDelay) * time.Second,
		maxDelay:     32 * time.Second,
	}
}

func (c *httpClient) SendDryRunCallback(ctx context.Context, req CallbackRequest) error {
	delay := c.initialDelay

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err := c.doSend(ctx, req, attempt+1)
		if err == nil {
			return nil
		}

		if attempt < c.maxRetries {
			c.l.Warnf(ctx, "Webhook attempt %d/%d failed, retrying in %v: %v",
				attempt+1, c.maxRetries+1, delay, err)

			time.Sleep(delay)

			// Exponential backoff
			delay *= 2
			if delay > c.maxDelay {
				delay = c.maxDelay
			}
		} else {
			c.l.Errorf(ctx, "Webhook failed after %d attempts: %v", c.maxRetries+1, err)
		}
	}

	return fmt.Errorf("webhook failed after %d attempts", c.maxRetries+1)
}

func (c *httpClient) doSend(ctx context.Context, req CallbackRequest, attempt int) error {
	// Marshal request to JSON
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	url := c.baseURL + dryRunCallbackEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", c.internalKey)

	// Send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProjectUnavailable, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.l.Infof(ctx, "Webhook callback sent successfully: job_id=%s, platform=%s, status=%s, attempt=%d",
			req.JobID, req.Platform, req.Status, attempt)
		return nil
	}

	return c.handleErrorResponse(resp)
}

func (c *httpClient) handleErrorResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrProjectUnauthorized
	case http.StatusRequestTimeout:
		return ErrProjectTimeout
	case http.StatusBadRequest:
		// 4xx errors should not be retried
		return fmt.Errorf("webhook returned client error %d (not retrying)", resp.StatusCode)
	default:
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Other 4xx errors should not be retried
			return fmt.Errorf("webhook returned client error %d (not retrying)", resp.StatusCode)
		}
		// 5xx errors should be retried
		return fmt.Errorf("%w: status %d", ErrProjectUnavailable, resp.StatusCode)
	}
}
