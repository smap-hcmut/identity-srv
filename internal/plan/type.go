package plan

import (
	"smap-api/internal/model"
	"smap-api/pkg/paginator"
)

type CreateInput struct {
	Name        string
	Code        string
	Description *string
	MaxUsage    int
}

type UpdateInput struct {
	ID          string
	Name        *string
	Description *string
	MaxUsage    *int
}

type PlanOutput struct {
	Plan model.Plan
}

type GetPlanOutput struct {
	Plans     []model.Plan
	Paginator paginator.Paginator
}

type GetOneInput struct {
	ID   string
	Code string
}

type ListInput struct {
	Filter Filter
}

type GetInput struct {
	Filter        Filter
	PaginateQuery paginator.PaginateQuery
}

type Filter struct {
	IDs   []string
	Codes []string
}
