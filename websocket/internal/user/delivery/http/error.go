package http

import (
	"smap-api/internal/user"
	"smap-api/pkg/errors"
)

const (
	ErrCodeWrongBody     = 140001
	ErrCodeUserNotFound  = 140002
	ErrCodeUserExists    = 140003
	ErrCodeFieldRequired = 140004
	ErrCodeInvalidID     = 140005
	ErrCodeWrongPassword = 140006
	ErrCodeWeakPassword  = 140007
	ErrCodeSamePassword  = 140008
	ErrCodeInvalidRole   = 140009
	ErrCodeUnauthorized  = 140010
)

func toHTTPError(err error) *errors.HTTPError {
	switch err {
	case user.ErrUserNotFound:
		return &errors.HTTPError{
			Code:    ErrCodeUserNotFound,
			Message: "User not found",
		}
	case user.ErrUserExists:
		return &errors.HTTPError{
			Code:    ErrCodeUserExists,
			Message: "User already exists",
		}
	case user.ErrFieldRequired:
		return &errors.HTTPError{
			Code:    ErrCodeFieldRequired,
			Message: "Field required",
		}
	case user.ErrWrongPassword:
		return &errors.HTTPError{
			Code:    ErrCodeWrongPassword,
			Message: "Wrong password",
		}
	case user.ErrWeakPassword:
		return &errors.HTTPError{
			Code:    ErrCodeWeakPassword,
			Message: "Password must be at least 8 characters",
		}
	case user.ErrSamePassword:
		return &errors.HTTPError{
			Code:    ErrCodeSamePassword,
			Message: "New password must be different from old password",
		}
	case user.ErrInvalidRole:
		return &errors.HTTPError{
			Code:    ErrCodeInvalidRole,
			Message: "Invalid role",
		}
	case user.ErrUnauthorized:
		return &errors.HTTPError{
			Code:    ErrCodeUnauthorized,
			Message: "Unauthorized",
		}
	default:
		return nil
	}
}
