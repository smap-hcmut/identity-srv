package http

import (
	"slices"

	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// @Summary List Plans
// @Description List all plans without pagination
// @Tags Plan
// @Accept json
// @Produce json
// @Param ids query []string false "Plan IDs"
// @Param codes query []string false "Plan codes"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /plans [GET]
func (h handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	query, sc, err := h.processListPlanRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "plan.http.List.processListPlanRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	plans, err := h.uc.List(ctx, sc, query.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "plan.http.List.List: %v", err)
		} else {
			h.l.Warnf(ctx, "plan.http.List.List: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, planListResp{Plans: h.newPlanListResp(plans)})
}

// @Summary Get Plans (Paginated)
// @Description Get plans with pagination
// @Tags Plan
// @Accept json
// @Produce json
// @Param ids query []string false "Plan IDs"
// @Param codes query []string false "Plan codes"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /plans/page [GET]
func (h handler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	query, sc, err := h.processGetPlanRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "plan.http.Get.processGetPlanRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Get(ctx, sc, query.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "plan.http.Get.Get: %v", err)
		} else {
			h.l.Warnf(ctx, "plan.http.Get.Get: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, planPageResp{
		Plans:     h.newPlanListResp(output.Plans),
		Paginator: output.Paginator,
	})
}

// @Summary Get Plan Detail
// @Description Get plan by ID
// @Tags Plan
// @Accept json
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /plans/{id} [GET]
func (h handler) Detail(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processDetailPlanRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "plan.http.Detail.processDetailPlanRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Detail(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "plan.http.Detail.Detail: %v", err)
		} else {
			h.l.Warnf(ctx, "plan.http.Detail.Detail: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newPlanResp(output.Plan))
}

// @Summary Create Plan
// @Description Create a new plan
// @Tags Plan
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param createPlanReq body createPlanReq true "Create Plan"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /plans [POST]
func (h handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processCreatePlanRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "plan.http.Create.processCreatePlanRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Create(ctx, sc, req.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "plan.http.Create.Create: %v", err)
		} else {
			h.l.Warnf(ctx, "plan.http.Create.Create: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newPlanResp(output.Plan))
}

// @Summary Update Plan
// @Description Update an existing plan
// @Tags Plan
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Plan ID"
// @Param updatePlanReq body updatePlanReq true "Update Plan"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /plans/{id} [PUT]
func (h handler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	req, id, sc, err := h.processUpdatePlanRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "plan.http.Update.processUpdatePlanRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Update(ctx, sc, req.toInput(id))
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "plan.http.Update.Update: %v", err)
		} else {
			h.l.Warnf(ctx, "plan.http.Update.Update: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newPlanResp(output.Plan))
}

// @Summary Delete Plan
// @Description Soft delete a plan
// @Tags Plan
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Plan ID"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /plans/{id} [DELETE]
func (h handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processDeletePlanRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "plan.http.Delete.processDeletePlanRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	err = h.uc.Delete(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "plan.http.Delete.Delete: %v", err)
		} else {
			h.l.Warnf(ctx, "plan.http.Delete.Delete: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, nil)
}
