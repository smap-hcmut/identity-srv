package http

import (
	"slices"

	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// @Summary List Subscriptions
// @Description List all subscriptions without pagination
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids query []string false "Subscription IDs"
// @Param user_ids query []string false "User IDs"
// @Param plan_ids query []string false "Plan IDs"
// @Param statuses query []string false "Statuses"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions [GET]
func (h handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	query, sc, err := h.processListSubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.List.processListSubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	subs, err := h.uc.List(ctx, sc, query.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.List.List: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.List.List: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, subscriptionListResp{Subscriptions: h.newSubscriptionListResp(subs)})
}

// @Summary Get Subscriptions (Paginated)
// @Description Get subscriptions with pagination
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids query []string false "Subscription IDs"
// @Param user_ids query []string false "User IDs"
// @Param plan_ids query []string false "Plan IDs"
// @Param statuses query []string false "Statuses"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions/page [GET]
func (h handler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	query, sc, err := h.processGetSubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Get.processGetSubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Get(ctx, sc, query.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.Get.Get: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.Get.Get: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, subscriptionPageResp{
		Subscriptions: h.newSubscriptionListResp(output.Subscriptions),
		Paginator:     output.Paginator,
	})
}

// @Summary Get Subscription Detail
// @Description Get subscription by ID
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions/{id} [GET]
func (h handler) Detail(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processDetailSubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Detail.processDetailSubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Detail(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.Detail.Detail: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.Detail.Detail: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newSubscriptionResp(output.Subscription))
}

// @Summary Get My Subscription
// @Description Get current user's active subscription
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Resp "Success"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions/me [GET]
func (h handler) GetMySubscription(c *gin.Context) {
	ctx := c.Request.Context()

	sc, err := h.processGetMySubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.GetMySubscription.processGetMySubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	sub, err := h.uc.GetActiveSubscription(ctx, sc, sc.UserID)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.GetMySubscription.GetActiveSubscription: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.GetMySubscription.GetActiveSubscription: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newSubscriptionResp(sub))
}

// @Summary Create Subscription
// @Description Create a new subscription
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param createSubscriptionReq body createSubscriptionReq true "Create Subscription"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions [POST]
func (h handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processCreateSubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Create.processCreateSubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	input, err := req.toInput()
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Create.toInput: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Create(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.Create.Create: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.Create.Create: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newSubscriptionResp(output.Subscription))
}

// @Summary Update Subscription
// @Description Update an existing subscription
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Param updateSubscriptionReq body updateSubscriptionReq true "Update Subscription"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions/{id} [PUT]
func (h handler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	req, id, sc, err := h.processUpdateSubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Update.processUpdateSubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	input, err := req.toInput(id)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Update.toInput: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Update(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.Update.Update: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.Update.Update: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newSubscriptionResp(output.Subscription))
}

// @Summary Delete Subscription
// @Description Soft delete a subscription
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions/{id} [DELETE]
func (h handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processDeleteSubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Delete.processDeleteSubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	err = h.uc.Delete(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.Delete.Delete: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.Delete.Delete: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, nil)
}

// @Summary Cancel Subscription
// @Description Cancel an active or trialing subscription
// @Tags Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /subscriptions/{id}/cancel [POST]
func (h handler) Cancel(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processCancelSubscriptionRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "subscription.http.Cancel.processCancelSubscriptionRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Cancel(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "subscription.http.Cancel.Cancel: %v", err)
		} else {
			h.l.Warnf(ctx, "subscription.http.Cancel.Cancel: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newSubscriptionResp(output.Subscription))
}
