package keyword

import (
	"context"
)

// KeywordSuggester defines the interface for suggesting keywords
type KeywordSuggester interface {
	Suggest(ctx context.Context, brandName string) ([]string, []string, error)
}

// MockKeywordSuggester is a mock implementation of KeywordSuggester
type MockKeywordSuggester struct{}

// NewMockKeywordSuggester creates a new MockKeywordSuggester
func NewMockKeywordSuggester() *MockKeywordSuggester {
	return &MockKeywordSuggester{}
}

// Suggest returns mock suggestions
func (m *MockKeywordSuggester) Suggest(ctx context.Context, brandName string) ([]string, []string, error) {
	// Mock logic: return some variations and negative keywords
	niche := []string{
		brandName + " review",
		brandName + " price",
		brandName + " specs",
		brandName + " problems",
	}
	negative := []string{
		"job",
		"hiring",
		"second hand",
		"used",
	}
	return niche, negative, nil
}
