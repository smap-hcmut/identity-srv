package session

import "errors"

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrUserNotVerified = errors.New("user not verified")
)
