package http

import (
	"errors"
	"identity-srv/internal/authentication"
	pkgErrors "identity-srv/pkg/errors"
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
	errConfigurationMissing = pkgErrors.NewHTTPError(20020, "Server configuration missing")
	errInvalidRedirectURL   = pkgErrors.NewHTTPError(20021, "Invalid redirect URL")
	errInternalSystem       = pkgErrors.NewHTTPError(20022, "Internal system error")
	errUserCreation         = pkgErrors.NewHTTPError(20023, "Failed to create or update user")
)

// mapError maps UseCase domain errors to HTTP errors
func (h handler) mapError(err error) error {
	switch {
	case errors.Is(err, authentication.ErrUserNotFound):
		return errUserNotFound
	case errors.Is(err, authentication.ErrUsernameExisted):
		return errUsernameExisted
	case errors.Is(err, authentication.ErrWrongPassword):
		return errWrongPassword
	case errors.Is(err, authentication.ErrWrongOTP):
		return errWrongOTP
	case errors.Is(err, authentication.ErrOTPExpired):
		return errOTPExpired
	case errors.Is(err, authentication.ErrTooManyAttempts):
		return errTooManyAttempts
	case errors.Is(err, authentication.ErrUserNotVerified):
		return errUserNotVerified
	case errors.Is(err, authentication.ErrInvalidProvider):
		return errInvalidProvider
	case errors.Is(err, authentication.ErrInvalidEmail):
		return errInvalidEmail
	case errors.Is(err, authentication.ErrUserVerified):
		return errUserVerified
	case errors.Is(err, authentication.ErrDomainNotAllowed):
		return errDomainNotAllowed
	case errors.Is(err, authentication.ErrAccountBlocked):
		return errAccountBlocked
	case errors.Is(err, authentication.ErrScopeNotFound):
		return errScopeNotFound
	case errors.Is(err, authentication.ErrConfigurationMissing):
		return errConfigurationMissing
	case errors.Is(err, authentication.ErrInvalidRedirectURL):
		return errInvalidRedirectURL
	case errors.Is(err, authentication.ErrRedirectURLNotAllowed):
		return errInvalidRedirectURL
	case errors.Is(err, authentication.ErrInternalSystem):
		return errInternalSystem
	case errors.Is(err, authentication.ErrUserCreation):
		return errUserCreation
	default:
		return err
	}
}

var NotFound = []error{
	errUserNotFound,
}
