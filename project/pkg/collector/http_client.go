package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"smap-project/config"
	"smap-project/pkg/log"
)

const (
	dryRunEndpoint = "/api/v1/collector/dry-run"
)

type httpClient struct {
	client  *http.Client
	l       log.Logger
	cfg     config.CollectorConfig
	baseURL string
}

func newHTTPClient(cfg config.CollectorConfig, l log.Logger) *httpClient {
	return &httpClient{
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		l:       l,
		cfg:     cfg,
		baseURL: cfg.BaseURL,
	}
}

func (c *httpClient) DryRun(ctx context.Context, keywords []string, limit int) ([]Post, error) {
	reqBody := DryRunRequest{
		Keywords: keywords,
		Limit:    limit,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	url := c.baseURL + dryRunEndpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCollectorUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response body: %v", ErrCollectorInvalidResponse, err)
	}

	var dryRunResp DryRunResponse
	if err := json.Unmarshal(body, &dryRunResp); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal response: %v", ErrCollectorInvalidResponse, err)
	}

	return dryRunResp.Posts, nil
}

func (c *httpClient) handleErrorResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusRequestTimeout:
		return ErrCollectorTimeout
	default:
		return fmt.Errorf("%w: status %d", ErrCollectorUnavailable, resp.StatusCode)
	}
}
