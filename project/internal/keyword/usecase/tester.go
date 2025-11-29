package usecase

import (
	"context"
)

// test returns mock posts for keyword testing
func (t *usecase) test(ctx context.Context, keywords []string) ([]interface{}, error) {
	// Mock logic: return some dummy posts
	posts := []interface{}{
		map[string]interface{}{
			"content": "Just bought a new " + keywords[0] + "! It's amazing.",
			"source":  "facebook",
			"date":    "2023-10-27T10:00:00Z",
		},
		map[string]interface{}{
			"content": "Looking for " + keywords[0] + " accessories.",
			"source":  "twitter",
			"date":    "2023-10-27T11:30:00Z",
		},
		map[string]interface{}{
			"content": "Is " + keywords[0] + " worth the price?",
			"source":  "reddit",
			"date":    "2023-10-27T12:15:00Z",
		},
	}
	return posts, nil
}
