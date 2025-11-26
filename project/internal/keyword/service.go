package keyword

import "context"

type service struct {
	suggester Suggester
	validator Validator
	tester    Tester
}

// New creates a new keyword service
func New(suggester Suggester, validator Validator, tester Tester) Service {
	return &service{
		suggester: suggester,
		validator: validator,
		tester:    tester,
	}
}

// Suggest returns niche and negative keyword suggestions
func (s *service) Suggest(ctx context.Context, brandName string) ([]string, []string, error) {
	return s.suggester.Suggest(ctx, brandName)
}

// Validate validates and normalizes keywords
func (s *service) Validate(ctx context.Context, keywords []string) ([]string, error) {
	return s.validator.Validate(ctx, keywords)
}

// Test performs a dry run test of keywords
func (s *service) Test(ctx context.Context, keywords []string) ([]interface{}, error) {
	return s.tester.Test(ctx, keywords)
}
