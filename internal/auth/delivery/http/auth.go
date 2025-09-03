package http

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/pkg/response"
)

// @Summary Register
// @Description Register
// @Tags Auth
// @Accept json
// @Produce json
// @Param registerReq body registerReq true "Register"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errEmailExisted(110004)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/auth/register [POST]
func (h handler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processRegisterRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "auth.http.Register.processRegisterRequest: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	_, err = h.uc.Register(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapErrorCode(err)
		if slices.Contains(NotFound, err) {
			h.l.Warnf(ctx, "auth.http.Register.Register.NotFound: %v", err)
		} else {
			h.l.Errorf(ctx, "auth.http.Register.Register: %v", err)
		}
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, nil)
}

// @Summary Send OTP
// @Description Send OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param sendOTPReq body sendOTPReq true "Send OTP"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errUserNotFound(110003), errWrongPassword(110005)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/auth/send-otp [POST]
func (h handler) SendOTP(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processSendOTPRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "auth.http.SendOTP.processSendOTPRequest: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	err = h.uc.SendOTP(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapErrorCode(err)
		if slices.Contains(NotFound, err) {
			h.l.Warnf(ctx, "auth.http.SendOTP.SendOTP.NotFound: %v", err)
		} else {
			h.l.Errorf(ctx, "auth.http.SendOTP.SendOTP: %v", err)
		}
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, nil)
}

// @Summary Verify OTP
// @Description Verify OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param verifyOTPReq body verifyOTPReq true "Verify OTP"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errOTPExpired(110006), errOTPNotMatch(110007)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/auth/verify-otp [POST]
func (h handler) VerifyOTP(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processVerifyOTPRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "auth.http.VerifyOTP.processVerifyOTPRequest: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	err = h.uc.VerifyOTP(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapErrorCode(err)
		if slices.Contains(NotFound, err) {
			h.l.Warnf(ctx, "auth.http.VerifyOTP.VerifyOTP.NotFound: %v", err)
		} else {
			h.l.Errorf(ctx, "auth.http.VerifyOTP.VerifyOTP: %v", err)
		}
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, nil)
}

// @Summary Login
// @Description Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginReq body loginReq true "Login"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errUserNotFound(110003), errWrongPassword(110005)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/auth/login [POST]
func (h handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processLoginRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "auth.http.Login.processLoginRequest: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	o, err := h.uc.Login(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapErrorCode(err)
		if slices.Contains(NotFound, err) {
			h.l.Warnf(ctx, "auth.http.Login.Login.NotFound: %v", err)
		} else {
			h.l.Errorf(ctx, "auth.http.Login.Login: %v", err)
		}
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, h.newLoginResp(o))
}

// @Summary Social Login
// @Description Social Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "Provider"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/auth/{provider} [POST]
func (h handler) SocialLogin(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processSocialLoginRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "auth.http.SocialLogin.processSocialLoginRequest: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	o, err := h.uc.SocialLogin(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapErrorCode(err)
		if slices.Contains(NotFound, err) {
			h.l.Warnf(ctx, "auth.http.SocialLogin.SocialLogin.NotFound: %v", err)
		} else {
			h.l.Errorf(ctx, "auth.http.SocialLogin.SocialLogin: %v", err)
		}
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, h.newSocialLoginResp(o))
}

// @Summary Social Callback
// @Description Social Callback
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "Provider"
// @Param code query string true "Code"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/auth/{provider}/callback [GET]
func (h handler) SocialCallback(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processSocialCallbackRequest(c)
	if err != nil {
		h.l.Warnf(ctx, "auth.http.SocialCallback.processSocialCallbackRequest: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	o, err := h.uc.SocialCallback(ctx, sc, req.toInput())
	if err != nil {
		mapErr := h.mapErrorCode(err)
		if slices.Contains(NotFound, err) {
			h.l.Warnf(ctx, "auth.http.SocialCallback.SocialCallback.NotFound: %v", err)
		} else {
			h.l.Errorf(ctx, "auth.http.SocialCallback.SocialCallback: %v", err)
		}
		response.Error(c, mapErr, h.d)
		return
	}

	response.OK(c, h.newSocialCallbackResp(o))
}
