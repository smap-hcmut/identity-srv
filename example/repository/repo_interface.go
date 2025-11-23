package repository

import (
	"context"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
)

//go:generate mockery --name=Repository
type Repository interface {
	EventRepository
	EventInstanceRepository
	RecurringTrackingRepository
}

type EventRepository interface {
	Create(ctx context.Context, sc models.Scope, opt CreateOptions) (models.Event, error)
	Detail(ctx context.Context, sc models.Scope, id string) (models.Event, error)
	Update(ctx context.Context, sc models.Scope, opt UpdateOptions) (models.Event, error)
	Delete(ctx context.Context, sc models.Scope, id string) error
	List(ctx context.Context, sc models.Scope, opt ListOptions) ([]models.Event, error)
	GetOne(ctx context.Context, sc models.Scope, opt GetOneOptions) (models.Event, error)
	ListByIDs(ctx context.Context, sc models.Scope, ids []string) ([]models.Event, error)
	UpdateAttendance(ctx context.Context, sc models.Scope, eventID string, status int) error
	UpdateRepeatUntil(ctx context.Context, sc models.Scope, id string, repeatUntil time.Time) (models.Event, error)
	SystemList(ctx context.Context, sc models.Scope, opt SystemListOptions) ([]models.Event, error)
}

type EventInstanceRepository interface {
	CreateRecurringInstance(ctx context.Context, sc models.Scope, opt CreateRecurringInstanceOptions) (models.RecurringInstance, error)
	CreateManyRecurringInstances(ctx context.Context, sc models.Scope, opt CreateManyRecurringInstancesOptions) ([]models.RecurringInstance, error)
	DetailRecurringInstance(ctx context.Context, sc models.Scope, id string, eventID string) (models.RecurringInstance, error)
	UpdateRecurringInstance(ctx context.Context, sc models.Scope, opt UpdateRecurringInstanceOptions) (models.RecurringInstance, error)
	DeleteRecurringInstance(ctx context.Context, sc models.Scope, opt DeleteRecurringInstanceOptions) error
	ListRecurringInstances(ctx context.Context, sc models.Scope, opt ListRecurringInstancesOptions) ([]models.RecurringInstance, error)
	GetOneRecurringInstance(ctx context.Context, sc models.Scope, opt GetOneRecurringInstanceOptions) (models.RecurringInstance, error)
	DeleteNextRecurringInstances(ctx context.Context, sc models.Scope, opt DeleteNextRecurringInstancesOptions) error
	UpdateAttendanceRecurringInstance(ctx context.Context, sc models.Scope, opt UpdateAttendanceRecurringInstanceOptions) error
	ListRecurringInstancesByEventIDs(ctx context.Context, sc models.Scope, opt ListEventInstancesByEventIDsOptions) ([]models.RecurringInstance, error)
}

type RecurringTrackingRepository interface {
	CreateRecurringTracking(ctx context.Context, sc models.Scope, opt CreateRecurringTrackingOptions) (models.RecurringTracking, error)
	GetGenRTsInDateRange(ctx context.Context, sc models.Scope, fromTime, toTime time.Time) ([]models.RecurringTracking, error)
	GetGenRTsNotInDateRange(ctx context.Context, sc models.Scope, fromTime, toTime time.Time) ([]models.RecurringTracking, error)
	UpdateRepeatUntilRecurringTrackings(ctx context.Context, sc models.Scope, opt UpdateRepeatUntilRecurringTrackingsOptions) error
	DeleteRecurringTracking(ctx context.Context, sc models.Scope, opt DeleteRecurringTrackingOptions) error
}
