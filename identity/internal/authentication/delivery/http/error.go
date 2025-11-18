package http

import (
	"smap-api/internal/authentication"
	pkgErrors "smap-api/pkg/errors"
)

var (
	errWrongBody       = pkgErrors.NewHTTPError(110002, "Wrong body")
	errUserNotFound    = pkgErrors.NewHTTPError(110003, "User not found")
	errUsernameExisted = pkgErrors.NewHTTPError(110004, "Username existed")
	errWrongPassword   = pkgErrors.NewHTTPError(110005, "Wrong password")
	errWrongOTP        = pkgErrors.NewHTTPError(110006, "Wrong OTP")
	errUserVerified    = pkgErrors.NewHTTPError(110007, "User verified")
	errOTPExpired      = pkgErrors.NewHTTPError(110008, "OTP expired")
	errTooManyAttempts = pkgErrors.NewHTTPError(110009, "Too many attempts")
	errUserNotVerified = pkgErrors.NewHTTPError(110010, "User not verified")
	errInvalidProvider = pkgErrors.NewHTTPError(110011, "Invalid provider")
	errInvalidEmail    = pkgErrors.NewHTTPError(110012, "Invalid email")
)

func (h handler) mapErrorCode(err error) error {
	switch err {
	case errWrongBody:
		return errWrongBody
	case authentication.ErrUserNotFound:
		return errUserNotFound
	case authentication.ErrUsernameExisted:
		return errUsernameExisted
	case authentication.ErrWrongPassword:
		return errWrongPassword
	case authentication.ErrWrongOTP:
		return errWrongOTP
	case authentication.ErrOTPExpired:
		return errOTPExpired
	case authentication.ErrTooManyAttempts:
		return errTooManyAttempts
	case authentication.ErrUserNotVerified:
		return errUserNotVerified
	case authentication.ErrInvalidProvider:
		return errInvalidProvider
	case authentication.ErrInvalidEmail:
		return errInvalidEmail
	case authentication.ErrUserVerified:
		return errUserVerified
	default:
		return err
	}
}

var NotFound = []error{
	errUserNotFound,
}
