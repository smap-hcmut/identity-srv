package results

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrDuplicate    = errors.New("duplicate")
	ErrTemporary    = errors.New("temporary")
	ErrPermission   = errors.New("permission denied")
)
