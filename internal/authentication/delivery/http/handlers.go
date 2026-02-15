package http

import (
	"identity-srv/pkg/response"

	"github.com/gin-gonic/gin"
)

// Logout
// @Summary Logout
// @Description Logout by expiring the authentication cookie. Requires authentication via cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} response.Resp "Success - Authentication cookie expired"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/logout [POST]
// @Security CookieAuth
func (h handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request
	sc, err := h.getScope(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	// 2. Call UseCase
	if err := h.uc.Logout(ctx, sc); err != nil {
		h.l.Errorf(ctx, "uc.Logout: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response
	h.expireAuthCookie(c)
	response.OK(c, nil)
}

// GetMe
// @Summary Get Current User
// @Description Get current authenticated user information.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} response.Resp{data=getMeResp} "Current user"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/me [GET]
// @Security CookieAuth
func (h handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request
	sc, err := h.getScope(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	// 2. Call UseCase
	user, err := h.uc.GetCurrentUser(ctx, sc)
	if err != nil {
		h.l.Errorf(ctx, "uc.GetCurrentUser: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response
	response.OK(c, h.newGetMeResp(user))
}
