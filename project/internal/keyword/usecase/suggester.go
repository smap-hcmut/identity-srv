package usecase

import (
	"context"
)

func (uc *usecase) suggestProcessing(ctx context.Context, brandName string) ([]string, []string, error) {
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
