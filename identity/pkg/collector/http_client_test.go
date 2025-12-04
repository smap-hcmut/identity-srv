package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"smap-api/config"
	"smap-api/pkg/log"

	"github.com/stretchr/testify/assert"
)

func newTestHTTPClient(serverURL string) *httpClient {
	cfg := config.CollectorConfig{
		BaseURL: serverURL,
		Timeout: 5,
	}
	return newHTTPClient(cfg, log.NewNopLogger())
}

func TestHTTPClient_DryRun_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, dryRunEndpoint, r.URL.Path)

		var req DryRunRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, []string{"kw1", "kw2"}, req.Keywords)
		assert.Equal(t, 10, req.Limit)

		resp := DryRunResponse{
			Posts: []Post{
				{ID: "post1"},
				{ID: "post2"},
			},
			TotalFound: 2,
			Limit:      10,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestHTTPClient(server.URL)
	posts, err := client.DryRun(context.Background(), []string{"kw1", "kw2"}, 10)

	assert.NoError(t, err)
	assert.Len(t, posts, 2)
	assert.Equal(t, "post1", posts[0].ID)
}

func TestHTTPClient_DryRun_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestHTTPClient(server.URL)
	client.client.Timeout = 5 * time.Millisecond

	_, err := client.DryRun(context.Background(), []string{"kw1"}, 10)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCollectorUnavailable)
}

func TestHTTPClient_DryRun_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := newTestHTTPClient(server.URL)
	_, err := client.DryRun(context.Background(), []string{"kw1"}, 10)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCollectorInvalidResponse)
}

func TestHTTPClient_DryRun_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestHTTPClient(server.URL)
	_, err := client.DryRun(context.Background(), []string{"kw1"}, 10)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCollectorUnavailable)
}
