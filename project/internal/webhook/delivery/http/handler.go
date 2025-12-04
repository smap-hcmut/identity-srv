package http

import (
	"smap-project/pkg/response"

	"github.com/gin-gonic/gin"
)

// DryRunCallback handles dry-run webhook callbacks from collector
func (h handler) DryRunCallback(c *gin.Context) {
	ctx := c.Request.Context()

	req, err := h.processCallbackReq(c)
	if err != nil {
		h.l.Errorf(ctx, "webhook.http.DryRunCallback.processCallbackReq: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// Handle callback
	if err := h.uc.HandleDryRunCallback(ctx, req); err != nil {
		err = h.mapErrorCode(err)
		h.l.Errorf(ctx, "webhook.http.DryRunCallback.HandleDryRunCallback: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	h.l.Infof(ctx, "Webhook callback processed successfully: job_id=%s, platform=%s, status=%s",
		req.JobID, req.Platform, req.Status)

	response.OK(c, nil)
}
