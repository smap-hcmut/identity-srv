package http

import (
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
// @Success 200 {object} UserResponse
// @Failure 401 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/me [get]
func (h Handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()
	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		return
	}
	sc := scope.NewScope(payload)

	output, err := h.uc.DetailMe(ctx, sc)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.GetMe.DetailMe: %v", err)
		reportError(c, err)
		return
	}

	response.OK(c, toUserResponse(output.User))
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Update the profile information of the currently authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body UpdateProfileRequest true "Update profile request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} errors.HTTPError
// @Failure 401 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/me [put]
func (h Handler) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()
	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		return
	}
	sc := scope.NewScope(payload)

	input, err := processUpdateProfileRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.UpdateProfile.processUpdateProfileRequest: %v", err)
		reportError(c, err)
		return
	}

	output, err := h.uc.UpdateProfile(ctx, sc, input)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.UpdateProfile.UpdateProfile: %v", err)
		reportError(c, err)
		return
	}

	response.OK(c, toUserResponse(output.User))
}

// ChangePassword godoc
// @Summary Change password
// @Description Change the password of the currently authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} nil
// @Failure 400 {object} errors.HTTPError
// @Failure 401 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/me/change-password [post]
func (h Handler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()
	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		return
	}
	sc := scope.NewScope(payload)

	input, err := processChangePasswordRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.ChangePassword.processChangePasswordRequest: %v", err)
		reportError(c, err)
		return
	}

	if err := h.uc.ChangePassword(ctx, sc, input); err != nil {
		h.l.Errorf(ctx, "user.delivery.http.ChangePassword.ChangePassword: %v", err)
		reportError(c, err)
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
// @Success 200 {object} UserResponse
// @Failure 401 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/{id} [get]
func (h Handler) GetDetail(c *gin.Context) {
	ctx := c.Request.Context()
	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		return
	}
	sc := scope.NewScope(payload)

	id, err := processIDParam(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.GetDetail.processIDParam: %v", err)
		reportError(c, err)
		return
	}

	output, err := h.uc.Detail(ctx, sc, id)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.GetDetail.Detail: %v", err)
		reportError(c, err)
		return
	}

	response.OK(c, toUserResponse(output.User))
}

// List godoc
// @Summary List users (Admin only)
// @Description Get a list of all users without pagination
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param ids[] query []string false "User IDs to filter"
// @Success 200 {object} ListUserResponse
// @Failure 401 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users [get]
func (h Handler) List(c *gin.Context) {
	ctx := c.Request.Context()
	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		return
	}
	sc := scope.NewScope(payload)

	input, err := processListRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.List.processListRequest: %v", err)
		reportError(c, err)
		return
	}

	users, err := h.uc.List(ctx, sc, input)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.List.List: %v", err)
		reportError(c, err)
		return
	}

	response.OK(c, toListUserResponse(users))
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
// @Success 200 {object} GetUserResponse
// @Failure 401 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /users/page [get]
func (h Handler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	payload, ok := scope.GetPayloadFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		return
	}
	sc := scope.NewScope(payload)

	input, err := processGetRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.Get.processGetRequest: %v", err)
		reportError(c, err)
		return
	}

	output, err := h.uc.Get(ctx, sc, input)
	if err != nil {
		h.l.Errorf(ctx, "user.delivery.http.Get.Get: %v", err)
		reportError(c, err)
		return
	}

	response.OK(c, toGetUserResponse(output))
}
