package keyword

import "context"

//go:generate mockery --name UseCase
type UseCase interface {
	// Suggest returns niche and negative keyword suggestions based on brand name
	Suggest(ctx context.Context, brandName string) (SuggestOutput, error)

	// Validate validates and normalizes keywords
	Validate(ctx context.Context, input ValidateInput) (ValidateOutput, error)

	// Test performs a dry run test of keywords
	Test(ctx context.Context, input TestInput) (TestOutput, error)
}
