package http

import (
	"smap-api/internal/model"

	"github.com/gin-gonic/gin"
)

func (h handler) processCreateSubscriptionRequest(c *gin.Context) (createSubscriptionReq, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)

	var req createSubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "subscription.http.processCreateSubscriptionRequest.ShouldBindJSON: %v", err)
		return createSubscriptionReq{}, model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "subscription.http.processCreateSubscriptionRequest.validate: %v", err)
		return createSubscriptionReq{}, model.Scope{}, err
	}

	return req, sc, nil
}

func (h handler) processUpdateSubscriptionRequest(c *gin.Context) (updateSubscriptionReq, string, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)
	id := c.Param("id")

	if id == "" {
		h.l.Errorf(ctx, "subscription.http.processUpdateSubscriptionRequest: %v", errInvalidID)
		return updateSubscriptionReq{}, "", model.Scope{}, errInvalidID
	}

	var req updateSubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "subscription.http.processUpdateSubscriptionRequest.ShouldBindJSON: %v", err)
		return updateSubscriptionReq{}, "", model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "subscription.http.processUpdateSubscriptionRequest.validate: %v", err)
		return updateSubscriptionReq{}, "", model.Scope{}, err
	}

	return req, id, sc, nil
}

func (h handler) processListSubscriptionRequest(c *gin.Context) (listSubscriptionQuery, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)

	var query listSubscriptionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.l.Errorf(ctx, "subscription.http.processListSubscriptionRequest.ShouldBindQuery: %v", err)
		return listSubscriptionQuery{}, model.Scope{}, errWrongBody
	}

	return query, sc, nil
}

func (h handler) processGetSubscriptionRequest(c *gin.Context) (getSubscriptionQuery, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)

	var query getSubscriptionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.l.Errorf(ctx, "subscription.http.processGetSubscriptionRequest.ShouldBindQuery: %v", err)
		return getSubscriptionQuery{}, model.Scope{}, errWrongBody
	}

	return query, sc, nil
}

func (h handler) processDetailSubscriptionRequest(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)
	id := c.Param("id")

	if id == "" {
		h.l.Errorf(ctx, "subscription.http.processDetailSubscriptionRequest: %v", errInvalidID)
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}

func (h handler) processDeleteSubscriptionRequest(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)
	id := c.Param("id")

	if id == "" {
		h.l.Errorf(ctx, "subscription.http.processDeleteSubscriptionRequest: %v", errInvalidID)
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}

func (h handler) processCancelSubscriptionRequest(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()
	sc := c.MustGet("scope").(model.Scope)
	id := c.Param("id")

	if id == "" {
		h.l.Errorf(ctx, "subscription.http.processCancelSubscriptionRequest: %v", errInvalidID)
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}

func (h handler) processGetMySubscriptionRequest(c *gin.Context) (model.Scope, error) {
	sc := c.MustGet("scope").(model.Scope)
	return sc, nil
}

func (h handler) processGetUserSubscriptionRequest(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()

	sc := c.MustGet("scope").(model.Scope)
	id := c.Param("id")

	if id == "" {
		h.l.Errorf(ctx, "subscription.http.processGetUserSubscriptionRequest: %v", errInvalidID)
		return "", model.Scope{}, errInvalidID
	}

	return id, sc, nil
}
