package mock

import "context"

// MockSuggester is a mock implementation of keyword.Suggester
type MockSuggester struct {
	SuggestFunc func(ctx context.Context, brandName string) ([]string, []string, error)
}

// Suggest calls the mock function
func (m *MockSuggester) Suggest(ctx context.Context, brandName string) ([]string, []string, error) {
	if m.SuggestFunc != nil {
		return m.SuggestFunc(ctx, brandName)
	}
	// Default mock behavior
	return []string{brandName + " review"}, []string{"job", "hiring"}, nil
}
