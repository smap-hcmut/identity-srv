package repository

import "errors"

var (
	ErrNotFound      = errors.New("plan not found")
	ErrAlreadyExists = errors.New("plan already exists")
)
