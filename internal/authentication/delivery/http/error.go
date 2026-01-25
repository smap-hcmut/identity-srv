package http

import (
	"smap-api/internal/authentication"
	pkgErrors "smap-api/pkg/errors"
)

var (
	errWrongBody       = pkgErrors.NewHTTPError(20001, "Wrong body")
	errUserNotFound    = pkgErrors.NewHTTPError(20002, "User not found")
	errUsernameExisted = pkgErrors.NewHTTPError(20003, "Username existed")
	errWrongPassword   = pkgErrors.NewHTTPError(20004, "Wrong password")
	errWrongOTP        = pkgErrors.NewHTTPError(20005, "Wrong OTP")
	errUserVerified    = pkgErrors.NewHTTPError(20006, "User verified")
	errOTPExpired      = pkgErrors.NewHTTPError(20007, "OTP expired")
	errTooManyAttempts = pkgErrors.NewHTTPError(20008, "Too many attempts")
	errUserNotVerified = pkgErrors.NewHTTPError(20009, "User not verified")
	errInvalidProvider = pkgErrors.NewHTTPError(20010, "Invalid provider")
	errInvalidEmail    = pkgErrors.NewHTTPError(20011, "Invalid email")
)

func (h handler) mapErrorCode(err error) error {
	switch err {
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
