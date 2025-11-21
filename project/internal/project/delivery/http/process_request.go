package http

import (
	"time"

	"smap-project/internal/project"
	"smap-project/pkg/paginator"
)

// CreateProjectRequest represents the HTTP request for creating a project
type CreateProjectRequest struct {
	Name                  string              `json:"name" binding:"required"`
	Description           *string             `json:"description"`
	Status                string              `json:"status" binding:"required"`
	FromDate              time.Time           `json:"from_date" binding:"required"`
	ToDate                time.Time           `json:"to_date" binding:"required"`
	BrandName             string              `json:"brand_name" binding:"required"`
	CompetitorNames       []string            `json:"competitor_names"`
	BrandKeywords         []string            `json:"brand_keywords" binding:"required"`
	CompetitorKeywordsMap map[string][]string `json:"competitor_keywords_map"`
}

// UpdateProjectRequest represents the HTTP request for updating a project
type UpdateProjectRequest struct {
	Name                  *string             `json:"name"`
	Description           *string             `json:"description"`
	Status                *string             `json:"status"`
	FromDate              *time.Time          `json:"from_date"`
	ToDate                *time.Time          `json:"to_date"`
	BrandName             *string             `json:"brand_name"`
	CompetitorNames       []string            `json:"competitor_names"`
	BrandKeywords         []string            `json:"brand_keywords"`
	CompetitorKeywordsMap map[string][]string `json:"competitor_keywords_map"`
}

// GetProjectsRequest represents the HTTP request for listing projects with filters
type GetProjectsRequest struct {
	IDs        []string `form:"ids"`
	Statuses   []string `form:"statuses"`
	SearchName *string  `form:"search_name"`
	Page       int      `form:"page"`
	Limit      int      `form:"limit"`
}

// ToCreateInput converts HTTP request to usecase input
func (r CreateProjectRequest) ToCreateInput() project.CreateInput {
	return project.CreateInput{
		Name:                  r.Name,
		Description:           r.Description,
		Status:                r.Status,
		FromDate:              r.FromDate,
		ToDate:                r.ToDate,
		BrandName:             r.BrandName,
		CompetitorNames:       r.CompetitorNames,
		BrandKeywords:         r.BrandKeywords,
		CompetitorKeywordsMap: r.CompetitorKeywordsMap,
	}
}

// ToUpdateInput converts HTTP request to usecase input
func (r UpdateProjectRequest) ToUpdateInput(id string) project.UpdateInput {
	return project.UpdateInput{
		ID:                    id,
		Name:                  r.Name,
		Description:           r.Description,
		Status:                r.Status,
		FromDate:              r.FromDate,
		ToDate:                r.ToDate,
		BrandName:             r.BrandName,
		CompetitorNames:       r.CompetitorNames,
		BrandKeywords:         r.BrandKeywords,
		CompetitorKeywordsMap: r.CompetitorKeywordsMap,
	}
}

// ToListInput converts HTTP request to usecase input
func (r GetProjectsRequest) ToListInput() project.ListInput {
	return project.ListInput{
		Filter: project.Filter{
			IDs:        r.IDs,
			Statuses:   r.Statuses,
			SearchName: r.SearchName,
		},
	}
}

// ToGetInput converts HTTP request to usecase input with pagination
func (r GetProjectsRequest) ToGetInput() project.GetInput {
	pq := paginator.NewPaginateQuery(r.Page, r.Limit)

	return project.GetInput{
		Filter: project.Filter{
			IDs:        r.IDs,
			Statuses:   r.Statuses,
			SearchName: r.SearchName,
		},
		PaginateQuery: pq,
	}
}
