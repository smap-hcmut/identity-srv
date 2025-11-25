package middleware

import (
	pkgErrors "smap-collector/pkg/errors"
)

var (
	errInvalidToken = pkgErrors.NewHTTPError(401, "invalid token")
	errPermission   = pkgErrors.NewPermissionError(60000, "Don't have permission")
	errPayment      = pkgErrors.NewPaymentError(60004, "Have to buy package")
)
