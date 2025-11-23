package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/response"
)

// @Summary Create event
// @Schemes
// @Description Create event
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token" default(Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vcC50YW5jYS52bi9hcGkvdjQvYXV0aC9zaWduaW4tdjIiLCJpYXQiOjE3MTY1ODUyNDAsIm5iZiI6MTcxNjU4NTI0MCwianRpIjoidFBJMldUa0JldThYdnJMZiIsInN1YiI6Ik5pdEpwZUp1dkF4M1pjYXdKIiwicHJ2IjoiMWM1NTIwZjcwYmFhNjU1ZGRjNTc0NmE2NzY0ZjM3MmExYjY1NWFhNiIsInNob3BfaWQiOiI1YzIwYTE5YzBiMDg4ODBmNTk0ZmM0NjgiLCJzaG9wX3VzZXJuYW1lIjoicmF2ZSIsInNob3BfcHJlZml4IjoidCIsInR5cGUiOiJhcGkifQ.DnxirM5IXQY3B9Vcc6Qco7c9f0ABGjoeLu_1LfHiRjE)"
// @Param lang header string false "Language" default(en)
// @Param body body createReq true "Body"
// @Produce json
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} detailResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/events [POST]
func (h handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processCreateRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Create.processCreateRequest: %v", err)
		response.Error(c, err)
		return
	}

	o, err := h.uc.Create(ctx, sc, req.toInput())
	if err != nil {
		if errors.Is(err, event.ErrRoomUnavailable) {
			response.ErrorWithData(c, errRoomUnavailable, h.newUnavailableRoomResp(o.UnavailableRooms))
			return
		}

		h.l.Errorf(ctx, "event.event.delivery.http.Create.Create: %v", err)
		mapErr := h.mapError(err)
		response.Error(c, mapErr)
		return
	}

	response.OK(c, nil)
}

// @Summary Get event detail
// @Schemes
// @Description Get event detail
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token" default(Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vcC50YW5jYS52bi9hcGkvdjQvYXV0aC9zaWduaW4tdjIiLCJpYXQiOjE3MTY1ODUyNDAsIm5iZiI6MTcxNjU4NTI0MCwianRpIjoidFBJMldUa0JldThYdnJMZiIsInN1YiI6Ik5pdEpwZUp1dkF4M1pjYXdKIiwicHJ2IjoiMWM1NTIwZjcwYmFhNjU1ZGRjNTc0NmE2NzY0ZjM3MmExYjY1NWFhNiIsInNob3BfaWQiOiI1YzIwYTE5YzBiMDg4ODBmNTk0ZmM0NjgiLCJzaG9wX3VzZXJuYW1lIjoicmF2ZSIsInNob3BfcHJlZml4IjoidCIsInR5cGUiOiJhcGkifQ.DnxirM5IXQY3B9Vcc6Qco7c9f0ABGjoeLu_1LfHiRjE)"
// @Param lang header string false "Language" default(en)
// @Param id path string true "ID"
// @Produce json
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} detailResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/events/{event_id}/{id} [GET]
func (h handler) Detail(c *gin.Context) {
	ctx := c.Request.Context()

	id, eventID, sc, err := h.processDetailRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Detail.processDetailRequest: %v", err)
		response.Error(c, err)
		return
	}

	e, err := h.uc.Detail(ctx, sc, id, eventID)
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Detail.Detail: %v", err)
		mapErr := h.mapError(err)
		response.Error(c, mapErr)
		return
	}

	response.OK(c, h.newDetailResp(e))
}

// @Summary Update event
// @Schemes
// @Description Update event
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token" default(Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vcC50YW5jYS52bi9hcGkvdjQvYXV0aC9zaWduaW4tdjIiLCJpYXQiOjE3MTY1ODUyNDAsIm5iZiI6MTcxNjU4NTI0MCwianRpIjoidFBJMldUa0JldThYdnJMZiIsInN1YiI6Ik5pdEpwZUp1dkF4M1pjYXdKIiwicHJ2IjoiMWM1NTIwZjcwYmFhNjU1ZGRjNTc0NmE2NzY0ZjM3MmExYjY1NWFhNiIsInNob3BfaWQiOiI1YzIwYTE5YzBiMDg4ODBmNTk0ZmM0NjgiLCJzaG9wX3VzZXJuYW1lIjoicmF2ZSIsInNob3BfcHJlZml4IjoidCIsInR5cGUiOiJhcGkifQ.DnxirM5IXQY3B9Vcc6Qco7c9f0ABGjoeLu_1LfHiRjE)"
// @Param lang header string false "Language" default(en)
// @Param body body updateReq true "Body"
// @Produce json
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} detailResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/events/{id} [PUT]
func (h handler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processUpdateRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Update.processUpdateRequest: %v", err)
		response.Error(c, err)
		return
	}

	err = h.uc.Update(ctx, sc, req.toInput())
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Update.Update: %v", err)
		mapErr := h.mapError(err)
		response.Error(c, mapErr)
		return
	}

	response.OK(c, nil)
}

