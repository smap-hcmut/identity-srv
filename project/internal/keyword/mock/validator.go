package mock

import "context"

// MockValidator is a mock implementation of keyword.Validator
type MockValidator struct {
	ValidateFunc    func(ctx context.Context, keywords []string) ([]string, error)
	ValidateOneFunc func(ctx context.Context, keyword string) (string, error)
}

// Validate calls the mock function
func (m *MockValidator) Validate(ctx context.Context, keywords []string) ([]string, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx, keywords)
	}
	// Default mock behavior - return as is
	return keywords, nil
}

// ValidateOne calls the mock function
func (m *MockValidator) ValidateOne(ctx context.Context, keyword string) (string, error) {
	if m.ValidateOneFunc != nil {
		return m.ValidateOneFunc(ctx, keyword)
	}
	// Default mock behavior - return as is
	return keyword, nil
}
