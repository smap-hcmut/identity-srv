package http

import (
	"github.com/nguyentantai21042004/smap-api/internal/upload"
	pkgErrors "github.com/nguyentantai21042004/smap-api/pkg/errors"
)

var (
	errWrongQuery    = pkgErrors.NewHTTPError(10501, "Wrong query")
	errNotFound      = pkgErrors.NewHTTPError(10503, "Upload not found")
	errFieldRequired = pkgErrors.NewHTTPError(10504, "Field required")
	errInvalidFile   = pkgErrors.NewHTTPError(10505, "Invalid file")
	errFileTooLarge  = pkgErrors.NewHTTPError(10506, "File too large")
	errInvalidBucket = pkgErrors.NewHTTPError(10507, "Invalid bucket")
)

func (h handler) mapErrorCode(err error) error {
	switch err {
	case upload.ErrUploadNotFound:
		return errNotFound
	case upload.ErrFieldRequired:
		return errFieldRequired
	case upload.ErrInvalidFile:
		return errInvalidFile
	case upload.ErrFileTooLarge:
		return errFileTooLarge
	case upload.ErrInvalidBucket:
		return errInvalidBucket
	default:
		return err
	}
}

var NotFound = []error{
	errNotFound,
}
