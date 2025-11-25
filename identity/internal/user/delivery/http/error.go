package http

import (
	"smap-api/internal/user"
	pkgErrors "smap-api/pkg/errors"
)

var (
	errWrongQuery    = pkgErrors.NewHTTPError(10000, "Wrong query")
	errWrongBody     = pkgErrors.NewHTTPError(10001, "Wrong body")
	errUserNotFound  = pkgErrors.NewHTTPError(10002, "User not found")
	errUserExists    = pkgErrors.NewHTTPError(10003, "User already exists")
	errFieldRequired = pkgErrors.NewHTTPError(10004, "Field required")
	errWrongPassword = pkgErrors.NewHTTPError(10006, "Wrong password")
	errWeakPassword  = pkgErrors.NewHTTPError(10007, "Password must be at least 8 characters")
	errSamePassword  = pkgErrors.NewHTTPError(10008, "New password must be different from old password")
	errInvalidRole   = pkgErrors.NewHTTPError(10009, "Invalid role")
	errUnauthorized  = pkgErrors.NewHTTPError(10010, "Unauthorized")
)

func (h handler) mapErrorCode(err error) error {
	switch err {
	case user.ErrUserNotFound:
		return errUserNotFound
	case user.ErrUserExists:
		return errUserExists
	case user.ErrFieldRequired:
		return errFieldRequired
	case user.ErrWrongPassword:
		return errWrongPassword
	case user.ErrWeakPassword:
		return errWeakPassword
	case user.ErrSamePassword:
		return errSamePassword
	case user.ErrInvalidRole:
		return errInvalidRole
	case user.ErrUnauthorized:
		return errUnauthorized
	default:
		return err
	}
}

var NotFound = []error{
	errUserNotFound,
}
