package repository

import (
	"smap-api/internal/model"
	"smap-api/pkg/paginator"
)

// Filter contains filtering options for subscription queries.
type Filter struct {
	IDs      []string
	UserIDs  []string
	PlanIDs  []string
	Statuses []model.SubscriptionStatus
}

// CreateOptions contains options for creating a subscription.
type CreateOptions struct {
	Subscription model.Subscription
}

// UpdateOptions contains options for updating a subscription.
// Only non-nil fields will be updated.
type UpdateOptions struct {
	Subscription model.Subscription
}

// GetOneOptions contains options for getting a single subscription.
type GetOneOptions struct {
	ID     string
	UserID string
	PlanID string
	Status *model.SubscriptionStatus
}

// ListOptions contains options for listing subscriptions.
type ListOptions struct {
	Filter Filter
}

// GetOptions contains options for paginated subscription listing.
type GetOptions struct {
	Filter        Filter
	PaginateQuery paginator.PaginateQuery
}

