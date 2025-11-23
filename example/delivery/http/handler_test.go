package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/jwt"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mockDeps struct {
	uc *event.MockUseCase
}

func initHandler(t *testing.T) (Handler, mockDeps) {
	l := log.InitializeTestZapLogger()
	uc := event.NewMockUseCase(t)

	return New(l, uc), mockDeps{
		uc: uc,
	}
}

func TestCreate(t *testing.T) {
	jwtPayload := jwt.Payload{
		UserID:     "user_id",
		ShopID:     "shop-id",
		ShopPrefix: "a",
	}

	scope := jwt.NewScope(jwtPayload)

	type mockUcCreate struct {
		isCalled bool
		input    event.CreateInput
		output   event.CreateEventOutput
		err      error
	}

	tcs := map[string]struct {
		body     string
		mockUC   mockUcCreate
		isUnauth bool
		wantCode int
		wantBody string
	}{
		"success": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			mockUC: mockUcCreate{
				isCalled: true,
				input: event.CreateInput{
					Title:         "Test Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       nil,
					Description:   "",
					CategoryID:    "",
					RepeatUntil:   nil,
					Notify:        false,
					System:        false,
					Alert:         nil,
					ObjectID:      "",
					Public:        false,
				},
				output: event.CreateEventOutput{},
				err:    nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"unauthorized": {
			body:     `{}`,
			isUnauth: true,
			wantCode: http.StatusUnauthorized,
			wantBody: `{
				"error_code": 401,
				"message": "Unauthorized"
			}`,
		},
		"invalid_body": {
			body: `{
				"title": "Test Event",
				"start_time": "invalid"
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_start_time": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "invalid-time-format",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_end_time": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "invalid-time-format",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_category_id": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"category_id": "invalid-object-id",
				"notify": false,
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_repeat_until": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "daily",
				"repeat_until": "invalid-time-format",
				"notify": false,
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_repeat_until_before_now": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "daily",
				"repeat_until": "2020-01-01 10:00:00",
				"notify": false,
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_alert_invalid_unit": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "invalid_unit",
					"num": 15,
					"hour": 0
				},
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_alert_allday_invalid_hour": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": true,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "day",
					"num": 1,
					"hour": 25
				},
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_alert_allday_negative_num": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": true,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "day",
					"num": -1,
					"hour": 9
				},
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_alert_empty_unit_with_num": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "",
					"num": 15,
					"hour": 0
				},
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_alert_unit_with_zero_num": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "minute",
					"num": 0,
					"hour": 0
				},
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_assign_without_branch": {
			body: `{
				"title": "Test Event",
				"branch_ids": [],
				"assign_ids": ["5f9c0b9b9c6b9a0001b9c6b5"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_department_without_branch": {
			body: `{
				"title": "Test Event",
				"branch_ids": [],
				"department_ids": ["5f9c0b9b9c6b9a0001b9c6b6"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"error_create": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"room_ids": ["5f9c0b9b9c6b9a0001b9c6b3"],
				"description": "Test description",
				"category_id": "5f9c0b9b9c6b9a0001b9c6b4",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "minute",
					"num": 15,
					"hour": 0
				},
				"assign_ids": [],
				"department_ids": [],
				"public": false
			}`,
			mockUC: mockUcCreate{
				isCalled: true,
				input: event.CreateInput{
					Title:         "Test Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       []string{"5f9c0b9b9c6b9a0001b9c6b3"},
					Description:   "Test description",
					CategoryID:    "5f9c0b9b9c6b9a0001b9c6b4",
					Notify:        true,
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					Public:        false,
					Alert: &models.DateConfig{
						Instant: false,
						Unit:    models.DateUnitMinute,
						Num:     15,
						Hour:    0,
					},
				},
				err: event.ErrRequiredField,
			},
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 141005,
				"message": "Required field"
			}`,
		},
		"room_unavailable": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"room_ids": ["5f9c0b9b9c6b9a0001b9c6b3"],
				"description": "Test description",
				"category_id": "5f9c0b9b9c6b9a0001b9c6b4",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "minute",
					"num": 15,
					"hour": 0
				},
				"assign_ids": [],
				"department_ids": [],
				"public": false
			}`,
			mockUC: mockUcCreate{
				isCalled: true,
				input: event.CreateInput{
					Title:         "Test Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       []string{"5f9c0b9b9c6b9a0001b9c6b3"},
					Description:   "Test description",
					CategoryID:    "5f9c0b9b9c6b9a0001b9c6b4",
					Notify:        true,
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					Public:        false,
					Alert: &models.DateConfig{
						Instant: false,
						Unit:    models.DateUnitMinute,
						Num:     15,
						Hour:    0,
					},
				},
				output: event.CreateEventOutput{
					EventInstance: event.EventInstance{},
					UnavailableRooms: []models.Room{
						{
							ID:   mongo.ObjectIDFromHexOrNil("5f9c0b9b9c6b9a0001b9c6b3"),
							Name: "Conference Room A",
						},
					},
				},
				err: event.ErrRoomUnavailable,
			},
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140008,
				"message": "Room unavailable",
				"data": [
					{
						"id": "5f9c0b9b9c6b9a0001b9c6b3",
						"name": "Conference Room A"
					}
				]
			}`,
		},
		"success_with_allday_valid_alert": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": true,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "day",
					"num": 1,
					"hour": 9
				},
				"public": false
			}`,
			mockUC: mockUcCreate{
				isCalled: true,
				input: event.CreateInput{
					Title:         "Test Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        true,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       nil,
					Description:   "",
					CategoryID:    "",
					RepeatUntil:   nil,
					Notify:        true,
					System:        false,
					Alert: &models.DateConfig{
						Instant: false,
						Unit:    models.DateUnitDay,
						Num:     1,
						Hour:    9,
					},
					ObjectID: "",
					Public:   false,
				},
				output: event.CreateEventOutput{},
				err:    nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"success_with_repeat_until": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "daily",
				"repeat_until": "2025-12-31 23:59:59",
				"notify": false,
				"public": false
			}`,
			mockUC: mockUcCreate{
				isCalled: true,
				input: event.CreateInput{
					Title:         "Test Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatDaily,
					RoomIDs:       nil,
					Description:   "",
					CategoryID:    "",
					RepeatUntil:   util.ToPointer(time.Date(2025, 12, 31, 23, 59, 59, 0, time.Local)),
					Notify:        false,
					System:        false,
					Alert:         nil,
					ObjectID:      "",
					Public:        false,
				},
				output: event.CreateEventOutput{},
				err:    nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"success_with_non_allday_valid_alert": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "minute",
					"num": 15,
					"hour": 0
				},
				"public": false
			}`,
			mockUC: mockUcCreate{
				isCalled: true,
				input: event.CreateInput{
					Title:         "Test Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       nil,
					Description:   "",
					CategoryID:    "",
					RepeatUntil:   nil,
					Notify:        true,
					System:        false,
					Alert: &models.DateConfig{
						Instant: false,
						Unit:    models.DateUnitMinute,
						Num:     15,
						Hour:    0,
					},
					ObjectID: "",
					Public:   false,
				},
				output: event.CreateEventOutput{},
				err:    nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"success_with_instant_alert": {
			body: `{
				"title": "Test Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": true,
					"unit": "",
					"num": 0,
					"hour": 0
				},
				"public": false
			}`,
			mockUC: mockUcCreate{
				isCalled: true,
				input: event.CreateInput{
					Title:         "Test Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       nil,
					Description:   "",
					CategoryID:    "",
					RepeatUntil:   nil,
					Notify:        true,
					System:        false,
					Alert: &models.DateConfig{
						Instant: true,
						Unit:    "",
						Num:     0,
						Hour:    0,
					},
					ObjectID: "",
					Public:   false,
				},
				output: event.CreateEventOutput{},
				err:    nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t)

			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)

			engine.POST("/api/v1/events", h.Create)

			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")

			if !tc.isUnauth {
				c.Request = c.Request.WithContext(
					jwt.SetPayloadToContext(c.Request.Context(), jwtPayload),
				)
			}

			if tc.mockUC.isCalled {
				deps.uc.EXPECT().Create(c.Request.Context(), scope, tc.mockUC.input).
					Return(tc.mockUC.output, tc.mockUC.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestDetail(t *testing.T) {
	jwtPayload := jwt.Payload{
		UserID:     "user_id",
		ShopID:     "shop-id",
		ShopPrefix: "a",
	}

	scope := jwt.NewScope(jwtPayload)

	type mockUcDetail struct {
		isCalled bool
		id       string
		eventID  string
		err      error
	}

	tcs := map[string]struct {
		id       string
		eventID  string
		mockUC   mockUcDetail
		isUnauth bool
		wantCode int
		wantBody string
	}{
		"success": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b2",
			mockUC: mockUcDetail{
				isCalled: true,
				id:       "5f9c0b9b9c6b9a0001b9c6b1",
				eventID:  "5f9c0b9b9c6b9a0001b9c6b2",
				err:      nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success",
				"data": {
					"id": "5f9c0b9b9c6b9a0001b9c6b1",
					"event_id": "5f9c0b9b9c6b9a0001b9c6b2",
					"title": "Test Event",
					"start_time": "2024-05-05 17:00:00",
					"end_time": "2024-05-05 18:00:00",
					"repeat": "none",
					"description": "Test description",
					"notify": true,
					"attendance": 0,
					"created_by": {},
					"system": false,
					"public": false,
					"timezone": {},
					"timezone_id": "000000000000000000000000"
				}
			}`,
		},
		"unauthorized": {
			id:       "5f9c0b9b9c6b9a0001b9c6b1",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b2",
			isUnauth: true,
			wantCode: http.StatusUnauthorized,
			wantBody: `{
				"error_code": 401,
				"message": "Unauthorized"
			}`,
		},
		"invalid_id": {
			id:       "invalid",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b2",
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"invalid_event_id": {
			id:       "5f9c0b9b9c6b9a0001b9c6b1",
			eventID:  "invalid",
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140002,
				"message": "Wrong query"
			}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t)

			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)

			path := fmt.Sprintf("/api/v1/events/%s/%s", tc.eventID, tc.id)
			engine.GET("/api/v1/events/:event_id/:id", h.Detail)

			c.Request = httptest.NewRequest(http.MethodGet, path, nil)

			if !tc.isUnauth {
				c.Request = c.Request.WithContext(
					jwt.SetPayloadToContext(c.Request.Context(), jwtPayload),
				)
			}

			if tc.mockUC.isCalled {
				id, _ := primitive.ObjectIDFromHex(tc.mockUC.id)
				eventID, _ := primitive.ObjectIDFromHex(tc.mockUC.eventID)

				deps.uc.EXPECT().Detail(c.Request.Context(), scope, tc.mockUC.id, tc.mockUC.eventID).
					Return(event.DetailOutput{
						EventInstance: event.EventInstance{
							ID:          id,
							EventID:     eventID,
							Title:       "Test Event",
							StartTime:   time.Date(2024, 5, 5, 10, 0, 0, 0, time.UTC),
							EndTime:     time.Date(2024, 5, 5, 11, 0, 0, 0, time.UTC),
							AllDay:      false,
							Repeat:      models.EventRepeatNone,
							Description: "Test description",
							Notify:      true,
						},
					}, tc.mockUC.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestDelete(t *testing.T) {
	jwtPayload := jwt.Payload{
		UserID:     "user_id",
		ShopID:     "shop-id",
		ShopPrefix: "a",
	}

	scope := jwt.NewScope(jwtPayload)

	type mockUcDelete struct {
		isCalled bool
		input    event.DeleteInput
		err      error
	}

	tcs := map[string]struct {
		id       string
		eventID  string
		typeVal  string
		mockUC   mockUcDelete
		isUnauth bool
		wantCode int
		wantBody string
	}{
		"success": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			mockUC: mockUcDelete{
				isCalled: true,
				input: event.DeleteInput{
					ID:      "5f9c0b9b9c6b9a0001b9c6b1",
					EventID: "5f9c0b9b9c6b9a0001b9c6b1",
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
					"error_code": 0,
					"message": "Success"
				}`,
		},
		"success_with_type": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b2",
			typeVal: "from",
			mockUC: mockUcDelete{
				isCalled: true,
				input: event.DeleteInput{
					ID:      "5f9c0b9b9c6b9a0001b9c6b1",
					EventID: "5f9c0b9b9c6b9a0001b9c6b2",
					Type:    models.EventActionFrom,
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
					"error_code": 0,
					"message": "Success"
				}`,
		},
		"unauthorized": {
			id:       "5f9c0b9b9c6b9a0001b9c6b1",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b1",
			isUnauth: true,
			wantCode: http.StatusUnauthorized,
			wantBody: `{
					"error_code": 401,
					"message": "Unauthorized"
				}`,
		},
		"invalid_id": {
			id:       "invalid-id",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b1",
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
					"error_code": 140002,
					"message": "Wrong query"
				}`,
		},
		"invalid_validation": {
			id:       "5f9c0b9b9c6b9a0001b9c6b1",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b2",
			typeVal:  "",
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
					"error_code": 140002,
					"message": "Wrong query"
				}`,
		},
		"invalid_validation_case2": {
			id:       "5f9c0b9b9c6b9a0001b9c6b1",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b1",
			typeVal:  "one",
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
					"error_code": 140002,
					"message": "Wrong query"
				}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t)

			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)

			path := fmt.Sprintf("/api/v1/events/%s/%s", tc.eventID, tc.id)
			if tc.typeVal != "" {
				path = fmt.Sprintf("%s?type=%s", path, tc.typeVal)
			}

			engine.DELETE("/api/v1/events/:event_id/:id", h.Delete)

			c.Request = httptest.NewRequest(http.MethodDelete, path, nil)

			if !tc.isUnauth {
				c.Request = c.Request.WithContext(
					jwt.SetPayloadToContext(c.Request.Context(), jwtPayload),
				)
			}

			if tc.mockUC.isCalled {
				deps.uc.EXPECT().Delete(c.Request.Context(), scope, tc.mockUC.input).
					Return(tc.mockUC.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestUpdate(t *testing.T) {
	jwtPayload := jwt.Payload{
		UserID:     "user_id",
		ShopID:     "shop-id",
		ShopPrefix: "a",
	}

	scope := jwt.NewScope(jwtPayload)

	type mockUcUpdate struct {
		isCalled bool
		input    event.UpdateInput
		err      error
	}

	tcs := map[string]struct {
		id       string
		eventID  string
		body     string
		mockUC   mockUcUpdate
		isUnauth bool
		wantCode int
		wantBody string
	}{
		"success": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"room_ids": [],
				"description": "Updated description",
				"category_id": "",
				"notify": false,
				"public": false
			}`,
			mockUC: mockUcUpdate{
				isCalled: true,
				input: event.UpdateInput{
					ID:            "5f9c0b9b9c6b9a0001b9c6b1",
					EventID:       "5f9c0b9b9c6b9a0001b9c6b1",
					Title:         "Updated Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       []string{},
					Type:          models.EventAction(""),
					CategoryID:    "",
					Description:   "Updated description",
					Notify:        false,
					RepeatUntil:   nil,
					Alert:         nil,
					Public:        false,
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"success_with_repeat_until": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "daily",
				"repeat_until": "2025-12-31 23:59:59",
				"notify": false,
				"public": false
			}`,
			mockUC: mockUcUpdate{
				isCalled: true,
				input: event.UpdateInput{
					ID:            "5f9c0b9b9c6b9a0001b9c6b1",
					EventID:       "5f9c0b9b9c6b9a0001b9c6b1",
					Title:         "Updated Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatDaily,
					RoomIDs:       nil,
					Type:          models.EventAction(""),
					CategoryID:    "",
					Description:   "",
					Notify:        false,
					RepeatUntil:   util.ToPointer(time.Date(2025, 12, 31, 23, 59, 59, 0, time.Local)),
					Alert:         nil,
					Public:        false,
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"success_with_alert": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "minute",
					"num": 15,
					"hour": 0
				},
				"public": false
			}`,
			mockUC: mockUcUpdate{
				isCalled: true,
				input: event.UpdateInput{
					ID:            "5f9c0b9b9c6b9a0001b9c6b1",
					EventID:       "5f9c0b9b9c6b9a0001b9c6b1",
					Title:         "Updated Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       nil,
					Type:          models.EventAction(""),
					CategoryID:    "",
					Description:   "",
					Notify:        true,
					RepeatUntil:   nil,
					Alert: &models.DateConfig{
						Instant: false,
						Unit:    models.DateUnitMinute,
						Num:     15,
						Hour:    0,
					},
					Public: false,
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"success_with_type_one": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b2",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"assign_ids": [],
				"department_ids": [],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"type": "one",
				"notify": false,
				"public": false
			}`,
			mockUC: mockUcUpdate{
				isCalled: true,
				input: event.UpdateInput{
					ID:            "5f9c0b9b9c6b9a0001b9c6b1",
					EventID:       "5f9c0b9b9c6b9a0001b9c6b2",
					Title:         "Updated Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     []string{},
					DepartmentIDs: []string{},
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       nil,
					Type:          models.EventActionOne,
					CategoryID:    "",
					Description:   "",
					Notify:        false,
					RepeatUntil:   nil,
					Alert:         nil,
					Public:        false,
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"unauthorized": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			isUnauth: true,
			wantCode: http.StatusUnauthorized,
			wantBody: `{
				"error_code": 401,
				"message": "Unauthorized"
			}`,
		},
		"invalid_body": {
			id:       "5f9c0b9b9c6b9a0001b9c6b1",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b1",
			body:     `{"invalid": "json"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_start_time": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "invalid-time",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_end_time": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "invalid-time",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_start_after_end": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 12:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_category_id": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"category_id": "invalid-id",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_invalid_repeat_until": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "daily",
				"repeat_until": "invalid-time",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_id_mismatch_no_type": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b2",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_id_same_with_type": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"type": "one",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_assign_without_branch": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": [],
				"assign_ids": ["5f9c0b9b9c6b9a0001b9c6b3"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_department_without_branch": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": [],
				"department_ids": ["5f9c0b9b9c6b9a0001b9c6b3"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"err_validation_alert_invalid_unit": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": true,
				"alert": {
					"instant": false,
					"unit": "invalid",
					"num": 15,
					"hour": 0
				},
				"public": false
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
		"error_update": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b1",
			body: `{
				"title": "Updated Event",
				"branch_ids": ["5f9c0b9b9c6b9a0001b9c6b1"],
				"timezone_id": "5f9c0b9b9c6b9a0001b9c6b2",
				"start_time": "2024-05-05 10:00:00",
				"end_time": "2024-05-05 11:00:00",
				"all_day": false,
				"repeat": "none",
				"notify": false,
				"public": false
			}`,
			mockUC: mockUcUpdate{
				isCalled: true,
				input: event.UpdateInput{
					ID:            "5f9c0b9b9c6b9a0001b9c6b1",
					EventID:       "5f9c0b9b9c6b9a0001b9c6b1",
					Title:         "Updated Event",
					BranchIDs:     []string{"5f9c0b9b9c6b9a0001b9c6b1"},
					AssignIDs:     nil,
					DepartmentIDs: nil,
					TimezoneID:    "5f9c0b9b9c6b9a0001b9c6b2",
					StartTime:     time.Date(2024, 5, 5, 10, 0, 0, 0, time.Local),
					EndTime:       time.Date(2024, 5, 5, 11, 0, 0, 0, time.Local),
					AllDay:        false,
					Repeat:        models.EventRepeatNone,
					RoomIDs:       nil,
					Type:          models.EventAction(""),
					CategoryID:    "",
					Description:   "",
					Notify:        false,
					RepeatUntil:   nil,
					Alert:         nil,
					Public:        false,
				},
				err: event.ErrEventNotFound,
			},
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 141004,
				"message": "Event not found"
			}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t)

			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)

			path := fmt.Sprintf("/api/v1/events/%s/%s", tc.eventID, tc.id)
			engine.PUT("/api/v1/events/:event_id/:id", h.Update)

			c.Request = httptest.NewRequest(http.MethodPut, path, strings.NewReader(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")

			if !tc.isUnauth {
				c.Request = c.Request.WithContext(
					jwt.SetPayloadToContext(c.Request.Context(), jwtPayload),
				)
			}

			c.Params = gin.Params{
				{Key: "event_id", Value: tc.eventID},
				{Key: "id", Value: tc.id},
			}

			if tc.mockUC.isCalled {
				deps.uc.EXPECT().Update(c.Request.Context(), scope, tc.mockUC.input).
					Return(tc.mockUC.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestList(t *testing.T) {
	jwtPayload := jwt.Payload{
		UserID:     "user_id",
		ShopID:     "shop-id",
		ShopPrefix: "a",
	}

	scope := jwt.NewScope(jwtPayload)

	type mockUcCalendar struct {
		isCalled bool
		input    event.CalendarInput
		output   event.CalendarOutput
		err      error
	}

	tcs := map[string]struct {
		query    string
		mockUC   mockUcCalendar
		isUnauth bool
		wantCode int
		wantBody string
	}{
		"success": {
			query: "start_time=2024-05-05%2017:00:00&end_time=2024-05-05%2018:00:00",
			mockUC: mockUcCalendar{
				isCalled: true,
				input: event.CalendarInput{
					Filter: event.Filter{
						StartTime: time.Date(2024, 5, 5, 17, 0, 0, 0, time.Local),
						EndTime:   time.Date(2024, 5, 5, 18, 0, 0, 0, time.Local),
					},
				},
				output: event.CalendarOutput{
					EventInstances: []event.EventInstance{
						{
							ID:         primitive.ObjectID{},
							Title:      "Test Event",
							EventID:    primitive.ObjectID{},
							StartTime:  time.Date(2024, 5, 5, 17, 0, 0, 0, time.Local),
							EndTime:    time.Date(2024, 5, 5, 18, 0, 0, 0, time.Local),
							AllDay:     false,
							Repeat:     models.EventRepeatNone,
							Notify:     true,
							System:     false,
							Attendance: 0,
							TimezoneID: primitive.ObjectID{},
						},
					},
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"data": [
					{
						"id": "000000000000000000000000",
						"title": "Test Event",
						"event_id": "000000000000000000000000",
						"start_time": "2024-05-05 17:00:00",
						"end_time": "2024-05-05 18:00:00",
						"all_day": false,
						"repeat": "none",
						"notify": true,
						"system": false,
						"attendance": 0,
						"repeat_until": null,
						"timezone": {},
						"public": false
					}
				],
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"unauthorized": {
			query:    "start_time=2024-05-05%2017:00:00&end_time=2024-05-05%2018:00:00",
			isUnauth: true,
			wantCode: http.StatusUnauthorized,
			wantBody: `{
				"error_code": 401,
				"message": "Unauthorized"
			}`,
		},
		"invalid_time": {
			query:    "start_time=invalid&end_time=invalid",
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t)

			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)

			engine.GET("/api/v1/calendar", h.List)

			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/calendar?"+tc.query, nil)
			c.Request.Header.Set("Content-Type", "application/json")

			if !tc.isUnauth {
				c.Request = c.Request.WithContext(
					jwt.SetPayloadToContext(c.Request.Context(), jwtPayload),
				)
			}

			if tc.mockUC.isCalled {
				deps.uc.EXPECT().Calendar(c.Request.Context(), scope, tc.mockUC.input).
					Return(tc.mockUC.output, tc.mockUC.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestUpdateAttendance(t *testing.T) {
	jwtPayload := jwt.Payload{
		UserID:     "user_id",
		ShopID:     "shop-id",
		ShopPrefix: "a",
	}

	scope := jwt.NewScope(jwtPayload)

	type mockUcUpdateAttendance struct {
		isCalled bool
		input    event.UpdateAttendanceInput
		err      error
	}

	tcs := map[string]struct {
		id       string
		eventID  string
		body     string
		mockUC   mockUcUpdateAttendance
		isUnauth bool
		wantCode int
		wantBody string
	}{
		"success": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b2",
			body: `{
				"status": 1
			}`,
			mockUC: mockUcUpdateAttendance{
				isCalled: true,
				input: event.UpdateAttendanceInput{
					ID:      "5f9c0b9b9c6b9a0001b9c6b1",
					EventID: "5f9c0b9b9c6b9a0001b9c6b2",
					Status:  1,
				},
				err: nil,
			},
			isUnauth: false,
			wantCode: http.StatusOK,
			wantBody: `{
				"error_code": 0,
				"message": "Success"
			}`,
		},
		"unauthorized": {
			id:       "5f9c0b9b9c6b9a0001b9c6b1",
			eventID:  "5f9c0b9b9c6b9a0001b9c6b2",
			isUnauth: true,
			wantCode: http.StatusUnauthorized,
			wantBody: `{
				"error_code": 401,
				"message": "Unauthorized"
			}`,
		},
		"invalid_status": {
			id:      "5f9c0b9b9c6b9a0001b9c6b1",
			eventID: "5f9c0b9b9c6b9a0001b9c6b2",
			body: `{
				"status": 2
			}`,
			isUnauth: false,
			wantCode: http.StatusBadRequest,
			wantBody: `{
				"error_code": 140003,
				"message": "Wrong body"
			}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t)

			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)

			path := fmt.Sprintf("/api/v1/events/attendance/%s/%s", tc.eventID, tc.id)
			engine.PATCH("/api/v1/events/attendance/:event_id/:id", h.UpdateAttendance)

			c.Request = httptest.NewRequest(http.MethodPatch, path, strings.NewReader(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")

			if !tc.isUnauth {
				c.Request = c.Request.WithContext(
					jwt.SetPayloadToContext(c.Request.Context(), jwtPayload),
				)
			}

			if tc.mockUC.isCalled {
				deps.uc.EXPECT().UpdateAttendance(c.Request.Context(), scope, tc.mockUC.input).
					Return(tc.mockUC.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}
