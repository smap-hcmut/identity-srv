package http

import (
	"time"

	"smap-project/internal/model"
	"smap-project/internal/project"
	"smap-project/pkg/paginator"
	"smap-project/pkg/scope"

	"github.com/gin-gonic/gin"
)

// CreateReq represents the HTTP request for creating a project
type CreateReq struct {
	Name                  string              `json:"name" binding:"required"`
	Description           *string             `json:"description"`
	Status                string              `json:"status" binding:"required"`
	FromDate              time.Time           `json:"from_date" binding:"required"`
	ToDate                time.Time           `json:"to_date" binding:"required"`
	BrandName             string              `json:"brand_name" binding:"required"`
	CompetitorNames       []string            `json:"competitor_names"`
	BrandKeywords         []string            `json:"brand_keywords" binding:"required"`
	CompetitorKeywordsMap map[string][]string `json:"competitor_keywords_map"`
	ExcludeKeywords       []string            `json:"exclude_keywords"`
}

func (r CreateReq) toInput() project.CreateInput {
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
		ExcludeKeywords:       r.ExcludeKeywords,
	}
}

// UpdateReq represents the HTTP request for updating a project
type UpdateReq struct {
	Name                  *string             `json:"name"`
	Description           *string             `json:"description"`
	Status                *string             `json:"status"`
	FromDate              *time.Time          `json:"from_date"`
	ToDate                *time.Time          `json:"to_date"`
	BrandName             *string             `json:"brand_name"`
	CompetitorNames       []string            `json:"competitor_names"`
	BrandKeywords         []string            `json:"brand_keywords"`
	CompetitorKeywordsMap map[string][]string `json:"competitor_keywords_map"`
	ExcludeKeywords       []string            `json:"exclude_keywords"`
}

func (r UpdateReq) toInput(id string) project.UpdateInput {
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
		ExcludeKeywords:       r.ExcludeKeywords,
	}
}

// GetReq represents the HTTP request for listing projects with filters
type GetReq struct {
	IDs        []string `form:"ids"`
	Statuses   []string `form:"statuses"`
	SearchName *string  `form:"search_name"`
	Page       int      `form:"page"`
	Limit      int      `form:"limit"`
}

func (r GetReq) toInput() project.GetInput {
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

func (r GetReq) toListInput() project.ListInput {
	return project.ListInput{
		Filter: project.Filter{
			IDs:        r.IDs,
			Statuses:   r.Statuses,
			SearchName: r.SearchName,
		},
	}
}

// Process functions
func (h handler) processListReq(c *gin.Context) (project.ListInput, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processListReq: unauthorized")
		return project.ListInput{}, model.Scope{}, errUnauthorized
	}

	var req GetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processListReq.ShouldBindQuery: %v", err)
		return project.ListInput{}, model.Scope{}, errWrongQuery
	}

	return req.toListInput(), sc, nil
}

func (h handler) processGetReq(c *gin.Context) (project.GetInput, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processGetReq: unauthorized")
		return project.GetInput{}, model.Scope{}, errUnauthorized
	}

	var req GetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processGetReq.ShouldBindQuery: %v", err)
		return project.GetInput{}, model.Scope{}, errWrongQuery
	}

	return req.toInput(), sc, nil
}

func (h handler) processDetailReq(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processDetailReq: unauthorized")
		return "", model.Scope{}, errUnauthorized
	}

	id := c.Param("id")
	if id == "" {
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}

func (h handler) processCreateReq(c *gin.Context) (project.CreateInput, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processCreateReq: unauthorized")
		return project.CreateInput{}, model.Scope{}, errUnauthorized
	}

	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processCreateReq.ShouldBindJSON: %v", err)
		return project.CreateInput{}, model.Scope{}, errWrongBody
	}

	return req.toInput(), sc, nil
}

func (h handler) processUpdateReq(c *gin.Context) (project.UpdateInput, string, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processUpdateReq: unauthorized")
		return project.UpdateInput{}, "", model.Scope{}, errUnauthorized
	}

	id := c.Param("id")
	if id == "" {
		return project.UpdateInput{}, "", model.Scope{}, errInvalidID
	}

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processUpdateReq.ShouldBindJSON: %v", err)
		return project.UpdateInput{}, "", model.Scope{}, errWrongBody
	}

	return req.toInput(id), id, sc, nil
}

func (h handler) processDeleteReq(c *gin.Context) (string, model.Scope, error) {
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

// SuggestKeywordsReq represents the HTTP request for suggesting keywords
type SuggestKeywordsReq struct {
	BrandName string `json:"brand_name" binding:"required"`
}

func (h handler) processSuggestKeywordsReq(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processSuggestKeywordsReq: unauthorized")
		return "", model.Scope{}, errUnauthorized
	}

	var req SuggestKeywordsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processSuggestKeywordsReq.ShouldBindJSON: %v", err)
		return "", model.Scope{}, errWrongBody
	}

	return req.BrandName, sc, nil
}

// DryRunKeywordsReq represents the HTTP request for dry running keywords
type DryRunKeywordsReq struct {
	Keywords []string `json:"keywords" binding:"required"`
}

func (h handler) processDryRunKeywordsReq(c *gin.Context) ([]string, model.Scope, error) {
	ctx := c.Request.Context()

	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "project.http.processDryRunKeywordsReq: unauthorized")
		return nil, model.Scope{}, errUnauthorized
	}

	var req DryRunKeywordsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "project.http.processDryRunKeywordsReq.ShouldBindJSON: %v", err)
		return nil, model.Scope{}, errWrongBody
	}

	return req.Keywords, sc, nil
}
