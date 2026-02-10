package http

import (
	"smap-api/internal/authentication"
	pkgErrors "smap-api/pkg/errors"
)

// --- HTTP error constants ---

var (
	errWrongBody            = pkgErrors.NewHTTPError(20001, "Wrong body")
	errUserNotFound         = pkgErrors.NewHTTPError(20002, "User not found")
	errUsernameExisted      = pkgErrors.NewHTTPError(20003, "Username existed")
	errWrongPassword        = pkgErrors.NewHTTPError(20004, "Wrong password")
	errWrongOTP             = pkgErrors.NewHTTPError(20005, "Wrong OTP")
	errUserVerified         = pkgErrors.NewHTTPError(20006, "User verified")
	errOTPExpired           = pkgErrors.NewHTTPError(20007, "OTP expired")
	errTooManyAttempts      = pkgErrors.NewHTTPError(20008, "Too many attempts")
	errUserNotVerified      = pkgErrors.NewHTTPError(20009, "User not verified")
	errInvalidProvider      = pkgErrors.NewHTTPError(20010, "Invalid provider")
	errInvalidEmail         = pkgErrors.NewHTTPError(20011, "Invalid email")
	errDomainNotAllowed     = pkgErrors.NewHTTPError(20012, "Domain not allowed")
	errAccountBlocked       = pkgErrors.NewHTTPError(20013, "Account blocked")
	errScopeNotFound        = pkgErrors.NewHTTPError(20014, "Scope not found")
	errMissingCode          = pkgErrors.NewHTTPError(20015, "Missing authorization code")
	errInvalidState         = pkgErrors.NewHTTPError(20016, "Invalid state parameter")
	errMissingJTIOrUserID   = pkgErrors.NewHTTPError(20017, "Must provide either jti or user_id")
	errConflictJTIAndUserID = pkgErrors.NewHTTPError(20018, "Cannot provide both jti and user_id")
	errMissingUserID        = pkgErrors.NewHTTPError(20019, "User ID is required")
)

// mapError maps UseCase domain errors to HTTP errors
func (h handler) mapError(err error) error {
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
	case authentication.ErrDomainNotAllowed:
		return errDomainNotAllowed
	case authentication.ErrAccountBlocked:
		return errAccountBlocked
	case authentication.ErrScopeNotFound:
		return errScopeNotFound
	default:
		return err
	}
}

var NotFound = []error{
	errUserNotFound,
}
