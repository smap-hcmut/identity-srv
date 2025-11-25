package http

import (
	"slices"
	"smap-api/pkg/response"
	"smap-api/pkg/scope"

	"github.com/gin-gonic/gin"
)

// GetMe godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} UserResp
// @Failure 401 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/me [get]
func (h handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	p, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		return
	}
	sc := scope.NewScope(p)

	o, err := h.uc.DetailMe(ctx, sc)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "user.delivery.http.GetMe.DetailMe: %v", err)
		} else {
			h.l.Warnf(ctx, "user.delivery.http.GetMe.DetailMe: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newUserResp(o))
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Update the profile information of the currently authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body UpdateProfileReq true "Update profile request"
// @Success 200 {object} UserResp
// @Failure 400 {object} errors.HTTPError
// @Failure 401 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/me [put]
func (h handler) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	ip, sc, err := h.processUpdateProfileReq(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.UpdateProfile.processUpdateProfileReq: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	o, err := h.uc.UpdateProfile(ctx, sc, ip)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "user.delivery.http.UpdateProfile.UpdateProfile: %v", err)
		} else {
			h.l.Warnf(ctx, "user.delivery.http.UpdateProfile.UpdateProfile: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newUserResp(o))
}

// ChangePassword godoc
// @Summary Change password
// @Description Change the password of the currently authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ChangePasswordReq true "Change password request"
// @Success 200 {object} nil
// @Failure 400 {object} errors.HTTPError
// @Failure 401 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/me/change-password [post]
func (h handler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()

	ip, sc, err := h.processChangePasswordReq(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.ChangePassword.processChangePasswordReq: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	if err := h.uc.ChangePassword(ctx, sc, ip); err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "user.delivery.http.ChangePassword.ChangePassword: %v", err)
		} else {
			h.l.Warnf(ctx, "user.delivery.http.ChangePassword.ChangePassword: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, nil)
}

// GetDetail godoc
// @Summary Get user by ID (Admin only)
// @Description Get detailed information of a specific user by ID
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Success 200 {object} UserResp
// @Failure 401 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/{id} [get]
func (h handler) GetDetail(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processIDParam(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.GetDetail.processIDParam: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	o, err := h.uc.Detail(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "user.delivery.http.GetDetail.Detail: %v", err)
		} else {
			h.l.Warnf(ctx, "user.delivery.http.GetDetail.Detail: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newUserResp(o))
}

// List godoc
// @Summary List users (Admin only)
// @Description Get a list of all users without pagination
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param ids[] query []string false "User IDs to filter"
// @Success 200 {object} ListUserResp
// @Failure 401 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users [get]
func (h handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	input, sc, err := h.processListRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.List.processListRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	users, err := h.uc.List(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "user.delivery.http.List.List: %v", err)
		} else {
			h.l.Warnf(ctx, "user.delivery.http.List.List: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newListUserResp(users))
}

// Get godoc
// @Summary Get users with pagination (Admin only)
// @Description Get a paginated list of users
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param ids[] query []string false "User IDs to filter"
// @Success 200 {object} GetUserResp
// @Failure 401 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/page [get]
func (h handler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	input, sc, err := h.processGetRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.Get.processGetRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Get(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "user.delivery.http.Get.Get: %v", err)
		} else {
			h.l.Warnf(ctx, "user.delivery.http.Get.Get: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newGetUserResp(output))
}
