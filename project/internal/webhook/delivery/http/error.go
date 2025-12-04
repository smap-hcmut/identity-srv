package http

import (
	pkgErrors "smap-project/pkg/errors"
)

var (
	errWrongBody        = pkgErrors.NewHTTPError(31001, "Wrong body")
	errUnauthorized     = pkgErrors.NewHTTPError(31002, "Unauthorized")
	errInvalidCallback  = pkgErrors.NewHTTPError(31004, "Invalid callback payload")
	errRedisPublishFail = pkgErrors.NewHTTPError(31005, "Failed to publish to Redis")
)

func (h handler) mapErrorCode(err error) error {
	// Map domain errors to HTTP errors if needed
	return err
}
