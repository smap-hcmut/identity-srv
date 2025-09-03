package http

import (
	"github.com/nguyentantai21042004/smap-api/internal/user"
	pkgErrors "github.com/nguyentantai21042004/smap-api/pkg/errors"
)

var (
	errWrongQuery       = pkgErrors.NewHTTPError(120001, "Wrong query")
	errWrongBody        = pkgErrors.NewHTTPError(120002, "Wrong body")
	errPermissionDenied = pkgErrors.NewHTTPError(120007, "Permission denied")
	errUserExisted      = pkgErrors.NewHTTPError(120008, "User existed")
	errUserNotFound     = pkgErrors.NewHTTPError(120009, "User not found")
	errRoleNotFound     = pkgErrors.NewHTTPError(120010, "Role not found")
	errInvalidAvatarURL = pkgErrors.NewHTTPError(120011, "Invalid avatar URL")
)

func (h handler) mapErrorCode(err error) error {
	switch err {
	case errWrongBody:
		return errWrongBody
	case user.ErrPermissionDenied:
		return errPermissionDenied
	case user.ErrUserExisted:
		return errUserExisted
	case user.ErrUserNotFound:
		return errUserNotFound
	case user.ErrRoleNotFound:
		return errRoleNotFound
	default:
		return err
	}
}

var NotFound = []error{
	errUserNotFound,
	errRoleNotFound,
}
