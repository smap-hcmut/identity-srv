package http

import (
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/response"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type createReq struct {
	Title         string             `json:"title" binding:"required"`
	BranchIDs     []string           `json:"branch_ids"`
	AssignIDs     []string           `json:"assign_ids"`
	DepartmentIDs []string           `json:"department_ids"`
	TimezoneID    string             `json:"timezone_id" binding:"required"`
	StartTime     string             `json:"start_time" binding:"required"`
	EndTime       string             `json:"end_time" binding:"required"`
	AllDay        bool               `json:"all_day"`
	Repeat        models.EventRepeat `json:"repeat" binding:"required"`
	RoomIDs       []string           `json:"room_ids"`
	Description   string             `json:"description"`
	CategoryID    string             `json:"category_id"`
	RepeatUntil   string             `json:"repeat_until"`
	Notify        bool               `json:"notify"`
	Alert         *dateConfig        `json:"alert"`
	Public        bool               `json:"public"`
}

type dateConfig struct {
	Instant bool            `json:"instant"`
	Unit    models.DateUnit `json:"unit"`
	Num     int             `json:"num"`
	Hour    int             `json:"hour"`
}

func (dateConfig dateConfig) validateDateConfig(isAllDay bool) error {
	if dateConfig.Instant {
		return nil
	}

	switch dateConfig.Unit {
	case models.DateUnitMinute, models.DateUnitHour, models.DateUnitDay, models.DateUnitWeek:
	default:
		return errWrongBody
	}

	if isAllDay {
		if dateConfig.Hour < 0 || dateConfig.Hour > 23 {
			return errWrongBody
		}

		if dateConfig.Num < 0 {
			return errWrongBody
		}

		return nil
	}

	if dateConfig.Unit != "" || dateConfig.Num > 0 {
		if dateConfig.Unit == "" || dateConfig.Num <= 0 {
			return errWrongBody
		}
	}

	return nil
}

func (r createReq) validate() error {
	_, err := util.StrToDateTime(r.StartTime)
	if err != nil {
		return errWrongBody
	}

	_, err = util.StrToDateTime(r.EndTime)
	if err != nil {
		return errWrongBody
	}

	if r.CategoryID != "" {
		_, err = primitive.ObjectIDFromHex(r.CategoryID)
		if err != nil {
			return errWrongBody
		}
	}

	if r.RepeatUntil != "" {
		repeatUntil, err := util.StrToDateTime(r.RepeatUntil)
		if err != nil {
			return errWrongBody
		}

		if repeatUntil.Before(time.Now()) {
			return errWrongBody
		}
	}

	if r.Notify && r.Alert != nil {
		if err := r.Alert.validateDateConfig(r.AllDay); err != nil {
			return err
		}
	}

	if len(r.BranchIDs) == 0 && (len(r.AssignIDs) > 0 || len(r.DepartmentIDs) > 0) {
		return errWrongBody
	}

	return nil
}

func (r createReq) toInput() event.CreateInput {
	startTime, _ := util.StrToDateTime(r.StartTime)
	endTime, _ := util.StrToDateTime(r.EndTime)

	input := event.CreateInput{
		Title:         r.Title,
		BranchIDs:     r.BranchIDs,
		AssignIDs:     r.AssignIDs,
		DepartmentIDs: r.DepartmentIDs,
		TimezoneID:    r.TimezoneID,
		StartTime:     startTime,
		EndTime:       endTime,
		AllDay:        r.AllDay,
		Repeat:        r.Repeat,
		RoomIDs:       r.RoomIDs,
		Description:   r.Description,
		CategoryID:    r.CategoryID,
		Notify:        r.Notify,
		Public:        r.Public,
	}

	if r.RepeatUntil != "" {
		repeatUntil, _ := util.StrToDateTime(r.RepeatUntil)
		input.RepeatUntil = &repeatUntil
	}

	if r.Notify && r.Alert != nil {
		input.Alert = &models.DateConfig{
			Instant: r.Alert.Instant,
			Unit:    r.Alert.Unit,
			Num:     r.Alert.Num,
			Hour:    r.Alert.Hour,
		}
	}

	return input
}

