package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/models"
	pkgErrors "github.com/nguyentantai21042004/smap-api/pkg/errors"
	"github.com/nguyentantai21042004/smap-api/pkg/scope"
)

func (h handler) processDetailMeRequest(c *gin.Context) (models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "internal.user.delivery.http.processDetailMeRequest.jwt.GetPayloadFromContext: %v", "payload not found")
		return models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	sc := scope.NewScope(payload)
	return sc, nil
}

func (h handler) processDetailRequest(c *gin.Context) (string, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "internal.user.delivery.http.processDetailMeRequest.jwt.GetPayloadFromContext: %v", "payload not found")
		return "", models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	ID := c.Param("id")
	if ID == "" {
		h.l.Warnf(ctx, "internal.user.http.Detail.ID: %v", "ID is required")
		return "", models.Scope{}, errWrongQuery
	}

	sc := scope.NewScope(payload)
	return ID, sc, nil
}

func (h handler) processUpdateAvatarRequest(c *gin.Context) (updateAvatarReq, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "internal.user.delivery.http.processUpdateAvatarRequest.jwt.GetPayloadFromContext: %v", "payload not found")
		return updateAvatarReq{}, models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	var req updateAvatarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "internal.user.delivery.http.processUpdateAvatarRequest.ShouldBindJSON: %v", err)
		return updateAvatarReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "internal.user.delivery.http.processUpdateAvatarRequest.validate: %v", err)
		return updateAvatarReq{}, models.Scope{}, err
	}

	return req, scope.NewScope(payload), nil
}
