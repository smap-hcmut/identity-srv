package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JWKS returns the JSON Web Key Set for JWT verification
// @Summary Get JSON Web Key Set
// @Description Returns public keys in JWKS format for JWT verification
// @Tags Authentication
// @Produce json
// @Success 200 {object} jwt.JWKS "JWKS with public keys"
// @Router /authentication/.well-known/jwks.json [get]
func (h *handler) JWKS(c *gin.Context) {
	jwks := h.jwtManager.GetJWKS()
	c.JSON(http.StatusOK, jwks)
}