type updateReq struct {
	ID            string             `uri:"id"`
	EventID       string             `uri:"event_id"`
	Title         string             `json:"title" binding:"required"`
	BranchIDs     []string           `json:"branch_ids"`
	AssignIDs     []string           `json:"assign_ids"`
	DepartmentIDs []string           `json:"department_ids"`
	TimezoneID    string             `json:"timezone_id" binding:"required"`
	StartTime     string             `json:"start_time" binding:"required"`
	EndTime       string             `json:"end_time" binding:"required"`
	AllDay        bool               `json:"all_day"`
	Repeat        models.EventRepeat `json:"repeat" binding:"required"`
	RoomIDs       []string           `json:"room_ids"`
	Type          models.EventAction `json:"type"`
	CategoryID    string             `json:"category_id"`
	Description   string             `json:"description"`
	Notify        bool               `json:"notify"`
	RepeatUntil   string             `json:"repeat_until"`
	Alert         *dateConfig        `json:"alert"`
	Public        bool               `json:"public"`
}

func (r updateReq) validate() error {
	if r.ID != r.EventID && r.Type == "" {
		return errWrongBody
	}

	if r.ID == r.EventID && r.Type != "" {
		return errWrongBody
	}

	if r.CategoryID != "" {
		_, err := primitive.ObjectIDFromHex(r.CategoryID)
		if err != nil {
			return errWrongBody
		}
	}

	startTime, err := util.StrToDateTime(r.StartTime)
	if err != nil {
		return errWrongBody
	}

	endTime, err := util.StrToDateTime(r.EndTime)
	if err != nil {
		return errWrongBody
	}

	if startTime.After(endTime) {
		return errWrongBody
	}

	if r.RepeatUntil != "" {
		_, err = util.StrToDateTime(r.RepeatUntil)
		if err != nil {
			return errWrongBody
		}
	}

	if r.Notify && r.Alert != nil {
		if err := r.Alert.validateDateConfig(r.AllDay); err != nil {
			return err
		}
	}

	if len(r.BranchIDs) == 0 && (len(r.AssignIDs) > 0 || len(r.DepartmentIDs) > 0) {
		return errWrongBody
	}

	return nil
}

func (r updateReq) toInput() event.UpdateInput {
	ip := event.UpdateInput{
		ID:            r.ID,
		EventID:       r.EventID,
		Title:         r.Title,
		BranchIDs:     r.BranchIDs,
		AssignIDs:     r.AssignIDs,
		DepartmentIDs: r.DepartmentIDs,
		TimezoneID:    r.TimezoneID,
		AllDay:        r.AllDay,
		Repeat:        r.Repeat,
		RoomIDs:       r.RoomIDs,
		Type:          r.Type,
		CategoryID:    r.CategoryID,
		Description:   r.Description,
		Notify:        r.Notify,
		Public:        r.Public,
	}

	startTime, _ := util.StrToDateTime(r.StartTime)
	endTime, _ := util.StrToDateTime(r.EndTime)
	ip.StartTime = startTime
	ip.EndTime = endTime

	if r.RepeatUntil != "" {
		repeatUntil, _ := util.StrToDateTime(r.RepeatUntil)
		ip.RepeatUntil = &repeatUntil
	}

	if r.Notify && r.Alert != nil {
		ip.Alert = &models.DateConfig{
			Instant: r.Alert.Instant,
			Unit:    r.Alert.Unit,
			Num:     r.Alert.Num,
			Hour:    r.Alert.Hour,
		}
	}

	return ip
}

type deleteReq struct {
	ID      string             `uri:"id"`
	EventID string             `uri:"event_id"`
	Type    models.EventAction `form:"type"`
}

func (r deleteReq) validate() error {
	if r.ID != r.EventID && r.Type == "" {
		return errWrongQuery
	}

	if r.ID == r.EventID && r.Type != "" {
		return errWrongQuery
	}

	return nil
}

func (r deleteReq) toInput() event.DeleteInput {
	return event.DeleteInput{
		ID:      r.ID,
		EventID: r.EventID,
		Type:    r.Type,
	}
}

