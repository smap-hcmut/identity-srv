package http

import (
	"identity-srv/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetAuditLogs queries audit logs with pagination and filters
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
// @Failure 403 {object} response.Resp "Forbidden"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /audit-logs [GET]
// @Security CookieAuth
func (h handler) GetAuditLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request
	opts, err := h.processGetAuditLogsRequest(c)
	if err != nil {
		response.Error(c, err, h.discord)
		return
	}

	// 2. Call Repository
	logs, totalCount, err := h.repo.Query(ctx, opts)
	if err != nil {
		h.l.Errorf(ctx, "audit.http.GetAuditLogs: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// 3. Response
	response.OK(c, h.newAuditLogsResp(logs, totalCount, opts.Page, opts.Limit))
}
