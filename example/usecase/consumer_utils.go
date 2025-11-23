package usecase

import (
	"context"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
)

// parseEventTimes parses start and end times based on timezone parsing needs
func (uc implUseCase) parseEventTimes(ctx context.Context, startTimeStr, endTimeStr string, needParseTimezone bool) (startTime, endTime time.Time, err error) {
	if needParseTimezone {
		startTime, err = util.StrToDateTimeParse(startTimeStr)
		if err != nil {
			uc.l.Warnf(ctx, "event.delivery.rabbitmq.consumer.ParseTime: %v", err)
			return time.Time{}, time.Time{}, err
		}

		endTime, err = util.StrToDateTimeParse(endTimeStr)
		if err != nil {
			uc.l.Warnf(ctx, "event.delivery.rabbitmq.consumer.ParseTime: %v", err)
			return time.Time{}, time.Time{}, err
		}
	} else {
		startTime, err = util.StrToDateTime(startTimeStr)
		if err != nil {
			uc.l.Warnf(ctx, "event.delivery.rabbitmq.consumer.ParseTime: %v", err)
			return time.Time{}, time.Time{}, err
		}

		endTime, err = util.StrToDateTime(endTimeStr)
		if err != nil {
			uc.l.Warnf(ctx, "event.delivery.rabbitmq.consumer.ParseTime: %v", err)
			return time.Time{}, time.Time{}, err
		}
	}
	return startTime, endTime, nil
}
