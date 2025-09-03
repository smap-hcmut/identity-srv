package user

import "errors"

var (
	ErrUserExisted      = errors.New("user existed")
	ErrUserNotFound     = errors.New("user not found")
	ErrRoleNotFound     = errors.New("role not found")
	ErrPermissionDenied = errors.New("permission denied")
)