type detailResp struct {
	ID            string             `json:"id"`
	EventID       string             `json:"event_id"`
	Title         string             `json:"title,omitempty"`
	AssignIDs     []string           `json:"assign_ids,omitempty"`
	DepartmentIDs []string           `json:"department_ids,omitempty"`
	TimezoneID    string             `json:"timezone_id"`
	StartTime     response.DateTime  `json:"start_time"`
	EndTime       response.DateTime  `json:"end_time"`
	AllDay        bool               `json:"all_day,omitempty"`
	Repeat        models.EventRepeat `json:"repeat,omitempty"`
	RoomIDs       []string           `json:"room_ids,omitempty"`
	Description   string             `json:"description,omitempty"`
	Attendance    int                `json:"attendance"`
	EventCategory *respObj           `json:"category,omitempty"`
	Timezone      respObj            `json:"timezone"`
	Assigns       []respObj          `json:"assigns,omitempty"`
	Depts         []respObj          `json:"departments,omitempty"`
	Rooms         []respObj          `json:"rooms,omitempty"`
	CreatedBy     respObj            `json:"created_by,omitempty"`
	BaseEvent     *baseEventResp     `json:"base_event,omitempty"`
	Notify        bool               `json:"notify"`
	System        bool               `json:"system"`
	RepeatUntil   *response.DateTime `json:"repeat_until,omitempty"`
	Alert         *dateConfig        `json:"alert,omitempty"`
	Branches      []respObj          `json:"branches,omitempty"`
	Public        bool               `json:"public"`
}

