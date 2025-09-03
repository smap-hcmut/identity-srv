package http

import (
	"github.com/nguyentantai21042004/smap-api/internal/auth"
	pkgErrors "github.com/nguyentantai21042004/smap-api/pkg/errors"
)

var (
	errWrongQuery      = pkgErrors.NewHTTPError(110001, "Wrong query")
	errWrongBody       = pkgErrors.NewHTTPError(110002, "Wrong body")
	errUserNotFound    = pkgErrors.NewHTTPError(110003, "User not found")
	errEmailExisted    = pkgErrors.NewHTTPError(110004, "Email existed")
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
	case auth.ErrUserNotFound:
		return errUserNotFound
	case auth.ErrEmailExisted:
		return errEmailExisted
	case auth.ErrWrongPassword:
		return errWrongPassword
	case auth.ErrWrongOTP:
		return errWrongOTP
	case auth.ErrOTPExpired:
		return errOTPExpired
	case auth.ErrTooManyAttempts:
		return errTooManyAttempts
	case auth.ErrUserNotVerified:
		return errUserNotVerified
	case auth.ErrInvalidProvider:
		return errInvalidProvider
	case auth.ErrInvalidEmail:
		return errInvalidEmail
	case auth.ErrUserVerified:
		return errUserVerified
	default:
		return err
	}
}

var NotFound = []error{
	errUserNotFound,
}
