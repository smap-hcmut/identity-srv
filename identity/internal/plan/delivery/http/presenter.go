package http

import (
	"smap-api/internal/model"
	"smap-api/internal/plan"
	"smap-api/pkg/paginator"
)

// Create Plan API Request
type createPlanReq struct {
	Name        string  `json:"name" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Description *string `json:"description"`
	MaxUsage    int     `json:"max_usage" binding:"required,min=0"`
}

func (r createPlanReq) validate() error {
	if r.Name == "" || r.Code == "" {
		return errFieldRequired
	}
	if r.MaxUsage < 0 {
		return errInvalidPlan
	}
	return nil
}

func (r createPlanReq) toInput() plan.CreateInput {
	return plan.CreateInput{
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		MaxUsage:    r.MaxUsage,
	}
}

// Update Plan API Request
type updatePlanReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	MaxUsage    *int    `json:"max_usage" binding:"omitempty,min=0"`
}

func (r updatePlanReq) validate() error {
	if r.MaxUsage != nil && *r.MaxUsage < 0 {
		return errInvalidPlan
	}
	return nil
}

func (r updatePlanReq) toInput(id string) plan.UpdateInput {
	return plan.UpdateInput{
		ID:          id,
		Name:        r.Name,
		Description: r.Description,
		MaxUsage:    r.MaxUsage,
	}
}

// List Plan Query
type listPlanQuery struct {
	IDs   []string `form:"ids"`
	Codes []string `form:"codes"`
}

func (q listPlanQuery) toInput() plan.ListInput {
	return plan.ListInput{
		Filter: plan.Filter{
			IDs:   q.IDs,
			Codes: q.Codes,
		},
	}
}

// Get Plan Query (paginated)
type getPlanQuery struct {
	IDs   []string `form:"ids"`
	Codes []string `form:"codes"`
	Page  int64    `form:"page"`
	Limit int64    `form:"limit"`
}

func (q getPlanQuery) toInput() plan.GetInput {
	return plan.GetInput{
		Filter: plan.Filter{
			IDs:   q.IDs,
			Codes: q.Codes,
		},
		PaginateQuery: paginator.PaginateQuery{
			Page:  int(q.Page),
			Limit: q.Limit,
		},
	}
}

// Response objects
type planResp struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description *string `json:"description,omitempty"`
	MaxUsage    int     `json:"max_usage"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (h handler) newPlanResp(p model.Plan) *planResp {
	return &planResp{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		MaxUsage:    p.MaxUsage,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h handler) newPlanListResp(plans []model.Plan) []planResp {
	resp := make([]planResp, len(plans))
	for i, p := range plans {
		resp[i] = *h.newPlanResp(p)
	}
	return resp
}

type planListResp struct {
	Plans []planResp `json:"plans"`
}

type planPageResp struct {
	Plans     []planResp          `json:"plans"`
	Paginator paginator.Paginator `json:"paginator"`
}
