package http

import (
	"time"

	"smap-project/internal/model"
	"smap-project/internal/project"
	"smap-project/pkg/paginator"
	"smap-project/pkg/scope"

	"github.com/gin-gonic/gin"
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

func (r CreateProjectRequest) toInput() project.CreateInput {
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

func (r UpdateProjectRequest) toInput(id string) project.UpdateInput {
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

// GetProjectsRequest represents the HTTP request for listing projects with filters
type GetProjectsRequest struct {
	IDs        []string `form:"ids"`
	Statuses   []string `form:"statuses"`
	SearchName *string  `form:"search_name"`
	Page       int      `form:"page"`
	Limit      int      `form:"limit"`
}

func (r GetProjectsRequest) toInput() project.GetInput {
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

func (r GetProjectsRequest) toListInput() project.ListInput {
	return project.ListInput{
		Filter: project.Filter{
			IDs:        r.IDs,
			Statuses:   r.Statuses,
			SearchName: r.SearchName,
		},
	}
}

// Process functions
func (h handler) processListRequest(c *gin.Context) (project.ListInput, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processListRequest: unauthorized")
		return project.ListInput{}, model.Scope{}, errUnauthorized
	}

	var req GetProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processListRequest.ShouldBindQuery: %v", err)
		return project.ListInput{}, model.Scope{}, errWrongQuery
	}

	return req.toListInput(), sc, nil
}

func (h handler) processGetRequest(c *gin.Context) (project.GetInput, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processGetRequest: unauthorized")
		return project.GetInput{}, model.Scope{}, errUnauthorized
	}

	var req GetProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processGetRequest.ShouldBindQuery: %v", err)
		return project.GetInput{}, model.Scope{}, errWrongQuery
	}

	return req.toInput(), sc, nil
}

func (h handler) processDetailRequest(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processDetailRequest: unauthorized")
		return "", model.Scope{}, errUnauthorized
	}

	id := c.Param("id")
	if id == "" {
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}

func (h handler) processCreateRequest(c *gin.Context) (project.CreateInput, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processCreateRequest: unauthorized")
		return project.CreateInput{}, model.Scope{}, errUnauthorized
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processCreateRequest.ShouldBindJSON: %v", err)
		return project.CreateInput{}, model.Scope{}, errWrongBody
	}

	return req.toInput(), sc, nil
}

func (h handler) processUpdateRequest(c *gin.Context) (project.UpdateInput, string, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processUpdateRequest: unauthorized")
		return project.UpdateInput{}, "", model.Scope{}, errUnauthorized
	}

	id := c.Param("id")
	if id == "" {
		return project.UpdateInput{}, "", model.Scope{}, errInvalidID
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processUpdateRequest.ShouldBindJSON: %v", err)
		return project.UpdateInput{}, "", model.Scope{}, errWrongBody
	}

	return req.toInput(id), id, sc, nil
}

func (h handler) processDeleteRequest(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processDeleteRequest: unauthorized")
		return "", model.Scope{}, errUnauthorized
	}

	id := c.Param("id")
	if id == "" {
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}
