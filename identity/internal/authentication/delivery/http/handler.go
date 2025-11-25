package http

import (
	"slices"

	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// @Summary Register
// @Description Register
// @Tags Authentication
// @Accept json
// @Produce json
// @Param registerReq body registerReq true "Register"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errEmailExisted(110004)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/register [POST]
func (h handler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processRegisterRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.Register.processRegisterRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	_, err = h.uc.Register(ctx, sc, req.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "authentication.http.Register.Register: %v", err)
			response.Error(c, err, h.discord)
			return
		} else {
			h.l.Warnf(ctx, "authentication.http.Register.Register: %v", err)
			response.Error(c, err, h.discord)
			return
		}
	}

	response.OK(c, nil)
}

// @Summary Send OTP
// @Description Send OTP
// @Tags Authentication
// @Accept json
// @Produce json
// @Param sendOTPReq body sendOTPReq true "Send OTP"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errUserNotFound(110003), errWrongPassword(110005)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/send-otp [POST]
func (h handler) SendOTP(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processSendOTPRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.SendOTP.processSendOTPRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	err = h.uc.SendOTP(ctx, sc, req.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "authentication.http.SendOTP.SendOTP: %v", err)
			response.Error(c, err, h.discord)
			return
		} else {
			h.l.Warnf(ctx, "authentication.http.SendOTP.SendOTP: %v", err)
			response.Error(c, err, h.discord)
			return
		}
	}

	response.OK(c, nil)
}

// @Summary Verify OTP
// @Description Verify OTP
// @Tags Authentication
// @Accept json
// @Produce json
// @Param verifyOTPReq body verifyOTPReq true "Verify OTP"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errOTPExpired(110006), errOTPNotMatch(110007)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/verify-otp [POST]
func (h handler) VerifyOTP(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processVerifyOTPRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.VerifyOTP.processVerifyOTPRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	err = h.uc.VerifyOTP(ctx, sc, req.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "authentication.http.VerifyOTP.VerifyOTP: %v", err)
			response.Error(c, err, h.discord)
			return
		} else {
			h.l.Warnf(ctx, "authentication.http.VerifyOTP.VerifyOTP: %v", err)
			response.Error(c, err, h.discord)
			return
		}
	}

	response.OK(c, nil)
}

// @Summary Login
// @Description Login
// @Tags Authentication
// @Accept json
// @Produce json
// @Param loginReq body loginReq true "Login"
// @Success 200 {object} response.Resp "Success"
// @Failure 400 {object} response.Resp "Bad Request, Error errWrongBody(110002), errUserNotFound(110003), errWrongPassword(110005)"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/login [POST]
func (h handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processLoginRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.Login.processLoginRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	o, err := h.uc.Login(ctx, sc, req.toInput())
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "authentication.http.Login.Login: %v", err)
			response.Error(c, err, h.discord)
			return
		} else {
			h.l.Warnf(ctx, "authentication.http.Login.Login: %v", err)
			response.Error(c, err, h.discord)
			return
		}
	}

	response.OK(c, h.newLoginResp(o))
}
