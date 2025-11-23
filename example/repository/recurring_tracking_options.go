package repository

import (
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
)

type CreateRecurringTrackingOptions struct {
	EventID      string
	Month        int32
	Year         int32
	Repeat       models.EventRepeat
	RepeatUntil  *time.Time
	StartEndTime []StartEndTime
}

type DeleteRecurringTrackingOptions struct {
	IDs     []string
	EventID string
	Month   *int32
	Year    *int32
}

type UpdateRepeatUntilRecurringTrackingsOptions struct {
	EventID     string
	RepeatUntil *time.Time
}
