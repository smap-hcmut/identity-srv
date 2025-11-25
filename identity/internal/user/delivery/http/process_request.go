package http

import (
	"smap-api/internal/model"
	"smap-api/internal/user"
	"smap-api/pkg/errors"

	"smap-api/pkg/scope"

	"github.com/gin-gonic/gin"
)

func (h handler) processListRequest(c *gin.Context) (user.ListInput, model.Scope, error) {
	ctx := c.Request.Context()

	p, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "user.delivery.http.processListRequest: %v", errors.NewUnauthorizedHTTPError())
		return user.ListInput{}, model.Scope{}, errors.NewUnauthorizedHTTPError()
	}

	var req ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "user.delivery.http.processListRequest: %v", errWrongBody)
		return user.ListInput{}, model.Scope{}, errWrongBody
	}

	sc := scope.NewScope(p)

	return req.toInput(), sc, nil
}

func (h handler) processGetRequest(c *gin.Context) (user.GetInput, model.Scope, error) {
	ctx := c.Request.Context()

	p, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "user.delivery.http.processGetRequest: %v", errors.NewUnauthorizedHTTPError())
		return user.GetInput{}, model.Scope{}, errors.NewUnauthorizedHTTPError()
	}

	var req GetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "user.delivery.http.processGetRequest: %v", errWrongBody)
		return user.GetInput{}, model.Scope{}, errWrongBody
	}

	sc := scope.NewScope(p)

	return req.toInput(), sc, nil
}

func (h handler) processUpdateProfileRequest(c *gin.Context) (user.UpdateProfileInput, model.Scope, error) {
	ctx := c.Request.Context()

	p, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "user.delivery.http.processUpdateProfileRequest: %v", errors.NewUnauthorizedHTTPError())
		return user.UpdateProfileInput{}, model.Scope{}, errors.NewUnauthorizedHTTPError()
	}

	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "user.delivery.http.processUpdateProfileRequest: %v", errWrongBody)
		return user.UpdateProfileInput{}, model.Scope{}, errWrongBody
	}

	sc := scope.NewScope(p)

	return req.toInput(), sc, nil
}

func (h handler) processChangePasswordRequest(c *gin.Context) (user.ChangePasswordInput, model.Scope, error) {
	ctx := c.Request.Context()

	p, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "user.delivery.http.processChangePasswordRequest: %v", errors.NewUnauthorizedHTTPError())
		return user.ChangePasswordInput{}, model.Scope{}, errors.NewUnauthorizedHTTPError()
	}

	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "user.delivery.http.processChangePasswordRequest: %v", errWrongBody)
		return user.ChangePasswordInput{}, model.Scope{}, errWrongBody
	}

	sc := scope.NewScope(p)

	return req.toInput(), sc, nil
}

func (h handler) processIDParam(c *gin.Context) (string, model.Scope, error) {
	ctx := c.Request.Context()

	p, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "user.delivery.http.processIDParam: %v", errors.NewUnauthorizedHTTPError())
		return "", model.Scope{}, errors.NewUnauthorizedHTTPError()
	}

	id := c.Param("id")
	if id == "" {
		h.l.Errorf(ctx, "user.delivery.http.processIDParam: %v", errWrongQuery)
		return "", model.Scope{}, errWrongQuery
	}

	sc := scope.NewScope(p)

	return id, sc, nil
}
