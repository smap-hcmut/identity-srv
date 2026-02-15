package usecase

import (
	"fmt"
	"identity-srv/internal/authentication"
	"net/url"
	"strings"
)

// ValidateRedirectURL validates if a redirect URL is in the allowed list
// Returns error if URL is not allowed (prevents open redirect attacks)
func (rv *RedirectValidator) ValidateRedirectURL(redirectURL string) error {
	if redirectURL == "" {
		return nil // Empty redirect is allowed (will use default)
	}

	// Parse the redirect URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return fmt.Errorf("%w: %v", authentication.ErrInvalidRedirectURL, err)
	}

	// Check if URL is relative (always allowed)
	if parsedURL.Scheme == "" && parsedURL.Host == "" {
		return nil
	}

	// For absolute URLs, check against allowed list
	redirectBase := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	for _, allowed := range rv.allowedURLs {
		// Parse allowed URL
		allowedURL, err := url.Parse(allowed)
		if err != nil {
			continue
		}

		allowedBase := fmt.Sprintf("%s://%s", allowedURL.Scheme, allowedURL.Host)

		// Exact match
		if redirectBase == allowedBase {
			return nil
		}

		// Wildcard subdomain match (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(parsedURL.Host, domain) {
				return nil
			}
		}
	}

	return fmt.Errorf("%w: %s", authentication.ErrRedirectURLNotAllowed, redirectURL)
}
