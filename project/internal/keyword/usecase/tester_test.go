package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"smap-project/pkg/collector"
	"smap-project/pkg/log"

	"github.com/stretchr/testify/assert"
)

type mockCollectorClient struct {
	shouldError bool
}

func (m *mockCollectorClient) DryRun(ctx context.Context, keywords []string, limit int) ([]collector.Post, error) {
	if m.shouldError {
		return nil, errors.New("collector error")
	}
	return []collector.Post{{ID: "post1"}}, nil
}

func TestTester_test_Success(t *testing.T) {
	uc := &usecase{
		l:               log.NewNopLogger(),
		collectorClient: &mockCollectorClient{shouldError: false},
		llmProvider:     &mockLLMProvider{},
		clock:           time.Now,
	}

	posts, err := uc.test(context.Background(), []string{"keyword1"})

	assert.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, collector.Post{ID: "post1"}, posts[0])
}

func TestTester_test_CollectorFailure(t *testing.T) {
	uc := &usecase{
		l:               log.NewNopLogger(),
		collectorClient: &mockCollectorClient{shouldError: true},
		llmProvider:     &mockLLMProvider{},
		clock:           time.Now,
	}

	_, err := uc.test(context.Background(), []string{"keyword1"})

	assert.Error(t, err)
	assert.Equal(t, "collector error", err.Error())
}

func TestTester_test_ValidationFailure(t *testing.T) {
	uc := &usecase{
		l:               log.NewNopLogger(),
		collectorClient: &mockCollectorClient{},
		llmProvider:     &mockLLMProvider{},
		clock:           time.Now,
	}

	_, err := uc.test(context.Background(), []string{"buy"}) // "buy" is a stopword

	assert.Error(t, err)
}
