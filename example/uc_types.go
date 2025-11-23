package event

import (
	"slices"
	"time"

	models "gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/paginator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateInput struct {
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
	Alert         *models.DateConfig
	ObjectID      string
	Public        bool
}

type Filter struct {
	IDs             []string
	ID              string
	StartTime       time.Time
	EndTime         time.Time
	NeedRepeat      *bool
	FilterAssignIDs []string
	NotifyTime      *time.Time
}

type CalendarInput struct {
	Filter
}

type CalendarOutput struct {
	EventInstances  []EventInstance
	EventCategories []models.EventCategory
	Users           []microservice.User
	Timezones       []models.Element
}

type UpdateInput struct {
	ID            string
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
	RepeatUntil   *time.Time
	RoomIDs       []string
	Description   string
	CategoryID    string
	Type          models.EventAction
	Notify        bool
	Alert         *models.DateConfig
	ObjectID      string
	Public        bool
}

type GetOneInput struct {
	Filter
	EventID string
}

type GetInput struct {
	Filter
	PagQuery paginator.PaginatorQuery
}

type DeleteInput struct {
	ID      string
	EventID string
	Type    models.EventAction
}

type DetailOutput struct {
	EventInstance
	Users         []microservice.User
	Departments   []microservice.Department
	Rooms         []models.Room
	Timezones     []models.Element
	EventCategory []models.EventCategory
	BaseEvent     EventInstance
	Branches      []microservice.Branch
}

type EventInstance struct {
	ID            primitive.ObjectID
	EventID       primitive.ObjectID
	ShopID        primitive.ObjectID
	Title         string
	BranchIDs     []primitive.ObjectID
	AssignIDs     []string
	DeclinedIDs   []string
	DepartmentIDs []primitive.ObjectID
	TimezoneID    primitive.ObjectID
	StartTime     time.Time
	EndTime       time.Time
	AllDay        bool
	Repeat        models.EventRepeat
	RoomIDs       []primitive.ObjectID
	CreatedByID   string
	RepeatUntil   *time.Time
	NotifyTime    *time.Time
	Description   string
	CategoryID    *primitive.ObjectID
	Attendance    int
	Notify        bool
	System        bool
	Alert         *models.DateConfig
	Public        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateSystemEventInput struct {
	Title             string
	AssignIDs         []string
	DepartmentIDs     []string
	TimezoneID        string
	StartTime         string
	EndTime           string
	AllDay            bool
	Repeat            models.EventRepeat
	RoomIDs           []string
	Description       string
	CategoryID        string
	RepeatUntil       *time.Time
	Notify            bool
	System            bool
	Alert             *models.DateConfig
	ObjectID          string
	NeedParseTimezone bool
}

type UpdateSystemEventInput struct {
	EventID           string
	Title             string
	AssignIDs         []string
	DepartmentIDs     []string
	TimezoneID        string
	StartTime         string
	EndTime           string
	AllDay            bool
	Repeat            models.EventRepeat
	RoomIDs           []string
	Description       string
	CategoryID        string
	RepeatUntil       *time.Time
	Notify            bool
	System            bool
	Alert             *models.DateConfig
	ObjectID          string
	NeedParseTimezone bool
}

// Change the receiver type to use the local alias
func EventToEventInstance(sc models.Scope, e models.Event) EventInstance {
	ei := EventInstance{
		ID:            e.ID,
		EventID:       e.ID,
		ShopID:        e.ShopID,
		Title:         e.Title,
		StartTime:     e.StartTime,
		EndTime:       e.EndTime,
		AllDay:        e.AllDay,
		Repeat:        e.Repeat,
		RoomIDs:       e.RoomIDs,
		BranchIDs:     e.BranchIDs,
		AssignIDs:     e.AssignIDs,
		DeclinedIDs:   e.DeclinedIDs,
		DepartmentIDs: e.DepartmentIDs,
		TimezoneID:    e.TimezoneID,
		Description:   e.Description,
		CategoryID:    e.CategoryID,
		CreatedByID:   e.CreatedByID,
		RepeatUntil:   e.RepeatUntil,
		NotifyTime:    e.NotifyTime,
		Notify:        e.Notify,
		System:        e.System,
		Alert:         e.Alert,
		Public:        e.Public,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	if slices.Contains(e.AcceptedIDs, sc.UserID) || e.CreatedByID == sc.UserID {
		ei.Attendance = 1
	} else if slices.Contains(e.DeclinedIDs, sc.UserID) {
		ei.Attendance = -1
	}

	return ei
}

func RecurringInstanceToEventInstance(sc models.Scope, e models.RecurringInstance) EventInstance {
	ei := EventInstance{
		ID:            e.ID,
		EventID:       e.EventID,
		ShopID:        e.ShopID,
		Title:         e.Title,
		StartTime:     e.StartTime,
		EndTime:       e.EndTime,
		AllDay:        e.AllDay,
		Repeat:        e.Repeat,
		RoomIDs:       e.RoomIDs,
		BranchIDs:     e.BranchIDs,
		AssignIDs:     e.AssignIDs,
		DeclinedIDs:   e.DeclinedIDs,
		DepartmentIDs: e.DepartmentIDs,
		TimezoneID:    e.TimezoneID,
		Description:   e.Description,
		CategoryID:    e.CategoryID,
		CreatedByID:   e.CreatedByID,
		RepeatUntil:   e.RepeatUntil,
		NotifyTime:    e.NotifyTime,
		Notify:        e.Notify,
		System:        e.System,
		Alert:         e.Alert,
		Public:        e.Public,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	if slices.Contains(e.AcceptedIDs, sc.UserID) || e.CreatedByID == sc.UserID {
		ei.Attendance = 1
	} else if slices.Contains(e.DeclinedIDs, sc.UserID) {
		ei.Attendance = -1
	}

	return ei
}

type UpdateAttendanceInput struct {
	ID      string
	EventID string
	Status  int
}

type PublishNotiEventInput struct {
	ID          string `json:"id"`
	EventID     string `json:"event_id"`
	CreatedByID string `json:"created_by_id,omitempty"`
	StartTime   string `json:"start_time"`
}

type CreateEventOutput struct {
	EventInstance
	UnavailableRooms []models.Room
}
