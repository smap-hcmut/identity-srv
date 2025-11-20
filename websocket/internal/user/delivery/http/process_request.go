package http

import (
	"smap-api/internal/user"
	"smap-api/pkg/errors"
	"smap-api/pkg/paginator"
	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func processListRequest(c *gin.Context) (user.ListInput, error) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return user.ListInput{}, &errors.HTTPError{
			Code:    ErrCodeWrongBody,
			Message: err.Error(),
		}
	}

	return user.ListInput{
		Filter: user.Filter{
			IDs: req.IDs,
		},
	}, nil
}

func processGetRequest(c *gin.Context) (user.GetInput, error) {
	var req GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return user.GetInput{}, &errors.HTTPError{
			Code:    ErrCodeWrongBody,
			Message: err.Error(),
		}
	}

	return user.GetInput{
		Filter: user.Filter{
			IDs: req.IDs,
		},
		PaginateQuery: paginator.PaginateQuery{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}, nil
}

func processUpdateProfileRequest(c *gin.Context) (user.UpdateProfileInput, error) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return user.UpdateProfileInput{}, &errors.HTTPError{
			Code:    ErrCodeWrongBody,
			Message: err.Error(),
		}
	}

	return user.UpdateProfileInput{
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
	}, nil
}

func processChangePasswordRequest(c *gin.Context) (user.ChangePasswordInput, error) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return user.ChangePasswordInput{}, &errors.HTTPError{
			Code:    ErrCodeWrongBody,
			Message: err.Error(),
		}
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return user.ChangePasswordInput{}, &errors.HTTPError{
			Code:    ErrCodeFieldRequired,
			Message: "Old password and new password are required",
		}
	}

	return user.ChangePasswordInput{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}, nil
}

func processIDParam(c *gin.Context) (string, error) {
	id := c.Param("id")
	if id == "" {
		return "", &errors.HTTPError{
			Code:    ErrCodeInvalidID,
			Message: "Invalid ID",
		}
	}
	return id, nil
}

func reportError(c *gin.Context, err error) {
	httpErr := toHTTPError(err)
	if httpErr != nil {
		response.Error(c, *httpErr, nil)
		return
	}

	response.Error(c, err, nil)
}
