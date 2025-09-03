package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/pkg/response"
)

// @Summary Create role
// @Schemes
// @Description Create role
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token"
// @Param lang header string false "Language" default(en)
// @Param name body string true "Name"
// @Produce json
// @Tags Roles
// @Accept json
// @Produce json
// @Success 200 {object} detailResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/roles [POST]
func (h handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processCreateRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "role.delivery.http.Create.processCreateRequest: %v", err)
		response.Error(c, h.mapError(err), h.d)
		return
	}

	e, err := h.uc.Create(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapError(err)
		h.l.Errorf(ctx, "role.delivery.http.Create.Create: %v", err)
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, h.newDetailResp(e))
}


// @Summary Update role
// @Schemes
// @Description Update role
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token"
// @Param lang header string false "Language" default(en)
// @Param id body string true "ID"
// @Param name body string true "Name"
// @Produce json
// @Tags Roles
// @Accept json
// @Produce json
// @Success 200 {object} detailResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/roles [PUT]
func (h handler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processUpdateRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "role.delivery.http.Update.processUpdateRequest: %v", err)
		response.Error(c, h.mapError(err), h.d)
		return
	}

	e, err := h.uc.Update(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapError(err)
		h.l.Errorf(ctx, "role.delivery.http.Update.Update: %v", err)
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, h.newUpdateResp(e))
}

// @Summary Delete role
// @Schemes
// @Description Delete role
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token"
// @Param lang header string false "Language" default(en)
// @Param ids body []string true "IDs"
// @Produce json
// @Tags Roles
// @Accept json
// @Produce json
// @Success 200 {object} interface{}
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/roles [DELETE]
func (h handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processDeleteRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "role.delivery.http.Delete.processDeleteRequest: %v", err)
		response.Error(c, h.mapError(err), h.d)
		return
	}

	err = h.uc.Delete(ctx, sc, req)
	if err != nil {
		mapErr := h.mapError(err)
		h.l.Errorf(ctx, "role.delivery.http.Delete.Delete: %v", err)
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, nil)
}

// @Summary Get single role
// @Schemes
// @Description Get single role (public API)
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token"
// @Param lang header string false "Language" default(en)
// @Param id path string true "ID"
// @Produce json
// @Tags Roles
// @Accept json
// @Produce json
// @Success 200 {object} detailResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/roles/{id} [GET]
func (h handler) GetOne(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processGetOneRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "role.delivery.http.GetOne.processGetOneRequest: %v", err)
		response.Error(c, h.mapError(err), h.d)
		return
	}

	o, err := h.uc.GetOne(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapError(err)
		h.l.Errorf(ctx, "role.delivery.http.GetOne.GetOne: %v", err)
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, h.newGetOneResp(o))
}

// @Summary Get roles
// @Schemes
// @Description Get roles with pagination (public API)
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token"
// @Param lang header string false "Language" default(en)
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(15)
// @Param ids query []string false "IDs"
// @Param alias query []string false "Alias"
// @Param code query []string false "Code"
// @Produce json
// @Tags Roles
// @Accept json
// @Produce json
// @Success 200 {object} getResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /api/v1/roles [GET]
func (h handler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	req, pq, sc, err := h.processGetRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "role.delivery.http.Get.processGetRequest: %v", err)
		response.Error(c, h.mapError(err), h.d)
		return
	}

	inp := req.toInput()
	pq.Adjust()
	inp.PagQuery = pq

	o, err := h.uc.Get(ctx, sc, inp)
	if err != nil {
		mapErr := h.mapError(err)
		h.l.Errorf(ctx, "role.delivery.http.Get.Get: %v", err)
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, h.newGetResp(o))
}
