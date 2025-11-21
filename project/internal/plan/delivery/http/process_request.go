package http

import (
	"smap-api/internal/model"

	"github.com/gin-gonic/gin"
)

func (h handler) processCreatePlanRequest(c *gin.Context) (createPlanReq, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)

	var req createPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "plan.http.processCreatePlanRequest.ShouldBindJSON: %v", err)
		return createPlanReq{}, model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "plan.http.processCreatePlanRequest.validate: %v", err)
		return createPlanReq{}, model.Scope{}, err
	}

	return req, sc, nil
}

func (h handler) processUpdatePlanRequest(c *gin.Context) (updatePlanReq, string, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)
	id := c.Param("id")

	if id == "" {
		return updatePlanReq{}, "", model.Scope{}, errInvalidID
	}

	var req updatePlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "plan.http.processUpdatePlanRequest.ShouldBindJSON: %v", err)
		return updatePlanReq{}, "", model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "plan.http.processUpdatePlanRequest.validate: %v", err)
		return updatePlanReq{}, "", model.Scope{}, err
	}

	return req, id, sc, nil
}

func (h handler) processListPlanRequest(c *gin.Context) (listPlanQuery, model.Scope, error) {
	ctx := c.Request.Context()

	var query listPlanQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.l.Warnf(ctx, "plan.http.processListPlanRequest.ShouldBindQuery: %v", err)
		return listPlanQuery{}, model.Scope{}, errWrongBody
	}

	return query, model.Scope{}, nil
}

func (h handler) processGetPlanRequest(c *gin.Context) (getPlanQuery, model.Scope, error) {
	ctx := c.Request.Context()

	var query getPlanQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.l.Warnf(ctx, "plan.http.processGetPlanRequest.ShouldBindQuery: %v", err)
		return getPlanQuery{}, model.Scope{}, errWrongBody
	}

	return query, model.Scope{}, nil
}

func (h handler) processDetailPlanRequest(c *gin.Context) (string, model.Scope, error) {
	id := c.Param("id")

	if id == "" {
		return "", model.Scope{}, errInvalidID
	}

	return id, model.Scope{}, nil
}

func (h handler) processDeletePlanRequest(c *gin.Context) (string, model.Scope, error) {
	sc := c.MustGet("scope").(model.Scope)
	id := c.Param("id")

	if id == "" {
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}

