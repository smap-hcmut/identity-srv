package event

import (
	"context"

	models "gitlab.com/gma-vietnam/tanca-connect/internal/models"
)

//go:generate mockery --name=UseCase
type UseCase interface {
	Create(ctx context.Context, sc models.Scope, input CreateInput) (CreateEventOutput, error)
	Detail(ctx context.Context, sc models.Scope, id string, eventID string) (DetailOutput, error)
	Calendar(ctx context.Context, sc models.Scope, input CalendarInput) (CalendarOutput, error)
	Update(ctx context.Context, sc models.Scope, input UpdateInput) error
	Delete(ctx context.Context, sc models.Scope, input DeleteInput) error
	GetOne(ctx context.Context, sc models.Scope, input GetOneInput) (EventInstance, error)
	UpdateAttendance(ctx context.Context, sc models.Scope, input UpdateAttendanceInput) error
	ListByIDs(ctx context.Context, sc models.Scope, ids []string) ([]models.Event, error)

	// Consumer UC
	CreateSystemEvent(ctx context.Context, sc models.Scope, input CreateSystemEventInput) error
	UpdateSystemEvent(ctx context.Context, sc models.Scope, input UpdateSystemEventInput) error

	// Job UC
	CheckNotifyEvent() error
}
