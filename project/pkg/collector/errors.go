package collector

import "errors"

var (
	// ErrCollectorUnavailable is returned when the collector service is unavailable.
	ErrCollectorUnavailable = errors.New("collector service unavailable")
	// ErrCollectorTimeout is returned when the collector request times out.
	ErrCollectorTimeout = errors.New("collector request timeout")
	// ErrCollectorInvalidResponse is returned when the collector returns an invalid response.
	ErrCollectorInvalidResponse = errors.New("collector returned invalid response")
)
