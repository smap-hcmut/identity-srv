package keyword

import "context"

// Service defines the interface for keyword operations
type Service interface {
	Suggest(ctx context.Context, brandName string) (niche []string, negative []string, err error)
	Validate(ctx context.Context, keywords []string) ([]string, error)
	Test(ctx context.Context, keywords []string) ([]interface{}, error)
}

// Suggester defines the interface for suggesting keywords
type Suggester interface {
	Suggest(ctx context.Context, brandName string) ([]string, []string, error)
}

// Validator defines the interface for validating keywords
type Validator interface {
	Validate(ctx context.Context, keywords []string) ([]string, error)
	ValidateOne(ctx context.Context, keyword string) (string, error)
}

// Tester defines the interface for testing keywords (dry run)
type Tester interface {
	Test(ctx context.Context, keywords []string) ([]interface{}, error)
}
