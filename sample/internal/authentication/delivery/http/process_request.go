package http

import (
	"smap-api/internal/model"

	"github.com/gin-gonic/gin"
)

func (h handler) processRegisterRequest(c *gin.Context) (registerReq, model.Scope, error) {
	ctx := c.Request.Context()

	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "authentication.http.processRegisterRequest.ShouldBindJSON: %v", err)
		return registerReq{}, model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "authentication.http.processRegisterRequest.validate: %v", err)
		return registerReq{}, model.Scope{}, errWrongBody
	}

	return req, model.Scope{}, nil
}

func (h handler) processSendOTPRequest(c *gin.Context) (sendOTPReq, model.Scope, error) {
	ctx := c.Request.Context()

	var req sendOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "authentication.http.processSendOTPRequest.ShouldBindJSON: %v", err)
		return sendOTPReq{}, model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "authentication.http.processSendOTPRequest.validate: %v", err)
		return sendOTPReq{}, model.Scope{}, errWrongBody
	}

	return req, model.Scope{}, nil
}

func (h handler) processVerifyOTPRequest(c *gin.Context) (verifyOTPReq, model.Scope, error) {
	ctx := c.Request.Context()

	var req verifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "authentication.http.processVerifyOTPRequest.ShouldBindJSON: %v", err)
		return verifyOTPReq{}, model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "authentication.http.processVerifyOTPRequest.validate: %v", err)
		return verifyOTPReq{}, model.Scope{}, errWrongBody
	}

	return req, model.Scope{}, nil
}

func (h handler) processLoginRequest(c *gin.Context) (loginReq, model.Scope, error) {
	ctx := c.Request.Context()

	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "authentication.http.processLoginRequest.ShouldBindJSON: %v", err)
		return loginReq{}, model.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "authentication.http.processLoginRequest.validate: %v", err)
		return loginReq{}, model.Scope{}, errWrongBody
	}

	return req, model.Scope{}, nil
}
