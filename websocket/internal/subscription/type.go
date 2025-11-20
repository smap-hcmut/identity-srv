package subscription

import (
	"smap-api/internal/model"
	"smap-api/pkg/paginator"
	"time"
)

type CreateInput struct {
	UserID      string
	PlanID      string
	Status      model.SubscriptionStatus
	TrialEndsAt *time.Time
	StartsAt    time.Time
	EndsAt      *time.Time
}

type UpdateInput struct {
	ID          string
	Status      *model.SubscriptionStatus
	TrialEndsAt *time.Time
	EndsAt      *time.Time
	CancelledAt *time.Time
}

type SubscriptionOutput struct {
	Subscription model.Subscription
}

type GetSubscriptionOutput struct {
	Subscriptions []model.Subscription
	Paginator     paginator.Paginator
}

type GetOneInput struct {
	ID     string
	UserID string
	PlanID string
}

type ListInput struct {
	Filter Filter
}

type GetInput struct {
	Filter        Filter
	PaginateQuery paginator.PaginateQuery
}

type Filter struct {
	IDs      []string
	UserIDs  []string
	PlanIDs  []string
	Statuses []model.SubscriptionStatus
}

