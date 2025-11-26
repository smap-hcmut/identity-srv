package keyword

import (
	"context"
	"testing"
)

func TestService_Suggest(t *testing.T) {
	ctx := context.Background()

	// Create mock suggester
	mockSuggester := &MockKeywordSuggester{}
	mockValidator := &SimpleKeywordValidator{}
	mockTester := &MockKeywordTester{}

	svc := New(mockSuggester, mockValidator, mockTester)

	niche, negative, err := svc.Suggest(ctx, "TestBrand")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(niche) == 0 {
		t.Error("expected niche keywords, got none")
	}

	if len(negative) == 0 {
		t.Error("expected negative keywords, got none")
	}
}

func TestService_Validate(t *testing.T) {
	ctx := context.Background()

	mockSuggester := &MockKeywordSuggester{}
	mockValidator := &SimpleKeywordValidator{}
	mockTester := &MockKeywordTester{}

	svc := New(mockSuggester, mockValidator, mockTester)

	keywords := []string{"  TestKeyword  ", "Another"}
	validated, err := svc.Validate(ctx, keywords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(validated) != 2 {
		t.Errorf("expected 2 validated keywords, got %d", len(validated))
	}

	// Check normalization (lowercase, trimmed)
	if validated[0] != "testkeyword" {
		t.Errorf("expected 'testkeyword', got '%s'", validated[0])
	}
}

func TestService_Test(t *testing.T) {
	ctx := context.Background()

	mockSuggester := &MockKeywordSuggester{}
	mockValidator := &SimpleKeywordValidator{}
	mockTester := &MockKeywordTester{}

	svc := New(mockSuggester, mockValidator, mockTester)

	keywords := []string{"test"}
	results, err := svc.Test(ctx, keywords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Error("expected test results, got none")
	}
}
