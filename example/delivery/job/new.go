package job

import (
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/cron"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
)

type Handler interface {
	CheckNotifyEvent()
	Register() []cron.JobInfo
}
type handler struct {
	l    log.Logger
	uc   event.UseCase
	cron cron.Cron
}

// NewHandler returns a new instance of the HTTPHandler interface
func New(l log.Logger, cronJ cron.Cron, uc event.UseCase) Handler {
	h := handler{
		l:    l,
		cron: cronJ,
		uc:   uc,
	}
	return h
}
