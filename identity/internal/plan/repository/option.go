package repository

import (
	"smap-api/internal/model"
	"smap-api/pkg/paginator"
)

// Filter contains filtering options for plan queries.
type Filter struct {
	IDs   []string
	Codes []string
}

// CreateOptions contains options for creating a plan.
type CreateOptions struct {
	Plan model.Plan
}

// UpdateOptions contains options for updating a plan.
// Only non-nil fields will be updated.
type UpdateOptions struct {
	Plan model.Plan
}

// GetOneOptions contains options for getting a single plan.
type GetOneOptions struct {
	Code string
	ID   string
}

// ListOptions contains options for listing plans.
type ListOptions struct {
	Filter Filter
}

// GetOptions contains options for paginated plan listing.
type GetOptions struct {
	Filter        Filter
	PaginateQuery paginator.PaginateQuery
}
