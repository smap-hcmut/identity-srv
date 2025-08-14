package response

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	pkgErrors "github.com/nguyentantai21042004/smap-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Resp is the response format.
type Resp struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Errors    any    `json:"errors,omitempty"`
}

// NewOKResp returns a new OK response with the given data.
func NewOKResp(data any) Resp {
	return Resp{
		ErrorCode: 0,
		Message:   "Success",
		Data:      data,
	}
}

// Ok returns a new OK response with the given data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, NewOKResp(data))
}

// Unauthorized returns a new Unauthorized response with the given data.
func Unauthorized(c *gin.Context) {
	c.JSON(parseError(pkgErrors.NewUnauthorizedHTTPError(), c, nil))
}

func Forbidden(c *gin.Context) {
	c.JSON(parseError(pkgErrors.NewForbiddenHTTPError(), c, nil))
}

func parseError(err error, c *gin.Context, d *discord.Discord) (int, Resp) {
	switch parsedErr := err.(type) {
	case *pkgErrors.ValidationError:
		return http.StatusBadRequest, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Error(),
		}
	case *pkgErrors.PermissionError:
		return http.StatusBadRequest, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Error(),
		}
	case *pkgErrors.ValidationErrorCollector:
		return http.StatusBadRequest, Resp{
			ErrorCode: ValidationErrorCode,
			Message:   ValidationErrorMsg,
			Errors:    parsedErr.Errors(),
		}
	case *pkgErrors.PermissionErrorCollector:
		return http.StatusBadRequest, Resp{
			ErrorCode: PermissionErrorCode,
			Message:   PermissionErrorMsg,
			Errors:    parsedErr.Errors(),
		}
	case *pkgErrors.HTTPError:
		statusCode := parsedErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}

		return statusCode, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Message,
		}
	default:
		if d != nil {
			stackTrace := captureStackTrace()
			sendDiscordMesssageAsync(c, d, buildInternalServerErrorDataForReportBug(c, err.Error(), stackTrace))
		}

		return http.StatusInternalServerError, Resp{
			ErrorCode: 500,
			Message:   DefaultErrorMessage,
		}
	}
}

// Error returns a new Error response with the given error.
func Error(c *gin.Context, err error, d *discord.Discord) {
	c.JSON(parseError(err, c, d))
}

// HttpError returns a new Error response with the given error.
func HttpError(c *gin.Context, err *pkgErrors.HTTPError) {
	c.JSON(parseError(err, c, nil))
}

// ErrorMapping is a map of error to HTTPError.
type ErrorMapping map[error]*pkgErrors.HTTPError

// ErrorWithMap returns a new Error response with the given error.
func ErrorWithMap(c *gin.Context, err error, eMap ErrorMapping) {
	if httpErr, ok := eMap[err]; ok {
		Error(c, httpErr, nil)
		return
	}

	Error(c, err, nil)
}

func PanicError(c *gin.Context, err any, d *discord.Discord) {
	if err == nil {
		c.JSON(parseError(nil, c, nil))
	} else {
		c.JSON(parseError(err.(error), c, nil))
	}
}

func captureStackTrace() []string {
	var pcs [defaultStackTraceDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return nil
	}

	var stackTrace []string
	for _, pc := range pcs[:n] {
		f := runtime.FuncForPC(pc)
		if f != nil {
			file, line := f.FileLine(pc)
			stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", file, line, f.Name()))
		}
	}

	return stackTrace
}
