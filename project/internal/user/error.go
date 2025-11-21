package user

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user already exists")
	ErrInvalidRole   = errors.New("invalid role")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrFieldRequired = errors.New("field required")
	ErrWrongPassword = errors.New("wrong password")
	ErrWeakPassword  = errors.New("password must be at least 8 characters")
	ErrSamePassword  = errors.New("new password must be different from old password")
)
