package job

import (
	"gitlab.com/gma-vietnam/tanca-connect/pkg/cron"
)

func (h handler) Register() []cron.JobInfo {
	return []cron.JobInfo{
		{CronTime: "* * * * *", Handler: h.CheckNotifyEvent}, // Chạy mỗi 1 phút
	}
}
