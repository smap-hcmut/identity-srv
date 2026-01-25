package http

import (
	"smap-api/internal/model"
	"smap-api/internal/subscription"
	"smap-api/pkg/paginator"
	"smap-api/pkg/response"
	"time"
)

// Create Subscription API Request
type createSubscriptionReq struct {
	UserID      string  `json:"user_id" binding:"required"`
	PlanID      string  `json:"plan_id" binding:"required"`
	Status      string  `json:"status" binding:"required"`
	TrialEndsAt *string `json:"trial_ends_at"`
	StartsAt    string  `json:"starts_at" binding:"required"`
	EndsAt      *string `json:"ends_at"`
}

func (r createSubscriptionReq) validate() error {
	if r.UserID == "" || r.PlanID == "" || r.Status == "" || r.StartsAt == "" {
		return errFieldRequired
	}
	status := model.SubscriptionStatus(r.Status)
	if !status.IsValid() {
		return errInvalidStatus
	}
	return nil
}

func (r createSubscriptionReq) toInput() (subscription.CreateInput, error) {
	startsAt, err := time.Parse(time.RFC3339, r.StartsAt)
	if err != nil {
		return subscription.CreateInput{}, errInvalidSubscription
	}

	input := subscription.CreateInput{
		UserID:   r.UserID,
		PlanID:   r.PlanID,
		Status:   model.SubscriptionStatus(r.Status),
		StartsAt: startsAt,
	}

	if r.TrialEndsAt != nil {
		trialEndsAt, err := time.Parse(time.RFC3339, *r.TrialEndsAt)
		if err != nil {
			return subscription.CreateInput{}, errInvalidSubscription
		}
		input.TrialEndsAt = &trialEndsAt
	}

	if r.EndsAt != nil {
		endsAt, err := time.Parse(time.RFC3339, *r.EndsAt)
		if err != nil {
			return subscription.CreateInput{}, errInvalidSubscription
		}
		input.EndsAt = &endsAt
	}

	return input, nil
}

// Update Subscription API Request
type updateSubscriptionReq struct {
	Status      *string `json:"status"`
	TrialEndsAt *string `json:"trial_ends_at"`
	EndsAt      *string `json:"ends_at"`
	CancelledAt *string `json:"cancelled_at"`
}

func (r updateSubscriptionReq) validate() error {
	if r.Status != nil {
		status := model.SubscriptionStatus(*r.Status)
		if !status.IsValid() {
			return errInvalidStatus
		}
	}
	return nil
}

func (r updateSubscriptionReq) toInput(id string) (subscription.UpdateInput, error) {
	input := subscription.UpdateInput{
		ID: id,
	}

	if r.Status != nil {
		status := model.SubscriptionStatus(*r.Status)
		input.Status = &status
	}

	if r.TrialEndsAt != nil {
		trialEndsAt, err := time.Parse(time.RFC3339, *r.TrialEndsAt)
		if err != nil {
			return subscription.UpdateInput{}, errInvalidSubscription
		}
		input.TrialEndsAt = &trialEndsAt
	}

	if r.EndsAt != nil {
		endsAt, err := time.Parse(time.RFC3339, *r.EndsAt)
		if err != nil {
			return subscription.UpdateInput{}, errInvalidSubscription
		}
		input.EndsAt = &endsAt
	}

	if r.CancelledAt != nil {
		cancelledAt, err := time.Parse(time.RFC3339, *r.CancelledAt)
		if err != nil {
			return subscription.UpdateInput{}, errInvalidSubscription
		}
		input.CancelledAt = &cancelledAt
	}

	return input, nil
}

// List Subscription Query
type listSubscriptionQuery struct {
	IDs      []string `form:"ids"`
	UserIDs  []string `form:"user_ids"`
	PlanIDs  []string `form:"plan_ids"`
	Statuses []string `form:"statuses"`
}

func (q listSubscriptionQuery) toInput() subscription.ListInput {
	statuses := make([]model.SubscriptionStatus, 0, len(q.Statuses))
	for _, s := range q.Statuses {
		statuses = append(statuses, model.SubscriptionStatus(s))
	}

	return subscription.ListInput{
		Filter: subscription.Filter{
			IDs:      q.IDs,
			UserIDs:  q.UserIDs,
			PlanIDs:  q.PlanIDs,
			Statuses: statuses,
		},
	}
}

// Get Subscription Query (paginated)
type getSubscriptionQuery struct {
	IDs      []string `form:"ids"`
	UserIDs  []string `form:"user_ids"`
	PlanIDs  []string `form:"plan_ids"`
	Statuses []string `form:"statuses"`
	Page     int64    `form:"page"`
	Limit    int64    `form:"limit"`
}

func (q getSubscriptionQuery) toInput() subscription.GetInput {
	statuses := make([]model.SubscriptionStatus, 0, len(q.Statuses))
	for _, s := range q.Statuses {
		statuses = append(statuses, model.SubscriptionStatus(s))
	}

	return subscription.GetInput{
		Filter: subscription.Filter{
			IDs:      q.IDs,
			UserIDs:  q.UserIDs,
			PlanIDs:  q.PlanIDs,
			Statuses: statuses,
		},
		PaginateQuery: paginator.PaginateQuery{
			Page:  int(q.Page),
			Limit: q.Limit,
		},
	}
}

// Response objects
type subscriptionResp struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	PlanID      string             `json:"plan_id"`
	Status      string             `json:"status"`
	TrialEndsAt *response.DateTime `json:"trial_ends_at,omitempty"`
	StartsAt    response.DateTime  `json:"starts_at"`
	EndsAt      *response.DateTime `json:"ends_at,omitempty"`
	CancelledAt *response.DateTime `json:"cancelled_at,omitempty"`
	CreatedAt   response.DateTime  `json:"created_at"`
	UpdatedAt   response.DateTime  `json:"updated_at"`
}

func (h handler) newSubscriptionResp(s model.Subscription) *subscriptionResp {
	resp := &subscriptionResp{
		ID:        s.ID,
		UserID:    s.UserID,
		PlanID:    s.PlanID,
		Status:    s.Status.String(),
		StartsAt:  response.DateTime(s.StartsAt),
		CreatedAt: response.DateTime(s.CreatedAt),
		UpdatedAt: response.DateTime(s.UpdatedAt),
	}

	if s.TrialEndsAt != nil {
		t := response.DateTime(*s.TrialEndsAt)
		resp.TrialEndsAt = &t
	}

	if s.EndsAt != nil {
		t := response.DateTime(*s.EndsAt)
		resp.EndsAt = &t
	}

	if s.CancelledAt != nil {
		t := response.DateTime(*s.CancelledAt)
		resp.CancelledAt = &t
	}

	return resp
}

func (h handler) newSubscriptionListResp(subs []model.Subscription) []subscriptionResp {
	resp := make([]subscriptionResp, len(subs))
	for i, s := range subs {
		resp[i] = *h.newSubscriptionResp(s)
	}
	return resp
}

type subscriptionListResp struct {
	Subscriptions []subscriptionResp `json:"subscriptions"`
}

type subscriptionPageResp struct {
	Subscriptions []subscriptionResp  `json:"subscriptions"`
	Paginator     paginator.Paginator `json:"paginator"`
}
