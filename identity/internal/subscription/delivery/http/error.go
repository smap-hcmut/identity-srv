package http

import (
	"smap-api/internal/subscription"
	pkgErrors "smap-api/pkg/errors"
)

var (
	errWrongBody                = pkgErrors.NewHTTPError(130001, "Wrong body")
	errSubscriptionNotFound     = pkgErrors.NewHTTPError(130002, "Subscription not found")
	errSubscriptionExists       = pkgErrors.NewHTTPError(130003, "Subscription already exists")
	errInvalidSubscription      = pkgErrors.NewHTTPError(130004, "Invalid subscription")
	errFieldRequired            = pkgErrors.NewHTTPError(130005, "Field required")
	errInvalidID                = pkgErrors.NewHTTPError(130006, "Invalid ID")
	errActiveSubscriptionExists = pkgErrors.NewHTTPError(130007, "Active subscription already exists for this user")
	errInvalidStatus            = pkgErrors.NewHTTPError(130008, "Invalid subscription status")
	errCannotCancel             = pkgErrors.NewHTTPError(130009, "Cannot cancel subscription")
)

func (h handler) mapErrorCode(err error) error {
	switch err {
	case errWrongBody:
		return errWrongBody
	case subscription.ErrSubscriptionNotFound:
		return errSubscriptionNotFound
	case subscription.ErrSubscriptionExists:
		return errSubscriptionExists
	case subscription.ErrInvalidSubscription:
		return errInvalidSubscription
	case subscription.ErrFieldRequired:
		return errFieldRequired
	case errInvalidID:
		return errInvalidID
	case subscription.ErrActiveSubscriptionExists:
		return errActiveSubscriptionExists
	case subscription.ErrInvalidStatus:
		return errInvalidStatus
	case subscription.ErrCannotCancel:
		return errCannotCancel
	default:
		return err
	}
}

var NotFound = []error{
	errSubscriptionNotFound,
}

