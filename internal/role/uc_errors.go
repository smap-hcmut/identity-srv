package role

import "errors"

var (
	ErrUserExisted   = errors.New("user existed")
	ErrRoleNotFound  = errors.New("role not found")
	ErrRequiredField = errors.New("required field")
)
