package http

import (
	"github.com/nguyentantai21042004/smap-api/internal/role"
	pkgErrors "github.com/nguyentantai21042004/smap-api/pkg/errors"
)

var (
	errWrongPaginationQuery = pkgErrors.NewHTTPError(140001, "Wrong pagination query")
	errWrongQuery           = pkgErrors.NewHTTPError(140002, "Wrong query")
	errWrongBody            = pkgErrors.NewHTTPError(140003, "Wrong body")

	// Role errors
	errRoleNotFound = pkgErrors.NewHTTPError(143001, "Role not found")
	errRequiredField  = pkgErrors.NewHTTPError(143002, "Required field")
)

func (h handler) mapError(err error) error {
	switch err {
	// Role errors
	case role.ErrRoleNotFound:
		return errRoleNotFound
	case role.ErrRequiredField:
		return errRequiredField

	default:
		panic(err)
	}
}