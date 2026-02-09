package http

import (
	"fmt"
	"strconv"
	"time"

	"smap-api/internal/audit/repository"
	"smap-api/pkg/discord"
	"smap-api/pkg/errors"
	pkgLog "smap-api/pkg/log"
	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type handler struct {
	l       pkgLog.Logger
	repo    repository.Repository
	discord *discord.Discord
}

func New(l pkgLog.Logger, repo repository.Repository, discord *discord.Discord) handler {
	return handler{
		l:       l,
		repo:    repo,
		discord: discord,
	}
}

// auditLogResp represents a single audit log entry in the response
type auditLogResp struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Action       string            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	IPAddress    string            `json:"ip_address"`
	UserAgent    string            `json:"user_agent"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
}

// auditLogsResp represents the paginated response
type auditLogsResp struct {
	Logs       []auditLogResp `json:"logs"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// @Summary Get Audit Logs
// @Description Query audit logs with pagination and filters. Requires ADMIN role.
// @Tags Audit
// @Accept json
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param action query string false "Filter by action (LOGIN, LOGOUT, etc.)"
// @Param from query string false "Filter from date (RFC3339 format)"
// @Param to query string false "Filter to date (RFC3339 format)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 50, max: 100)"
// @Success 200 {object} response.Resp{data=auditLogsResp} "Audit logs"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 403 {object} response.Resp "Forbidden - Requires ADMIN role"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /audit-logs [GET]
// @Security CookieAuth
func (h handler) GetAuditLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	userID := c.Query("user_id")
	action := c.Query("action")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	// Parse pagination
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		response.Error(c, errors.NewHTTPError(400, "Invalid page number"), h.discord)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		response.Error(c, errors.NewHTTPError(400, "Invalid limit"), h.discord)
		return
	}

	// Enforce max limit
	if limit > 100 {
		limit = 100
	}

	// Parse date filters
	var from, to *time.Time
	if fromStr != "" {
		parsedFrom, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			response.Error(c, errors.NewHTTPError(400, "Invalid from date format (use RFC3339)"), h.discord)
			return
		}
		from = &parsedFrom
	}

	if toStr != "" {
		parsedTo, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			response.Error(c, errors.NewHTTPError(400, "Invalid to date format (use RFC3339)"), h.discord)
			return
		}
		to = &parsedTo
	}

	// Query audit logs
	logs, totalCount, err := h.repo.Query(ctx, repository.QueryOptions{
		UserID: userID,
		Action: action,
		From:   from,
		To:     to,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		h.l.Errorf(ctx, "audit.http.GetAuditLogs.Query: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// Convert to response format
	logsResp := make([]auditLogResp, len(logs))
	for i, log := range logs {
		// Helper to safely dereference string pointers
		getString := func(s *string) string {
			if s != nil {
				return *s
			}
			return ""
		}

		// Convert metadata from map[string]interface{} to map[string]string
		metadata := make(map[string]string)
		for k, v := range log.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			} else {
				metadata[k] = fmt.Sprint(v)
			}
		}

		logsResp[i] = auditLogResp{
			ID:           log.ID,
			UserID:       getString(log.UserID),
			Action:       log.Action,
			ResourceType: getString(log.ResourceType),
			ResourceID:   getString(log.ResourceID),
			Metadata:     metadata,
			IPAddress:    getString(log.IPAddress),
			UserAgent:    getString(log.UserAgent),
			CreatedAt:    log.CreatedAt,
			ExpiresAt:    log.ExpiresAt,
		}
	}

	response.OK(c, auditLogsResp{
		Logs:       logsResp,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	})
}
