package rabbitmq

import (
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/internal/resources/notification"
)

type CreateEventMsg struct {
	Title             string             `json:"title"`
	ShopID            string             `json:"shop_id"`
	CategoryID        string             `json:"category_id"`
	TimezoneID        string             `json:"timezone_id"`
	StartTime         string             `json:"start_time"`
	EndTime           string             `json:"end_time"`
	AllDay            bool               `json:"all_day"`
	Repeat            models.EventRepeat `json:"repeat"`
	AssignIDs         []string           `json:"assign_ids"`
	DepartmentIDs     []string           `json:"department_ids"`
	System            bool               `json:"system"`
	ObjectID          string             `json:"object_id"`
	NeedParseTimezone bool               `json:"need_parse_timezone"`
}

type ShopScope struct {
	ID     string `json:"id"`
	Suffix string `json:"suffix"`
}

type NotiData struct {
	Data     interface{}               `json:"data"`
	Activity notification.ActivityType `json:"activity"`
}

type MultiLangObj struct {
	Heading string `json:"heading"`
	Content string `json:"content"`
}

type PushNotiMsg struct {
	ShopScope     ShopScope               `json:"shop_scope"`
	Content       string                  `json:"content"`
	Heading       string                  `json:"heading"`
	UserIDs       []string                `json:"user_ids"`
	CreatedUserID string                  `json:"created_user_id"`
	En            MultiLangObj            `json:"en"`
	Ja            MultiLangObj            `json:"ja"`
	Data          NotiData                `json:"data"`
	Source        notification.SourceType `json:"source"`
}

type UpdateNotiTimeEventMsg struct {
	ShopID            string            `json:"shop_id"`
	RegularEventAlert models.DateConfig `json:"regular_event_alert"`
	AllDayEventAlert  models.DateConfig `json:"all_day_event_alert"`
}

type UpdateRequestEventIDMsg struct {
	ShopID    string `json:"shop_id"`
	EventID   string `json:"event_id"`
	RequestID string `json:"request_id"`
}

type DeleteEventMsg struct {
	ShopID  string `json:"shop_id"`
	EventID string `json:"event_id"`
}

type UpdateTaskEventIDMsg struct {
	ShopID  string `json:"shop_id"`
	EventID string `json:"event_id"`
	TaskID  string `json:"task_id"`
}

type UpdateSystemEventMsg struct {
	EventID           string             `json:"event_id"`
	Title             string             `json:"title"`
	ShopID            string             `json:"shop_id"`
	CategoryID        string             `json:"category_id"`
	TimezoneID        string             `json:"timezone_id"`
	StartTime         string             `json:"start_time"`
	EndTime           string             `json:"end_time"`
	AllDay            bool               `json:"all_day"`
	Repeat            models.EventRepeat `json:"repeat"`
	AssignIDs         []string           `json:"assign_ids"`
	DepartmentIDs     []string           `json:"department_ids"`
	System            bool               `json:"system"`
	ObjectID          string             `json:"object_id"`
	NeedParseTimezone bool               `json:"need_parse_timezone"`
}
