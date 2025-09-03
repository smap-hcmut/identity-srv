package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/models"
)

func (h handler) processRegisterRequest(c *gin.Context) (registerReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "auth.http.processRegisterRequest.ShouldBindJSON: %v", err)
		return registerReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "auth.http.processRegisterRequest.validate: %v", err)
		return registerReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}

func (h handler) processSendOTPRequest(c *gin.Context) (sendOTPReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req sendOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "auth.http.processSendOTPRequest.ShouldBindJSON: %v", err)
		return sendOTPReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "auth.http.processSendOTPRequest.validate: %v", err)
		return sendOTPReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}

func (h handler) processVerifyOTPRequest(c *gin.Context) (verifyOTPReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req verifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "auth.http.processVerifyOTPRequest.ShouldBindJSON: %v", err)
		return verifyOTPReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "auth.http.processVerifyOTPRequest.validate: %v", err)
		return verifyOTPReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}

func (h handler) processLoginRequest(c *gin.Context) (loginReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Warnf(ctx, "auth.http.processLoginRequest.ShouldBindJSON: %v", err)
		return loginReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "auth.http.processLoginRequest.validate: %v", err)
		return loginReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}

func (h handler) processSocialLoginRequest(c *gin.Context) (socialLoginReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req socialLoginReq
	if err := c.ShouldBindUri(&req); err != nil {
		h.l.Warnf(ctx, "auth.http.processSocialLoginRequest.ShouldBindUri: %v", err)
		return socialLoginReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "auth.http.processSocialLoginRequest.validate: %v", err)
		return socialLoginReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}

func (h handler) processSocialCallbackRequest(c *gin.Context) (socialCallbackReq, models.Scope, error) {
	ctx := c.Request.Context()

	var req socialCallbackReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.l.Warnf(ctx, "auth.http.processSocialCallbackRequest.ShouldBindQuery: %v", err)
		return socialCallbackReq{}, models.Scope{}, errWrongBody
	}

	if err := c.ShouldBindUri(&req); err != nil {
		h.l.Warnf(ctx, "auth.http.processSocialCallbackRequest.ShouldBindUri: %v", err)
		return socialCallbackReq{}, models.Scope{}, errWrongBody
	}

	if err := req.validate(); err != nil {
		h.l.Warnf(ctx, "auth.http.processSocialCallbackRequest.validate: %v", err)
		return socialCallbackReq{}, models.Scope{}, errWrongBody
	}

	return req, models.Scope{}, nil
}
