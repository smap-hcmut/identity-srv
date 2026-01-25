package repository

import "errors"

var (
	ErrNotFound      = errors.New("subscription not found")
	ErrAlreadyExists = errors.New("subscription already exists")
)

