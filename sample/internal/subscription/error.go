package subscription

import "errors"

var (
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrSubscriptionExists       = errors.New("subscription already exists")
	ErrInvalidSubscription      = errors.New("invalid subscription")
	ErrFieldRequired            = errors.New("field required")
	ErrActiveSubscriptionExists = errors.New("active subscription already exists for this user")
	ErrInvalidStatus            = errors.New("invalid subscription status")
	ErrCannotCancel             = errors.New("cannot cancel subscription")
)