// @Summary Delete event
// @Schemes
// @Description Delete event
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token" default(Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vcC50YW5jYS52bi9hcGkvdjQvYXV0aC9zaWduaW4tdjIiLCJpYXQiOjE3MTY1ODUyNDAsIm5iZiI6MTcxNjU4NTI0MCwianRpIjoidFBJMldUa0JldThYdnJMZiIsInN1YiI6Ik5pdEpwZUp1dkF4M1pjYXdKIiwicHJ2IjoiMWM1NTIwZjcwYmFhNjU1ZGRjNTc0NmE2NzY0ZjM3MmExYjY1NWFhNiIsInNob3BfaWQiOiI1YzIwYTE5YzBiMDg4ODBmNTk0ZmM0NjgiLCJzaG9wX3VzZXJuYW1lIjoicmF2ZSIsInNob3BfcHJlZml4IjoidCIsInR5cGUiOiJhcGkifQ.DnxirM5IXQY3B9Vcc6Qco7c9f0ABGjoeLu_1LfHiRjE)"
// @Param lang header string false "Language" default(en)
// @Param type query string true "Type" default(one)
// @Produce json
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} interface{}
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/events/{event_id}/{id} [DELETE]
func (h handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processDeleteRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Delete.processDeleteRequest: %v", err)
		response.Error(c, err)
		return
	}

	err = h.uc.Delete(ctx, sc, req.toInput())
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Delete.Delete: %v", err)
		mapErr := h.mapError(err)
		response.Error(c, mapErr)
		return
	}

	response.OK(c, nil)
}

// @Summary Get all events for select
// @Schemes
// @Description Get all events for select
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token" default(Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vcC50YW5jYS52bi9hcGkvdjQvYXV0aC9zaWduaW4tdjIiLCJpYXQiOjE3MTY1ODUyNDAsIm5iZiI6MTcxNjU4NTI0MCwianRpIjoidFBJMldUa0JldThYdnJMZiIsInN1YiI6Ik5pdEpwZUp1dkF4M1pjYXdKIiwicHJ2IjoiMWM1NTIwZjcwYmFhNjU1ZGRjNTc0NmE2NzY0ZjM3MmExYjY1NWFhNiIsInNob3BfaWQiOiI1YzIwYTE5YzBiMDg4ODBmNTk0ZmM0NjgiLCJzaG9wX3VzZXJuYW1lIjoicmF2ZSIsInNob3BfcHJlZml4IjoidCIsInR5cGUiOiJhcGkifQ.DnxirM5IXQY3B9Vcc6Qco7c9f0ABGjoeLu_1LfHiRjE)"
// @Param lang header string false "Language" default(en)
// @Param ids query []string false "IDs"
// @Param start_time query string true "Start time"
// @Param end_time query string true "End time"
// @Produce json
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} []calendarItemResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/calendar [GET]
func (h handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processCalendarRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Calendar.processCalendarRequest: %v", err)
		response.Error(c, err)
		return
	}

	o, err := h.uc.Calendar(ctx, sc, req.toInput())
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.Calendar.Calendar: %v", err)
		mapErr := h.mapError(err)
		response.Error(c, mapErr)
		return
	}

	response.OK(c, h.newCalendarResp(o))
}

// @Summary Update event attendance
// @Schemes
// @Description Update event attendance
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token" default(Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vcC50YW5jYS52bi9hcGkvdjQvYXV0aC9zaWduaW4tdjIiLCJpYXQiOjE3MTY1ODUyNDAsIm5iZiI6MTcxNjU4NTI0MCwianRpIjoidFBJMldUa0JldThYdnJMZiIsInN1YiI6Ik5pdEpwZUp1dkF4M1pjYXdKIiwicHJ2IjoiMWM1NTIwZjcwYmFhNjU1ZGRjNTc0NmE2NzY0ZjM3MmExYjY1NWFhNiIsInNob3BfaWQiOiI1YzIwYTE5YzBiMDg4ODBmNTk0ZmM0NjgiLCJzaG9wX3VzZXJuYW1lIjoicmF2ZSIsInNob3BfcHJlZml4IjoidCIsInR5cGUiOiJhcGkifQ.DnxirM5IXQY3B9Vcc6Qco7c9f0ABGjoeLu_1LfHiRjE)"
// @Param lang header string false "Language" default(en)
// @Param event_id path string true "Event ID"
// @Param id path string true "ID"
// @Param status body int true "Status"
// @Produce json
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} interface{}
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/events/attendance/{event_id}/{id} [PATCH]
func (h handler) UpdateAttendance(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processUpdateAttendanceRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.UpdateAttendance.processUpdateAttendanceRequest: %v", err)
		response.Error(c, err)
		return
	}

	err = h.uc.UpdateAttendance(ctx, sc, req.toInput())
	if err != nil {
		h.l.Errorf(ctx, "event.event.delivery.http.UpdateAttendance.UpdateAttendance: %v", err)
		mapErr := h.mapError(err)
		response.Error(c, mapErr)
		return
	}

	response.OK(c, nil)
}
