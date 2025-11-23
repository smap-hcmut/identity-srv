package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	pkgErrors "gitlab.com/gma-vietnam/tanca-connect/pkg/errors"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h handler) processCreateRequest(c *gin.Context) (createReq, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := jwt.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "event.delivery.http.processCreateRequest.GetPayloadFromContext: unauthorized")
		return createReq{}, models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processCreateRequest.ShouldBindJSON: %v", errWrongBody)
		return createReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processCreateRequest.validate: %v", errWrongBody)
		return createReq{}, models.Scope{}, errWrongBody
	}

	sc := jwt.NewScope(payload)

	return req, sc, nil
}

func (h handler) processDetailRequest(c *gin.Context) (string, string, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := jwt.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "event.delivery.http.processDetailRequest.GetPayloadFromContext: unauthorized")
		return "", "", models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	id := c.Param("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processDetailRequest.ObjectIDFromHex: %v", err)
		return "", "", models.Scope{}, errWrongBody
	}

	eventID := c.Param("event_id")
	if _, err := primitive.ObjectIDFromHex(eventID); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processDetailRequest.ObjectIDFromHex: %v", err)
		return "", "", models.Scope{}, errWrongQuery
	}

	sc := jwt.NewScope(payload)

	return id, eventID, sc, nil
}

func (h handler) processUpdateRequest(c *gin.Context) (updateReq, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := jwt.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateRequest.GetPayloadFromContext: unauthorized")
		return updateReq{}, models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	var req updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateRequest.ShouldBindJSON: %v", err)
		return updateReq{}, models.Scope{}, errWrongBody
	}

	if err := c.ShouldBindUri(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateRequest.ShouldBindUri: %v", errWrongBody)
		return updateReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateRequest.validate: %v", errWrongBody)
		return updateReq{}, models.Scope{}, errWrongBody
	}

	if _, err := primitive.ObjectIDFromHex(req.ID); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateRequest.ObjectIDFromHex: %v", err)
		return updateReq{}, models.Scope{}, errWrongBody
	}

	sc := jwt.NewScope(payload)

	return req, sc, nil
}

func (h handler) processDeleteRequest(c *gin.Context) (deleteReq, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := jwt.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "event.delivery.http.processDeleteRequest.GetPayloadFromContext: unauthorized")
		return deleteReq{}, models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	var req deleteReq
	if err := c.ShouldBindUri(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processDeleteRequest.ShouldBindUri: %v", errWrongQuery)
		return deleteReq{}, models.Scope{}, errWrongQuery
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processDeleteRequest.ShouldBindQuery: %v", errWrongQuery)
		return deleteReq{}, models.Scope{}, errWrongQuery
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processDeleteRequest.validate: %v", err)
		return deleteReq{}, models.Scope{}, err
	}

	_, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processDeleteRequest.ObjectIDFromHex: %v", err)
		return deleteReq{}, models.Scope{}, errWrongBody
	}

	sc := jwt.NewScope(payload)

	return req, sc, nil
}

func (h handler) processCalendarRequest(c *gin.Context) (calendarReq, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := jwt.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "event.delivery.http.processCalendarRequest.GetPayloadFromContext: unauthorized")
		return calendarReq{}, models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	var req calendarReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processCalendarRequest.ShouldBindQuery: %v", errWrongBody)
		return calendarReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processCalendarRequest.validate: %v", errWrongBody)
		return calendarReq{}, models.Scope{}, errWrongBody
	}

	sc := jwt.NewScope(payload)

	return req, sc, nil
}

func (h handler) processUpdateAttendanceRequest(c *gin.Context) (updateAttendanceReq, models.Scope, error) {
	ctx := c.Request.Context()

	payload, ok := jwt.GetPayloadFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateAttendanceRequest.GetPayloadFromContext: unauthorized")
		return updateAttendanceReq{}, models.Scope{}, pkgErrors.NewUnauthorizedHTTPError()
	}

	var req updateAttendanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateAttendanceRequest.ShouldBindJSON: %v", errWrongBody)
		return updateAttendanceReq{}, models.Scope{}, errWrongBody
	}

	if err := c.ShouldBindUri(&req); err != nil {
		h.l.Errorf(ctx, "event.delivery.http.processUpdateAttendanceRequest.ShouldBindUri: %v", errWrongBody)
		return updateAttendanceReq{}, models.Scope{}, errWrongBody
	}

	sc := jwt.NewScope(payload)

	return req, sc, nil
}
