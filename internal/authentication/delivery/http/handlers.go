package http

import (
	"net/http"

	"smap-api/pkg/response"

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

// JWKS returns the JSON Web Key Set for JWT verification
// @Summary Get JSON Web Key Set
// @Description Returns public keys in JWKS format for JWT verification
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]interface{} "JWKS with public keys"
// @Router /authentication/.well-known/jwks.json [get]
func (h handler) JWKS(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Call UseCase
	jwks, err := h.uc.GetJWKS(ctx)
	if err != nil {
		h.l.Errorf(ctx, "uc.GetJWKS: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// 2. Response
	c.JSON(http.StatusOK, jwks)
}
