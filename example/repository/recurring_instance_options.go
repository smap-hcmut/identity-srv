package repository

import (
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/paginator"
)

type CreateRecurringInstanceOptions struct {
	EventID       string
	Title         string
	BranchIDs     []string
	AssignIDs     []string
	DepartmentIDs []string
	TimezoneID    string
	StartTime     time.Time
	EndTime       time.Time
	AllDay        bool
	Repeat        models.EventRepeat
	RoomIDs       []string
	Description   string
	CategoryID    string
	RepeatUntil   *time.Time
	Notify        bool
	System        bool
	NotifyTime    *time.Time
	Alert         *models.DateConfig
	Public        bool
}

type UpdateRecurringInstanceOptions struct {
	ID            string
	Model         models.RecurringInstance
	EventID       string
	Title         string
	BranchIDs     []string
	AssignIDs     []string
	DepartmentIDs []string
	TimezoneID    string
	StartTime     time.Time
	EndTime       time.Time
	AllDay        bool
	RoomIDs       []string
	Description   string
	CategoryID    string
	Notify        bool
	NotifyTime    *time.Time
	Alert         *models.DateConfig
	Public        bool
}

type DeleteRecurringInstanceOptions struct {
	IDs     []string
	EventID string
}

type ListRecurringInstancesOptions struct {
	Filter
	EventID string
}

type GetOneRecurringInstanceOptions struct {
	Filter
	EventID string
}

type GetRecurringInstancesOptions struct {
	Filter
	PagQuery paginator.PaginatorQuery
	EventID  string
}

type PatchRecurringInstanceOptions struct {
	ID           string
	ExceptionDay *time.Time
	RepeatUntil  *time.Time
}

type DeleteNextRecurringInstancesOptions struct {
	EventID string
	Date    time.Time
}

type StartEndTime struct {
	StartTime time.Time
	EndTime   time.Time
}

type ListEventInstancesByEventIDsOptions struct {
	EventIDs   []string
	StartTime  time.Time
	EndTime    time.Time
	NotifyTime *time.Time
}

type CreateManyRecurringInstancesOptions struct {
	EventID            string
	RecurringInstances []CreateRecurringInstanceOptions
}

type UpdateAttendanceRecurringInstanceOptions struct {
	ID      string
	EventID string
	Status  int
}
