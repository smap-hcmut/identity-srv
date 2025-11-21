package http

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidID          = errors.New("invalid project ID")
	ErrInvalidPagination  = errors.New("invalid pagination parameters")
	ErrInvalidDateRange   = errors.New("to_date must be after from_date")
	ErrInvalidStatus      = errors.New("invalid project status")
	ErrMissingRequired    = errors.New("missing required fields")
)
