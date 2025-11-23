package repository

import (
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/paginator"
)

type CreateOptions struct {
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
	ObjectID      string
	Public        bool
}

type UpdateOptions struct {
	ID            string
	Model         models.Event
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
	Alert         *models.DateConfig
	NotifyTime    *time.Time
	ObjectID      string
	Public        bool
}

type UpdateRepeatUntilOptions struct {
	ID          string
	RepeatUntil *time.Time
}

type Filter struct {
	IDs                []string
	ID                 string
	StartTime          time.Time
	EndTime            time.Time
	NeedRepeat         *bool
	BranchIDs          []string
	DepartmentIDs      []string
	ExcludeCategoryIDs []string
}

type ListOptions struct {
	Filter
}

type GetOneOptions struct {
	Filter
}

type GetOptions struct {
	Filter
	PagQuery paginator.PaginatorQuery
}

type SystemListOptions struct {
	IDs           []string
	ID            string
	StartTime     time.Time
	EndTime       time.Time
	NeedRepeat    *bool
	FilterUserIDs []string
	NotifyTime    *time.Time
}
