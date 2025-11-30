package usecase

import (
	"context"
	"testing"
	"time"

	"smap-project/pkg/log"

	"github.com/stretchr/testify/assert"
)

type mockAmbiguousLLMProvider struct {
	mockLLMProvider
	ambiguousKeyword string
	ambiguousContext string
}

func (m *mockAmbiguousLLMProvider) CheckAmbiguity(ctx context.Context, keyword string) (bool, string, error) {
	if keyword == m.ambiguousKeyword {
		return true, m.ambiguousContext, nil
	}
	return false, "", nil
}

func TestValidator_validateOne_AmbiguityCheck(t *testing.T) {
	llmProvider := &mockAmbiguousLLMProvider{
		ambiguousKeyword: "apple",
		ambiguousContext: "fruit or tech",
	}
	uc := &usecase{
		l:           log.NewNopLogger(),
		llmProvider: llmProvider,
		clock:       time.Now,
	}

	// Test with ambiguous keyword
	_, err := uc.validateOne(context.Background(), "apple")
	assert.NoError(t, err) // Should not return an error, just a warning

	// Test with non-ambiguous keyword
	_, err = uc.validateOne(context.Background(), "vinfast")
	assert.NoError(t, err)

	// Test with multi-word keyword (should skip ambiguity check)
	_, err = uc.validateOne(context.Background(), "vinfast vf9")
	assert.NoError(t, err)
}
