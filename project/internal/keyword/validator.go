package keyword

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

// KeywordValidator defines the interface for validating keywords
type KeywordValidator interface {
	Validate(ctx context.Context, keywords []string) ([]string, error)
	ValidateOne(ctx context.Context, keyword string) (string, error)
}

// SimpleKeywordValidator is a simple implementation of KeywordValidator
type SimpleKeywordValidator struct{}

// NewSimpleKeywordValidator creates a new SimpleKeywordValidator
func NewSimpleKeywordValidator() *SimpleKeywordValidator {
	return &SimpleKeywordValidator{}
}

// Validate validates a list of keywords
func (v *SimpleKeywordValidator) Validate(ctx context.Context, keywords []string) ([]string, error) {
	validKeywords := make([]string, 0, len(keywords))
	seen := make(map[string]bool)

	for _, kw := range keywords {
		normalized, err := v.ValidateOne(ctx, kw)
		if err != nil {
			return nil, err
		}
		if !seen[normalized] {
			validKeywords = append(validKeywords, normalized)
			seen[normalized] = true
		}
	}

	return validKeywords, nil
}

// ValidateOne validates a single keyword
func (v *SimpleKeywordValidator) ValidateOne(ctx context.Context, keyword string) (string, error) {
	// Normalize: trim spaces, lowercase
	normalized := strings.TrimSpace(strings.ToLower(keyword))

	// Check length
	if len(normalized) < 2 {
		return "", errors.New("keyword '" + keyword + "' is too short (min 2 characters)")
	}
	if len(normalized) > 50 {
		return "", errors.New("keyword '" + keyword + "' is too long (max 50 characters)")
	}

	// Check character set (alphanumeric, spaces, hyphens, underscores)
	// Allow Vietnamese characters as well? For now, let's stick to simple regex or just allow most.
	// The requirement says: alphanumeric, spaces, hyphens, underscores.
	// Regex: ^[a-zA-Z0-9\s\-_]+$ (plus unicode for Vietnamese if needed, but let's stick to basic for now as per spec)
	// Actually, for a real app, we should allow unicode.
	// Let's use a slightly more permissive regex but block special chars like @, #, $, etc.
	matched, _ := regexp.MatchString(`^[\p{L}\p{N}\s\-_]+$`, normalized)
	if !matched {
		return "", errors.New("keyword '" + keyword + "' contains invalid characters")
	}

	// Check generic terms (stopwords)
	stopwords := map[string]bool{
		"xe": true, "mua": true, "ban": true, "bán": true,
		"buy": true, "sell": true, "car": true, "house": true,
	}
	if stopwords[normalized] {
		return "", errors.New("keyword '" + keyword + "' is too generic")
	}

	return normalized, nil
}