func (h handler) newDetailResp(e event.DetailOutput) detailResp {
	mapUser := map[string]microservice.User{}
	for _, user := range e.Users {
		mapUser[user.ID] = user
	}
	mapCategory := map[string]models.EventCategory{}
	for _, category := range e.EventCategory {
		mapCategory[category.ID.Hex()] = category
	}
	mapTimezone := map[string]models.Element{}
	for _, timezone := range e.Timezones {
		mapTimezone[timezone.ID.Hex()] = timezone
	}
	mapDept := map[string]microservice.Department{}
	for _, dept := range e.Departments {
		mapDept[dept.ID] = dept
	}
	mapRoom := map[string]models.Room{}
	for _, room := range e.Rooms {
		mapRoom[room.ID.Hex()] = room
	}

	out := detailResp{
		ID:            e.EventInstance.ID.Hex(),
		EventID:       e.EventInstance.EventID.Hex(),
		Title:         e.EventInstance.Title,
		AssignIDs:     e.EventInstance.AssignIDs,
		DepartmentIDs: mongo.HexFromObjectIDsOrNil(e.EventInstance.DepartmentIDs),
		TimezoneID:    e.EventInstance.TimezoneID.Hex(),
		StartTime:     response.DateTime(e.EventInstance.StartTime),
		EndTime:       response.DateTime(e.EventInstance.EndTime),
		AllDay:        e.EventInstance.AllDay,
		Repeat:        e.EventInstance.Repeat,
		RoomIDs:       mongo.HexFromObjectIDsOrNil(e.EventInstance.RoomIDs),
		Description:   e.EventInstance.Description,
		Attendance:    e.EventInstance.Attendance,
		Notify:        e.EventInstance.Notify,
		System:        e.EventInstance.System,
		Public:        e.EventInstance.Public,
	}

	if e.EventInstance.CategoryID != nil {
		if _, ok := mapCategory[e.EventInstance.CategoryID.Hex()]; ok {
			out.EventCategory = &respObj{
				ID:    e.EventInstance.CategoryID.Hex(),
				Name:  mapCategory[e.EventInstance.CategoryID.Hex()].Name,
				Color: mapCategory[e.EventInstance.CategoryID.Hex()].Color,
				Key:   mapCategory[e.EventInstance.CategoryID.Hex()].Key,
			}
		}
	}

	if _, ok := mapTimezone[e.EventInstance.TimezoneID.Hex()]; ok {
		out.Timezone = respObj{
			ID:   e.EventInstance.TimezoneID.Hex(),
			Name: mapTimezone[e.EventInstance.TimezoneID.Hex()].Name,
		}
	}

	if len(e.EventInstance.AssignIDs) > 0 {
		assigns := make([]respObj, 0, len(e.EventInstance.AssignIDs))
		for _, assignID := range e.EventInstance.AssignIDs {
			if _, ok := mapUser[assignID]; ok {
				assigns = append(assigns, respObj{
					ID:       assignID,
					Name:     mapUser[assignID].Name,
					Username: mapUser[assignID].Name,
				})
			}
		}
		out.Assigns = assigns
	}

	if len(e.EventInstance.DepartmentIDs) > 0 {
		out.Depts = make([]respObj, 0, len(e.Departments))
		for _, dept := range e.Departments {
			if _, ok := mapDept[dept.ID]; ok {
				out.Depts = append(out.Depts, respObj{
					ID:   dept.ID,
					Name: mapDept[dept.ID].Name,
				})
			}
		}
	}

	if len(e.EventInstance.RoomIDs) > 0 {
		out.Rooms = make([]respObj, 0, len(e.Rooms))
		for _, room := range e.Rooms {
			if _, ok := mapRoom[room.ID.Hex()]; ok {
				out.Rooms = append(out.Rooms, respObj{
					ID:   room.ID.Hex(),
					Name: mapRoom[room.ID.Hex()].Name,
				})
			}
		}
	}

	if e.EventInstance.CreatedByID != "" {
		out.CreatedBy = respObj{
			ID:       e.EventInstance.CreatedByID,
			Name:     mapUser[e.EventInstance.CreatedByID].Name,
			Username: mapUser[e.EventInstance.CreatedByID].Name,
		}
	}

	if !e.BaseEvent.ID.IsZero() {
		item := newBaseEventResp(e.BaseEvent, mapUser, mapDept, mapRoom, mapCategory, mapTimezone)
		out.BaseEvent = &item
	}
	if e.EventInstance.RepeatUntil != nil {
		repeatUntil := response.DateTime(*e.EventInstance.RepeatUntil)
		out.RepeatUntil = &repeatUntil
	}

	if e.EventInstance.Alert != nil && e.EventInstance.Notify {
		out.Alert = &dateConfig{
			Instant: e.EventInstance.Alert.Instant,
			Unit:    e.EventInstance.Alert.Unit,
			Num:     e.EventInstance.Alert.Num,
			Hour:    e.EventInstance.Alert.Hour,
		}
	}

	if len(e.Branches) > 0 {
		out.Branches = make([]respObj, 0, len(e.Branches))
		for _, branch := range e.Branches {
			out.Branches = append(out.Branches, respObj{
				ID:   branch.ID,
				Name: branch.Name,
			})
		}
	}

	return out
}

type respObj struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Color    string `json:"color,omitempty"`
	Key      string `json:"key,omitempty"`
	Username string `json:"username,omitempty"`
}

type baseEventResp struct {
	ID            string             `json:"id"`
	EventID       string             `json:"event_id"`
	Title         string             `json:"title,omitempty"`
	StartTime     response.DateTime  `json:"start_time"`
	EndTime       response.DateTime  `json:"end_time"`
	AllDay        bool               `json:"all_day,omitempty"`
	Repeat        models.EventRepeat `json:"repeat,omitempty"`
	Assigns       []respObj          `json:"assigns,omitempty"`
	Depts         []respObj          `json:"departments,omitempty"`
	Rooms         []respObj          `json:"rooms,omitempty"`
	Description   string             `json:"description,omitempty"`
	EventCategory *respObj           `json:"category,omitempty"`
	Timezone      respObj            `json:"timezone,omitempty"`
	Attendance    int                `json:"attendance,omitempty"`
	CreatedBy     respObj            `json:"created_by,omitempty"`
	Notify        bool               `json:"notify,omitempty"`
	System        bool               `json:"system,omitempty"`
	RepeatUntil   *response.DateTime `json:"repeat_until,omitempty"`
	Alert         *dateConfig        `json:"alert,omitempty"`
	Public        bool               `json:"public"`
}

