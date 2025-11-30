package http

import (
	"slices"

	"smap-api/pkg/response"
	"smap-api/pkg/scope"

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
// @Description Login with email and password. Returns user information and sets authentication token as HttpOnly cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param loginReq body loginReq true "Login credentials"
// @Success 200 {object} response.Resp{data=loginResp} "Success - Authentication cookie set in Set-Cookie header (smap_auth_token)"
// @Header 200 {string} Set-Cookie "smap_auth_token=<JWT>; Path=/; Domain=.tantai.dev; HttpOnly; Secure; SameSite=Lax; Max-Age=7200"
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

	// Calculate Max-Age based on "Remember Me" flag
	maxAge := h.cookieConfig.MaxAge
	if req.Remember {
		maxAge = h.cookieConfig.MaxAgeRemember
	}

	// Set HttpOnly cookie with JWT token
	c.SetCookie(
		h.cookieConfig.Name,
		o.Token.AccessToken,
		maxAge,
		"/",
		h.cookieConfig.Domain,
		h.cookieConfig.Secure,
		true, // HttpOnly - always true for security
	)

	// Manually add SameSite attribute (Gin doesn't support it directly)
	h.addSameSiteAttribute(c, h.cookieConfig.SameSite)

	response.OK(c, h.newLoginResp(o))
}

// @Summary Logout
// @Description Logout by expiring the authentication cookie. Requires authentication via cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} response.Resp "Success - Authentication cookie expired"
// @Header 200 {string} Set-Cookie "smap_auth_token=; Path=/; Domain=.tantai.dev; HttpOnly; Secure; Max-Age=-1"
// @Failure 401 {object} response.Resp "Unauthorized - Missing or invalid authentication cookie"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/logout [POST]
// @Security CookieAuth
func (h handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// Get scope from context (set by Auth middleware)
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "authentication.http.Logout.GetScopeFromContext: scope not found")
		response.Unauthorized(c)
		return
	}

	// Call logout usecase (for any cleanup logic)
	err := h.uc.Logout(ctx, sc)
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.Logout.Logout: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// Expire authentication cookie by setting MaxAge to -1
	c.SetCookie(
		h.cookieConfig.Name,
		"",
		-1, // MaxAge: -1 expires the cookie immediately
		"/",
		h.cookieConfig.Domain,
		h.cookieConfig.Secure,
		true, // HttpOnly
	)

	response.OK(c, nil)
}

// @Summary Get Current User
// @Description Get current authenticated user information. Requires authentication via cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} response.Resp{data=getMeResp} "Success - Returns current user information"
// @Failure 401 {object} response.Resp "Unauthorized - Missing or invalid authentication cookie"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/me [GET]
// @Security CookieAuth
func (h handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	// Get scope from context (set by Auth middleware)
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "authentication.http.GetMe.GetScopeFromContext: scope not found")
		response.Unauthorized(c)
		return
	}

	// Call GetCurrentUser usecase
	o, err := h.uc.GetCurrentUser(ctx, sc)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "authentication.http.GetMe.GetCurrentUser: %v", err)
			response.Error(c, err, h.discord)
			return
		} else {
			h.l.Warnf(ctx, "authentication.http.GetMe.GetCurrentUser: %v", err)
			response.Error(c, err, h.discord)
			return
		}
	}

	response.OK(c, h.newGetMeResp(o))
}

// addSameSiteAttribute manually adds SameSite attribute to the last Set-Cookie header
// This is a workaround since Gin's SetCookie doesn't support SameSite parameter
func (h handler) addSameSiteAttribute(c *gin.Context, sameSite string) {
	cookies := c.Writer.Header()["Set-Cookie"]
	if len(cookies) > 0 {
		// Get the last cookie (the one we just set)
		lastCookie := cookies[len(cookies)-1]
		// Add SameSite attribute
		lastCookie += "; SameSite=" + sameSite
		// Update the header
		cookies[len(cookies)-1] = lastCookie
		c.Writer.Header()["Set-Cookie"] = cookies
	}
}
