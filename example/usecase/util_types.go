package usecase

import (
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/internal/resources/notification"
)

type getEventNotiInput struct {
	EI           event.EventInstance
	UserIDs      []string
	Type         notification.SourceType
	AssigneeName string
	CreatedName  string
	TimeText     string
	DateText     string
	OldTimeText  string
	OldDateText  string
	EventTitle   string
	Duration     time.Duration
	DeleteType   models.EventAction
}
