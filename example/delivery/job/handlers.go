package job

import "context"

func (h handler) CheckNotifyEvent() {
	ctx := context.Background()

	h.l.Infof(ctx, "event.delivery.job.CheckNotifyEvent: Start schedule CheckNotifyEvent")

	err := h.uc.CheckNotifyEvent()
	if err != nil {
		h.l.Errorf(ctx, "event.delivery.job.CheckNotifyEvent: %v", err)
		return
	}

	h.l.Infof(ctx, "event.delivery.job.CheckNotifyEvent: End schedule CheckNotifyEvent")
}
