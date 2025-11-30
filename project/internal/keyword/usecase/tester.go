package usecase

import (
	"context"
)

func (t *usecase) test(ctx context.Context, keywords []string) ([]interface{}, error) {
	validatedKeywords, err := t.validate(ctx, keywords)
	if err != nil {
		return nil, err
	}

	posts, err := t.collectorClient.DryRun(ctx, validatedKeywords, 10)
	if err != nil {
		return nil, err
	}

	// Convert posts to []interface{}
	results := make([]interface{}, len(posts))
	for i, p := range posts {
		results[i] = p
	}

	return results, nil
}

