package http

import (
	"strconv"
	"time"

	"smap-api/internal/audit/repository"
	pkgErrors "smap-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

var (
	errInvalidPage     = pkgErrors.NewHTTPError(30001, "Invalid page number")
	errInvalidLimit    = pkgErrors.NewHTTPError(30002, "Invalid limit")
	errInvalidFromDate = pkgErrors.NewHTTPError(30003, "Invalid from date format (use RFC3339)")
	errInvalidToDate   = pkgErrors.NewHTTPError(30004, "Invalid to date format (use RFC3339)")
)

func (h handler) processGetAuditLogsRequest(c *gin.Context) (repository.QueryOptions, error) {
	// Parse pagination
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return repository.QueryOptions{}, errInvalidPage
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return repository.QueryOptions{}, errInvalidLimit
	}
	if limit > 100 {
		limit = 100
	}

	// Parse date filters
	var from, to *time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		parsedFrom, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return repository.QueryOptions{}, errInvalidFromDate
		}
		from = &parsedFrom
	}

	if toStr := c.Query("to"); toStr != "" {
		parsedTo, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return repository.QueryOptions{}, errInvalidToDate
		}
		to = &parsedTo
	}

	return repository.QueryOptions{
		UserID: c.Query("user_id"),
		Action: c.Query("action"),
		From:   from,
		To:     to,
		Page:   page,
		Limit:  limit,
	}, nil
}
