package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h handler) processCreateRequest(c *gin.Context) (createReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processCreateRequest.ShouldBindJSON: %v", err)
		return createReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}

func (h handler) processGetOneRequest(c *gin.Context) (getOneReq, models.Scope, error) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processGetOneRequest.ObjectIDFromHex: %v", err)
		return getOneReq{}, models.Scope{}, errWrongQuery
	}

	return getOneReq{ID: id}, models.Scope{}, nil
}

func (h handler) processUpdateRequest(c *gin.Context) (updateReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processUpdateRequest.ShouldBindJSON: %v", err)
		return updateReq{}, models.Scope{}, errWrongBody
	}

	if _, err := primitive.ObjectIDFromHex(req.ID); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processUpdateRequest.ObjectIDFromHex: %v", err)
		return updateReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}

func (h handler) processDeleteRequest(c *gin.Context) ([]string, models.Scope, error) {
	ctx := c.Request.Context()

	var req deleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processDeleteRequest.ShouldBindJSON: %v", err)
		return nil, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processDeleteRequest.validate: %v", err)
		return nil, models.Scope{}, errWrongBody
	}

	return req.IDs, models.Scope{}, nil
}

func (h handler) processGetRequest(c *gin.Context) (getReq, paginator.PaginateQuery, models.Scope, error) {
	ctx := c.Request.Context()

	var req getReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processGetRequest.ShouldBindQuery: %v", err)
		return getReq{}, paginator.PaginateQuery{}, models.Scope{}, errWrongQuery
	}

	var pq paginator.PaginateQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		h.l.Warnf(ctx, "role.delivery.http.processGetRequest.ShouldBindQuery: %v", err)
		return getReq{}, paginator.PaginateQuery{}, models.Scope{}, errWrongQuery
	}

	return req, pq, models.Scope{}, nil
}
