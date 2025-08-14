package upload

import "errors"

var (
	ErrUploadNotFound = errors.New("upload not found")
	ErrFieldRequired  = errors.New("field required")
	ErrInvalidFile    = errors.New("invalid file")
	ErrFileTooLarge   = errors.New("file too large")
	ErrInvalidBucket  = errors.New("invalid bucket")
)