type calendarReq struct {
	IDs       []string `form:"ids[]"`
	StartTime string   `form:"start_time" binding:"required"`
	EndTime   string   `form:"end_time" binding:"required"`
}

func (r calendarReq) validate() error {
	_, err := util.StrToDateTime(r.StartTime)
	if err != nil {
		return err
	}

	_, err = util.StrToDateTime(r.EndTime)
	if err != nil {
		return err
	}

	return nil
}

func (r calendarReq) toInput() event.CalendarInput {
	startTime, _ := util.StrToDateTime(r.StartTime)
	endTime, _ := util.StrToDateTime(r.EndTime)

	return event.CalendarInput{
		Filter: event.Filter{
			IDs:       r.IDs,
			StartTime: startTime,
			EndTime:   endTime,
		},
	}
}

func newBaseEventResp(event event.EventInstance,
	userMap map[string]microservice.User,
	deptMap map[string]microservice.Department,
	roomMap map[string]models.Room,
	categoryMap map[string]models.EventCategory,
	timezoneMap map[string]models.Element,
) baseEventResp {
	i := baseEventResp{
		ID:          event.ID.Hex(),
		Title:       event.Title,
		EventID:     event.EventID.Hex(),
		StartTime:   response.DateTime(event.StartTime),
		EndTime:     response.DateTime(event.EndTime),
		AllDay:      event.AllDay,
		Repeat:      event.Repeat,
		Description: event.Description,
		Attendance:  event.Attendance,
		Notify:      event.Notify,
		System:      event.System,
		Public:      event.Public,
	}

	if len(event.AssignIDs) > 0 {
		assigns := make([]respObj, 0, len(event.AssignIDs))
		for _, id := range event.AssignIDs {
			if _, ok := userMap[id]; ok {
				assigns = append(assigns, respObj{
					ID:       id,
					Name:     userMap[id].Name,
					Username: userMap[id].Name,
				})
			}
		}
		i.Assigns = assigns
	}

	if len(event.DepartmentIDs) > 0 {
		depts := make([]respObj, 0, len(event.DepartmentIDs))
		for _, id := range event.DepartmentIDs {
			if _, ok := deptMap[id.Hex()]; ok {
				depts = append(depts, respObj{
					ID:   id.Hex(),
					Name: deptMap[id.Hex()].Name,
				})
			}
		}
		i.Depts = depts
	}

	if len(event.RoomIDs) > 0 {
		rooms := make([]respObj, 0, len(event.RoomIDs))
		for _, id := range event.RoomIDs {
			if _, ok := roomMap[id.Hex()]; ok {
				rooms = append(rooms, respObj{
					ID:   id.Hex(),
					Name: roomMap[id.Hex()].Name,
				})
			}
		}
		i.Rooms = rooms
	}

	if event.CategoryID != nil {
		if _, ok := categoryMap[event.CategoryID.Hex()]; ok {
			i.EventCategory = &respObj{
				ID:    event.CategoryID.Hex(),
				Name:  categoryMap[event.CategoryID.Hex()].Name,
				Color: categoryMap[event.CategoryID.Hex()].Color,
				Key:   categoryMap[event.CategoryID.Hex()].Key,
			}
		}
	}

	if _, ok := timezoneMap[event.TimezoneID.Hex()]; ok {
		i.Timezone = respObj{
			ID:   event.TimezoneID.Hex(),
			Name: timezoneMap[event.TimezoneID.Hex()].Name,
		}
	}

	if event.CreatedByID != "" {
		i.CreatedBy = respObj{
			ID:       event.CreatedByID,
			Name:     userMap[event.CreatedByID].Name,
			Username: userMap[event.CreatedByID].Name,
		}
	}

	if event.RepeatUntil != nil {
		repeatUntil := response.DateTime(*event.RepeatUntil)
		i.RepeatUntil = &repeatUntil
	}

	if event.Alert != nil && event.Notify {
		i.Alert = &dateConfig{
			Instant: event.Alert.Instant,
			Unit:    event.Alert.Unit,
			Num:     event.Alert.Num,
			Hour:    event.Alert.Hour,
		}
	}

	return i
}

