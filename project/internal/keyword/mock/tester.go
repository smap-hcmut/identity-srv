package mock

import "context"

// MockTester is a mock implementation of keyword.Tester
type MockTester struct {
	TestFunc func(ctx context.Context, keywords []string) ([]interface{}, error)
}

// Test calls the mock function
func (m *MockTester) Test(ctx context.Context, keywords []string) ([]interface{}, error) {
	if m.TestFunc != nil {
		return m.TestFunc(ctx, keywords)
	}
	// Default mock behavior
	return []interface{}{
		map[string]interface{}{
			"content": "Sample post",
			"source":  "mock",
		},
	}, nil
}
