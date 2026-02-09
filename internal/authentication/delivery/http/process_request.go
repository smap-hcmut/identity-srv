package http

import (
	"smap-api/internal/model"

	"github.com/gin-gonic/gin"
)

func (h handler) processLoginRequest(c *gin.Context) (loginReq, model.Scope, error) {
	ctx := c.Request.Context()

	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "authentication.http.processLoginRequest.ShouldBindJSON: %v", err)
		return loginReq{}, model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Errorf(ctx, "authentication.http.processLoginRequest.validate: %v", err)
		return loginReq{}, model.Scope{}, errWrongBody
	}

	return req, model.Scope{}, nil
}