type calendarItemResp struct {
	ID            string             `json:"id"`
	Title         string             `json:"title"`
	EventID       string             `json:"event_id"`
	StartTime     response.DateTime  `json:"start_time"`
	EndTime       response.DateTime  `json:"end_time"`
	AllDay        bool               `json:"all_day"`
	Repeat        models.EventRepeat `json:"repeat"`
	Notify        bool               `json:"notify"`
	System        bool               `json:"system"`
	Attendance    int                `json:"attendance"`
	RepeatUntil   *response.DateTime `json:"repeat_until"`
	EventCategory *respObj           `json:"category,omitempty"`
	Assigns       []respObj          `json:"assigns,omitempty"`
	Timezone      respObj            `json:"timezone,omitempty"`
	CreatedByID   string             `json:"created_by_id,omitempty"`
	Public        bool               `json:"public"`
}

func (h handler) newCalendarResp(o event.CalendarOutput) []calendarItemResp {
	mapEventCategory := map[string]models.EventCategory{}
	for _, category := range o.EventCategories {
		mapEventCategory[category.ID.Hex()] = category
	}

	mapUser := map[string]microservice.User{}
	for _, user := range o.Users {
		mapUser[user.ID] = user
	}

	mapTimezone := map[string]models.Element{}
	for _, timezone := range o.Timezones {
		mapTimezone[timezone.ID.Hex()] = timezone
	}

	items := make([]calendarItemResp, len(o.EventInstances))
	for i, event := range o.EventInstances {
		item := calendarItemResp{
			ID:          event.ID.Hex(),
			Title:       event.Title,
			EventID:     event.EventID.Hex(),
			StartTime:   response.DateTime(event.StartTime),
			EndTime:     response.DateTime(event.EndTime),
			AllDay:      event.AllDay,
			Repeat:      event.Repeat,
			Notify:      event.Notify,
			System:      event.System,
			Attendance:  event.Attendance,
			CreatedByID: event.CreatedByID,
			Public:      event.Public,
		}

		if event.RepeatUntil != nil {
			repeatUntil := response.DateTime(*event.RepeatUntil)
			item.RepeatUntil = &repeatUntil
		}

		if event.CategoryID != nil {
			if _, ok := mapEventCategory[event.CategoryID.Hex()]; ok {
				item.EventCategory = &respObj{
					ID:    event.CategoryID.Hex(),
					Name:  mapEventCategory[event.CategoryID.Hex()].Name,
					Color: mapEventCategory[event.CategoryID.Hex()].Color,
					Key:   mapEventCategory[event.CategoryID.Hex()].Key,
				}
			}
		}

		if len(event.AssignIDs) > 0 {
			assigns := make([]respObj, 0, len(event.AssignIDs))
			for _, id := range event.AssignIDs {
				assigns = append(assigns, respObj{
					ID:       id,
					Name:     mapUser[id].Name,
					Username: mapUser[id].Name,
				})
			}
			item.Assigns = assigns
		}

		if _, ok := mapTimezone[event.TimezoneID.Hex()]; ok {
			item.Timezone = respObj{
				ID:   event.TimezoneID.Hex(),
				Name: mapTimezone[event.TimezoneID.Hex()].Name,
				Key:  mapTimezone[event.TimezoneID.Hex()].Key,
			}
		}

		items[i] = item
	}

	return items
}

type updateAttendanceReq struct {
	ID      string `uri:"id"`
	EventID string `uri:"event_id"`
	Status  int    `json:"status" binding:"required,oneof=-1 1"`
}

func (r updateAttendanceReq) toInput() event.UpdateAttendanceInput {
	return event.UpdateAttendanceInput{
		ID:      r.ID,
		EventID: r.EventID,
		Status:  r.Status,
	}
}

type unavailableRoomResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h handler) newUnavailableRoomResp(rooms []models.Room) []unavailableRoomResp {
	resp := make([]unavailableRoomResp, len(rooms))
	for i, room := range rooms {
		resp[i] = unavailableRoomResp{
			ID:   room.ID.Hex(),
			Name: room.Name,
		}
	}
	return resp
}
