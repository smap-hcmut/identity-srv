package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/jwt"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
	pkgmongo "gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mockDeps struct {
	db  *pkgmongo.MockDatabase
	col *pkgmongo.MockCollection
	cur *pkgmongo.MockCursor
}

var (
	fixedTime = time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
)

func initRepo(t *testing.T, mockTime time.Time) (repository.Repository, mockDeps) {
	l := log.InitializeTestZapLogger()

	db := pkgmongo.NewMockDatabase(t)
	col := pkgmongo.NewMockCollection(t)
	cur := pkgmongo.NewMockCursor(t)

	repo := &implRepository{
		l:     l,
		db:    db,
		clock: func() time.Time { return mockTime.UTC() },
	}
	return repo, mockDeps{
		db:  db,
		col: col,
		cur: cur,
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

type dummySingleResult struct {
	result models.RecurringInstance
	err    error
}

func (d *dummySingleResult) Decode(v interface{}) error {
	*(v.(*models.RecurringInstance)) = d.result
	return d.err
}

func ptrInt32(i int32) *int32 {
	return &i
}

func TestEventRepository_Create(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	timeZoneID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	categoryID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439022")
	objectID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439033")
	branchID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439044")
	departmentID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439055")
	roomID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439066")

	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
	// Use the actual start time from input to generate the correct ID
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Hour)
	repeatUntil := startTime.Add(24 * time.Hour * 7)
	notifyTime := startTime.Add(-time.Hour)

	scope := models.Scope{
		UserID: "user-id",
		ShopID: mockShopID.Hex(),
	}

	systemScope := models.Scope{
		UserID: "system-user-id",
		ShopID: mockShopID.Hex(),
	}

	// invalidScope := models.Scope{
	// 	UserID: "user-id",
	// 	ShopID: "invalid-shop-id", // Invalid hex for ObjectID
	// }

	tcs := map[string]struct {
		mockCol    func(t *testing.T) *pkgmongo.MockCollection
		mockSetup  func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection)
		scope      models.Scope
		input      repository.CreateOptions
		wantErr    bool
		errMessage string
		validate   func(t *testing.T, event models.Event)
	}{
		"success_minimal": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return pkgmongo.NewMockCollection(t)
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// Calculate period and year for expected collection name
				expectedID := primitive.NewObjectIDFromTimestamp(startTime)
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(expectedID)
				collectionName := fmt.Sprintf("%s_%d_%d", "events", year, period)

				deps.db.EXPECT().Collection(collectionName).Return(mockCol)
				mockCol.EXPECT().InsertOne(ctx, mock.Anything).Return(nil, nil)
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Meeting Event A",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
			},
			wantErr: false,
			validate: func(t *testing.T, event models.Event) {
				require.Equal(t, "Meeting Event A", event.Title)
				require.Equal(t, mockShopID, event.ShopID)
				require.Equal(t, timeZoneID, event.TimezoneID)
				require.Equal(t, startTime, event.StartTime)
				require.Equal(t, endTime, event.EndTime)
				require.Equal(t, "user-id", event.CreatedByID)
				require.Equal(t, models.EventRepeatNone, event.Repeat)
				require.Equal(t, false, event.AllDay)
			},
		},
		"success_all_day_event": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return pkgmongo.NewMockCollection(t)
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				expectedID := primitive.NewObjectIDFromTimestamp(startTime)
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(expectedID)
				collectionName := fmt.Sprintf("%s_%d_%d", "events", year, period)

				deps.db.EXPECT().Collection(collectionName).Return(mockCol)
				mockCol.EXPECT().InsertOne(ctx, mock.Anything).Return(nil, nil)
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "All Day Event",
				StartTime:  startTime,
				EndTime:    endTime, // This will be ignored for all-day events
				TimezoneID: timeZoneID.Hex(),
				AllDay:     true,
			},
			wantErr: false,
			validate: func(t *testing.T, event models.Event) {
				require.Equal(t, "All Day Event", event.Title)
				require.Equal(t, true, event.AllDay)
				// Start time should be beginning of day
				require.Equal(t, util.StartOfDay(startTime), event.StartTime)
				// End time should be end of day
				require.Equal(t, util.EndOfDay(startTime), event.EndTime)
			},
		},
		"success_system_event": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return pkgmongo.NewMockCollection(t)
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				expectedID := primitive.NewObjectIDFromTimestamp(startTime)
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(expectedID)
				collectionName := fmt.Sprintf("%s_%d_%d", "events", year, period)

				deps.db.EXPECT().Collection(collectionName).Return(mockCol)
				mockCol.EXPECT().InsertOne(ctx, mock.Anything).Return(nil, nil)
			},
			scope: systemScope,
			input: repository.CreateOptions{
				Title:      "System Event",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
				System:     true,
			},
			wantErr: false,
			validate: func(t *testing.T, event models.Event) {
				require.Equal(t, "System Event", event.Title)
				require.Equal(t, true, event.System)
				// System events should not have CreatedByID
				require.Empty(t, event.CreatedByID)
			},
		},
		"success_full_event": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return pkgmongo.NewMockCollection(t)
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				expectedID := primitive.NewObjectIDFromTimestamp(startTime)
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(expectedID)
				collectionName := fmt.Sprintf("%s_%d_%d", "events", year, period)

				deps.db.EXPECT().Collection(collectionName).Return(mockCol)
				mockCol.EXPECT().InsertOne(ctx, mock.Anything).Return(nil, nil)
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:         "Complete Event",
				BranchIDs:     []string{branchID.Hex()},
				AssignIDs:     []string{"assign-1", "assign-2"},
				DepartmentIDs: []string{departmentID.Hex()},
				TimezoneID:    timeZoneID.Hex(),
				StartTime:     startTime,
				EndTime:       endTime,
				AllDay:        false,
				Repeat:        models.EventRepeatWeekly,
				RoomIDs:       []string{roomID.Hex()},
				Description:   "Event description",
				CategoryID:    categoryID.Hex(),
				RepeatUntil:   &repeatUntil,
				Notify:        true,
				System:        false,
				NotifyTime:    &notifyTime,
				Alert: &models.DateConfig{
					Num:     15,
					Unit:    models.DateUnitMinute,
					Hour:    9,
					Instant: false,
				},
				ObjectID: objectID.Hex(),
			},
			wantErr: false,
			validate: func(t *testing.T, event models.Event) {
				// Check ID generation
				expectedID := primitive.NewObjectIDFromTimestamp(startTime)
				require.Equal(t, expectedID.Timestamp().Unix(), event.ID.Timestamp().Unix())

				// Check basic fields
				require.Equal(t, "Complete Event", event.Title)
				require.Equal(t, mockShopID, event.ShopID)
				require.Equal(t, timeZoneID, event.TimezoneID)
				require.Equal(t, startTime, event.StartTime)
				require.Equal(t, endTime, event.EndTime)
				require.Equal(t, "user-id", event.CreatedByID)

				// Check IDs
				require.Len(t, event.BranchIDs, 1)
				require.Equal(t, branchID, event.BranchIDs[0])
				require.Len(t, event.AssignIDs, 2)
				require.Equal(t, []string{"assign-1", "assign-2"}, event.AssignIDs)
				require.Len(t, event.DepartmentIDs, 1)
				require.Equal(t, departmentID, event.DepartmentIDs[0])
				require.Len(t, event.RoomIDs, 1)
				require.Equal(t, roomID, event.RoomIDs[0])

				// Check pointers
				require.NotNil(t, event.CategoryID)
				require.Equal(t, categoryID, *event.CategoryID)
				require.NotNil(t, event.RepeatUntil)
				require.Equal(t, repeatUntil, *event.RepeatUntil)
				require.NotNil(t, event.NotifyTime)
				require.Equal(t, notifyTime, *event.NotifyTime)
				require.NotNil(t, event.ObjectID)
				require.Equal(t, objectID, *event.ObjectID)

				// Check Alert configuration
				require.NotNil(t, event.Alert)
				require.Equal(t, 15, event.Alert.Num)
				require.Equal(t, models.DateUnitMinute, event.Alert.Unit)
				require.Equal(t, 9, event.Alert.Hour)
				require.Equal(t, false, event.Alert.Instant)

				// Check other fields
				require.Equal(t, "Event description", event.Description)
				require.Equal(t, models.EventRepeatWeekly, event.Repeat)
				require.Equal(t, true, event.Notify)
				require.Equal(t, false, event.System)
			},
		},
		// "error_invalid_shop_id": {
		// 	mockCol: func(t *testing.T) *pkgmongo.MockCollection {
		// 		return nil // Won't be used
		// 	},
		// 	mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
		// 		// No collection setup needed as we expect early failure
		// 	},
		// 	scope: invalidScope,
		// 	input: repository.CreateOptions{
		// 		Title:      "Invalid Shop ID Event",
		// 		StartTime:  startTime,
		// 		EndTime:    endTime,
		// 		TimezoneID: timeZoneID.Hex(),
		// 		AllDay:     false,
		// 	},
		// 	wantErr:    true,
		// 	errMessage: errors.New("the provided hex string is not a valid ObjectID").Error(),
		// },
		"error_invalid_timezone_id": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return nil // Won't be used
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// No collection setup needed as we expect early failure
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Invalid Timezone ID Event",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: "invalid-timezone-id",
				AllDay:     false,
			},
			wantErr:    true,
			errMessage: errors.New("the provided hex string is not a valid ObjectID").Error(),
		},
		"error_invalid_department_ids": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return nil
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:         "Invalid Department ID Event",
				StartTime:     startTime,
				EndTime:       endTime,
				TimezoneID:    timeZoneID.Hex(),
				AllDay:        false,
				DepartmentIDs: []string{"invalid-department-id"},
			},
			wantErr:    true,
			errMessage: pkgmongo.ErrInvalidObjectID.Error(),
		},
		"error_invalid_room_ids": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return nil // Won't be used
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// No collection setup needed as we expect early failure
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Invalid Room ID Event",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
				RoomIDs:    []string{"invalid-room-id"},
			},
			wantErr:    true,
			errMessage: pkgmongo.ErrInvalidObjectID.Error(),
		},
		"error_invalid_category_id": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return nil // Won't be used
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// No collection setup needed as we expect early failure
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Invalid Category ID Event",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
				CategoryID: "invalid-category-id",
			},
			wantErr:    true,
			errMessage: errors.New("the provided hex string is not a valid ObjectID").Error(),
		},
		"error_invalid_object_id": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return nil // Won't be used
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// No collection setup needed as we expect early failure
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Invalid Object ID Event",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
				ObjectID:   "invalid-object-id",
			},
			wantErr:    true,
			errMessage: errors.New("the provided hex string is not a valid ObjectID").Error(),
		},
		"error_invalid_branch_ids": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return nil // Won't be used
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// No collection setup needed as we expect early failure
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Invalid Branch ID Event",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
				BranchIDs:  []string{"invalid-branch-id"},
			},
			wantErr:    true,
			errMessage: pkgmongo.ErrInvalidObjectID.Error(),
		},
		"error_database_insert_failure": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return pkgmongo.NewMockCollection(t)
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				expectedID := primitive.NewObjectIDFromTimestamp(startTime)
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(expectedID)
				collectionName := fmt.Sprintf("%s_%d_%d", "events", year, period)

				deps.db.EXPECT().Collection(collectionName).Return(mockCol)
				// Simulate database error during insert
				mockCol.EXPECT().InsertOne(ctx, mock.Anything).Return(nil, errors.New("database error"))
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Database Error Event",
				StartTime:  startTime,
				EndTime:    endTime,
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
			},
			wantErr:    true,
			errMessage: errors.New("database error").Error(),
		},
		"success_different_period_collection": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return pkgmongo.NewMockCollection(t)
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// Use a different quarter to test period collection calculation
				differentPeriodTime := time.Date(2024, 7, 15, 10, 0, 0, 0, time.UTC) // Q3
				expectedID := primitive.NewObjectIDFromTimestamp(differentPeriodTime)
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(expectedID)
				collectionName := fmt.Sprintf("%s_%d_%d", "events", year, period)

				deps.db.EXPECT().Collection(collectionName).Return(mockCol)
				mockCol.EXPECT().InsertOne(ctx, mock.Anything).Return(nil, nil)
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Q3 Period Event",
				StartTime:  time.Date(2024, 7, 15, 10, 0, 0, 0, time.UTC), // Q3
				EndTime:    time.Date(2024, 7, 15, 11, 0, 0, 0, time.UTC),
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
			},
			wantErr: false,
			validate: func(t *testing.T, event models.Event) {
				differentPeriodTime := time.Date(2024, 7, 15, 10, 0, 0, 0, time.UTC)
				// Verify ID is based on the Q3 date
				expectedID := primitive.NewObjectIDFromTimestamp(differentPeriodTime)
				require.Equal(t, expectedID.Timestamp().Unix(), event.ID.Timestamp().Unix())

				// Verify period calculation
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(event.ID)
				require.Equal(t, int32(3), period) // Should be Q3
				require.Equal(t, int32(2024), year)
			},
		},
		"success_different_year_collection": {
			mockCol: func(t *testing.T) *pkgmongo.MockCollection {
				return pkgmongo.NewMockCollection(t)
			},
			mockSetup: func(ctx context.Context, t *testing.T, deps mockDeps, mockCol *pkgmongo.MockCollection) {
				// Use a different year to test year collection calculation
				differentYearTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC) // Next year
				expectedID := primitive.NewObjectIDFromTimestamp(differentYearTime)
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(expectedID)
				collectionName := fmt.Sprintf("%s_%d_%d", "events", year, period)

				deps.db.EXPECT().Collection(collectionName).Return(mockCol)
				mockCol.EXPECT().InsertOne(ctx, mock.Anything).Return(nil, nil)
			},
			scope: scope,
			input: repository.CreateOptions{
				Title:      "Next Year Event",
				StartTime:  time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
				EndTime:    time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC),
				TimezoneID: timeZoneID.Hex(),
				AllDay:     false,
			},
			wantErr: false,
			validate: func(t *testing.T, event models.Event) {
				differentYearTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
				// Verify ID is based on the next year date
				expectedID := primitive.NewObjectIDFromTimestamp(differentYearTime)
				require.Equal(t, expectedID.Timestamp().Unix(), event.ID.Timestamp().Unix())

				// Verify period calculation
				period, year := pkgmongo.GetPeriodAndYearFromObjectID(event.ID)
				require.Equal(t, int32(1), period)  // Should be Q1
				require.Equal(t, int32(2025), year) // Should be 2025
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)
			mockCol := tc.mockCol(t)

			if tc.mockSetup != nil {
				tc.mockSetup(ctx, t, deps, mockCol)
			}

			gotEvent, gotErr := repo.Create(ctx, tc.scope, tc.input)

			if tc.wantErr {
				require.Error(t, gotErr)
				require.Contains(t, gotErr.Error(), tc.errMessage)
				require.Equal(t, models.Event{}, gotEvent) // Should return empty event on error
				return
			}

			require.NoError(t, gotErr)

			// Run test-specific validation logic if provided
			if tc.validate != nil {
				tc.validate(t, gotEvent)
			}

			// Common validations
			require.NotEmpty(t, gotEvent.ID)
			require.Equal(t, mockTime.UTC(), gotEvent.CreatedAt)
			require.Equal(t, mockTime.UTC(), gotEvent.UpdatedAt)
		})
	}
}

func TestEventRepository_Detail(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
	scope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     mockShopID.Hex(),
		ShopPrefix: "t",
	})
	mockEvent := models.Event{
		ID:        mockID,
		Title:     "Meeting Event A",
		ShopID:    mockShopID,
		CreatedAt: mockTime,
		UpdatedAt: mockTime,
	}

	tcs := map[string]struct {
		input      string
		setupMocks func(ctx context.Context, deps *mockDeps)
		wantRes    models.Event
		wantErr    error
		isErrEvent bool
	}{
		"success": {
			input: mockID.Hex(),
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
				result := pkgmongo.NewMockSingleResult(t)
				result.On("Decode", mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						event := args[0].(*models.Event)
						*event = mockEvent
					})
				deps.col.EXPECT().FindOne(ctx, mock.MatchedBy(func(filter bson.M) bool {
					idMatch := reflect.DeepEqual(filter["_id"], mockID)
					shopMatch := reflect.DeepEqual(filter["shop_id"], mockShopID)
					deletedMatch := filter["deleted_at"] == nil
					return idMatch && shopMatch && deletedMatch
				})).Return(result)
			},
			wantRes:    mockEvent,
			wantErr:    nil,
			isErrEvent: false,
		},
		"invalid id": {
			input: "invalid-id",
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				// Vì hàm Detail vẫn chạy GetPeriodAndYearFromObjectID với một ObjectID không hợp lệ
				// Nên chúng ta cần mock Collection để nó không gây lỗi
				dummyID, _ := primitive.ObjectIDFromHex("000000000000000000000000")
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(dummyID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)

				// Không cần mock FindOne vì code sẽ return error trước khi gọi đến nó
			},
			wantRes:    models.Event{},
			wantErr:    primitive.ErrInvalidHex,
			isErrEvent: true,
		},
		"event not found": {
			input: mockID.Hex(),
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
				result := pkgmongo.NewMockSingleResult(t)
				result.On("Decode", mock.Anything).Return(mongo.ErrNoDocuments)
				deps.col.EXPECT().FindOne(ctx, mock.MatchedBy(func(filter bson.M) bool {
					idMatch := reflect.DeepEqual(filter["_id"], mockID)
					shopMatch := reflect.DeepEqual(filter["shop_id"], mockShopID)
					deletedMatch := filter["deleted_at"] == nil
					return idMatch && shopMatch && deletedMatch
				})).Return(result)
			},
			wantRes:    models.Event{},
			wantErr:    mongo.ErrNoDocuments,
			isErrEvent: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)
			tc.setupMocks(ctx, &deps)

			gotEvent, gotErr := repo.Detail(ctx, scope, tc.input)

			if tc.isErrEvent {
				require.Error(t, gotErr)
				require.Equal(t, tc.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			require.Equal(t, tc.wantRes, gotEvent)
		})
	}
}

func TestEventRepository_Update(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockCategoryID, _ := primitive.ObjectIDFromHex("111122223333444455556677")
	mockTimezoneID, _ := primitive.ObjectIDFromHex("333344445555666677778888")
	mockObjectID, _ := primitive.ObjectIDFromHex("222233334444555566667777")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
	scope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     mockShopID.Hex(),
		ShopPrefix: "t",
	})

	updatedTitle := "Updated Meeting Event A"
	updatedDescription := "Updated description"
	updatedRoomIDs := []string{"888899990000111122223333"}
	updatedBranchIDs := []string{"444455556666777788889999"}
	updatedDepartmentIDs := []string{"555566667777888899990000"}
	updatedAssignIDs := []string{"user-1", "user-2"}

	baseEvent := models.Event{
		ID:        mockID,
		Title:     updatedTitle,
		ShopID:    mockShopID,
		CreatedAt: mockTime,
		UpdatedAt: mockTime,
	}

	alert := &models.DateConfig{
		Num:  15,
		Unit: models.DateUnitMinute,
	}

	tcs := map[string]struct {
		input      repository.UpdateOptions
		scope      models.Scope
		setupMocks func(ctx context.Context, deps *mockDeps)
		validate   func(t *testing.T, event models.Event)
		wantErr    bool
		errMessage string
	}{
		"success_basic_update": {
			input: repository.UpdateOptions{
				ID:    mockID.Hex(),
				Title: updatedTitle,
				Model: baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)

				deps.col.EXPECT().UpdateOne(ctx, mock.MatchedBy(func(filter bson.M) bool {
					idMatch := reflect.DeepEqual(filter["_id"], mockID)
					shopMatch := reflect.DeepEqual(filter["shop_id"], mockShopID)
					deletedMatch := filter["deleted_at"] == nil
					return idMatch && shopMatch && deletedMatch
				}), mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
			},
			validate: func(t *testing.T, event models.Event) {
				require.Equal(t, mockID, event.ID)
				require.Equal(t, updatedTitle, event.Title)
				require.Equal(t, mockShopID, event.ShopID)
			},
			wantErr: false,
		},
		"success_full_update": {
			input: repository.UpdateOptions{
				ID:            mockID.Hex(),
				Title:         updatedTitle,
				AssignIDs:     updatedAssignIDs,
				BranchIDs:     updatedBranchIDs,
				DepartmentIDs: updatedDepartmentIDs,
				TimezoneID:    mockTimezoneID.Hex(),
				StartTime:     mockTime.Add(time.Hour),
				EndTime:       mockTime.Add(2 * time.Hour),
				AllDay:        true,
				Repeat:        models.EventRepeatWeekly,
				RoomIDs:       updatedRoomIDs,
				Description:   updatedDescription,
				CategoryID:    mockCategoryID.Hex(),
				Notify:        true,
				Alert:         alert,
				ObjectID:      mockObjectID.Hex(),
				Model:         baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)

				deps.col.EXPECT().UpdateOne(ctx, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
			},
			validate: func(t *testing.T, event models.Event) {
				require.Equal(t, mockID, event.ID)
				require.Equal(t, updatedTitle, event.Title)
				require.Equal(t, updatedDescription, event.Description)
				require.Equal(t, updatedAssignIDs, event.AssignIDs)
				require.Equal(t, true, event.AllDay)
				require.Equal(t, models.EventRepeatWeekly, event.Repeat)
				require.Equal(t, mockTimezoneID, event.TimezoneID)
				require.Equal(t, mockCategoryID, *event.CategoryID)
				require.Equal(t, mockObjectID, *event.ObjectID)
				require.Equal(t, alert.Num, event.Alert.Num)
				require.Equal(t, alert.Unit, event.Alert.Unit)

				// Verify ObjectIDs were properly converted from strings
				roomIDs, _ := pkgmongo.ObjectIDsFromHexs(updatedRoomIDs)
				branchIDs, _ := pkgmongo.ObjectIDsFromHexs(updatedBranchIDs)
				departmentIDs, _ := pkgmongo.ObjectIDsFromHexs(updatedDepartmentIDs)

				require.Equal(t, roomIDs, event.RoomIDs)
				require.Equal(t, branchIDs, event.BranchIDs)
				require.Equal(t, departmentIDs, event.DepartmentIDs)
			},
			wantErr: false,
		},
		"success_with_unset": {
			input: repository.UpdateOptions{
				ID:            mockID.Hex(),
				Title:         updatedTitle,
				AssignIDs:     []string{},
				BranchIDs:     []string{},
				DepartmentIDs: []string{},
				RoomIDs:       []string{},
				Description:   "",
				Model:         baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)

				deps.col.EXPECT().UpdateOne(ctx, mock.Anything, mock.MatchedBy(func(u bson.M) bool {
					// Verify that $unset exists in the update
					_, hasUnset := u["$unset"]
					return hasUnset
				})).Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
			},
			validate: func(t *testing.T, event models.Event) {
				require.Equal(t, mockID, event.ID)
				require.Equal(t, updatedTitle, event.Title)
				require.Empty(t, event.Description)
				require.Nil(t, event.AssignIDs)
				require.Nil(t, event.BranchIDs)
				require.Nil(t, event.DepartmentIDs)
				require.Nil(t, event.RoomIDs)
			},
			wantErr: false,
		},
		"invalid_id": {
			input: repository.UpdateOptions{
				ID:    "invalid-id",
				Title: updatedTitle,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				dummyID, _ := primitive.ObjectIDFromHex("000000000000000000000000")
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(dummyID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: "the provided hex string is not a valid ObjectID",
		},
		"document_not_found": {
			input: repository.UpdateOptions{
				ID:    mockID.Hex(),
				Title: updatedTitle,
				Model: baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)

				deps.col.EXPECT().UpdateOne(ctx, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}, pkgmongo.ErrNoDocuments)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: "mongo: no documents in result",
		},
		"invalid_category_id": {
			input: repository.UpdateOptions{
				ID:         mockID.Hex(),
				Title:      updatedTitle,
				CategoryID: "invalid-category-id",
				Model:      baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				// Even though we expect to fail before database call, the code might still try to access the collection
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: errors.New("the provided hex string is not a valid ObjectID").Error(),
		},
		"invalid_branch_ids": {
			input: repository.UpdateOptions{
				ID:        mockID.Hex(),
				Title:     updatedTitle,
				BranchIDs: []string{"invalid-branch-id"},
				Model:     baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				// Even though we expect to fail before database call, the code might still try to access the collection
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: pkgmongo.ErrInvalidObjectID.Error(),
		},
		"error_invalid_timezone_id": {
			input: repository.UpdateOptions{
				ID:         mockID.Hex(),
				Title:      updatedTitle,
				TimezoneID: "invalid-timezone-id",
				Model:      baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: "the provided hex string is not a valid ObjectID",
		},

		"error_invalid_department_ids": {
			input: repository.UpdateOptions{
				ID:            mockID.Hex(),
				Title:         updatedTitle,
				DepartmentIDs: []string{"invalid-department-id"},
				Model:         baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: pkgmongo.ErrInvalidObjectID.Error(),
		},

		"error_invalid_room_ids": {
			input: repository.UpdateOptions{
				ID:      mockID.Hex(),
				Title:   updatedTitle,
				RoomIDs: []string{"invalid-room-id"},
				Model:   baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: pkgmongo.ErrInvalidObjectID.Error(),
		},

		"error_invalid_object_id": {
			input: repository.UpdateOptions{
				ID:       mockID.Hex(),
				Title:    updatedTitle,
				ObjectID: "invalid-object-id",
				Model:    baseEvent,
			},
			scope: scope,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			validate:   nil,
			wantErr:    true,
			errMessage: "the provided hex string is not a valid ObjectID",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			if tc.setupMocks != nil {
				tc.setupMocks(ctx, &deps)
			}

			gotEvent, gotErr := repo.Update(ctx, tc.scope, tc.input)

			if tc.wantErr {
				require.Error(t, gotErr)
				require.Contains(t, gotErr.Error(), tc.errMessage)
				require.Equal(t, models.Event{}, gotEvent)
				return
			}

			require.NoError(t, gotErr)

			if tc.validate != nil {
				tc.validate(t, gotEvent)
			}
		})
	}
}

func TestEventRepository_Delete(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     mockShopID.Hex(),
		ShopPrefix: "t",
	})

	invalidShopIDScope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     "invalid-shop-id",
		ShopPrefix: "t",
	})

	type mockDeleteMany struct {
		isCalled bool
		filter   bson.M
		result   int64
		err      error
	}

	tcs := map[string]struct {
		input      string
		scope      models.Scope
		mockColl   mockDeleteMany
		shouldMock bool
		wantErr    error
	}{
		"success": {
			input: mockID.Hex(),
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: true,
				filter: bson.M{
					"_id":        mockID,
					"shop_id":    mockShopID,
					"deleted_at": nil,
				},
				result: 1,
				err:    nil,
			},
			shouldMock: true,
			wantErr:    nil,
		},
		"error_deleting": {
			input: mockID.Hex(),
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: true,
				filter: bson.M{
					"_id":        mockID,
					"shop_id":    mockShopID,
					"deleted_at": nil,
				},
				result: 0,
				err:    pkgmongo.ErrNoDocuments,
			},
			shouldMock: true,
			wantErr:    pkgmongo.ErrNoDocuments,
		},
		"invalid_id": {
			input: "invalid-id",
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: false,
			wantErr:    pkgmongo.ErrInvalidObjectID,
		},
		"invalid_shop_id": {
			input: mockID.Hex(),
			scope: invalidShopIDScope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: true,
			wantErr:    primitive.ErrInvalidHex,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			repo, deps := initRepo(t, mockTime)

			if tc.shouldMock {
				deps.db.EXPECT().Collection(mock.AnythingOfType("string")).Return(deps.col)
			}

			if tc.mockColl.isCalled {
				deps.col.EXPECT().DeleteSoftOne(ctx, tc.mockColl.filter).
					Return(tc.mockColl.result, tc.mockColl.err)
			}

			gotErr := repo.Delete(ctx, tc.scope, tc.input)

			if tc.wantErr != nil {
				require.Error(t, gotErr)
				require.Equal(t, tc.wantErr, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}

func TestEventRepository_UpdateRepeatUntil(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	scope := models.Scope{
		ShopID: mockShopID.Hex(),
	}
	repeatUntil := fixedTime.Add(24 * time.Hour)

	tcs := map[string]struct {
		id         string
		repeatTime time.Time
		setupMocks func(ctx context.Context, deps *mockDeps)
		wantErr    error
		isErr      bool
	}{
		"success": {
			id:         mockID.Hex(),
			repeatTime: repeatUntil,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
				deps.col.EXPECT().UpdateOne(
					ctx,
					bson.M{
						"shop_id":    mockShopID,
						"deleted_at": nil,
						"_id":        mockID,
					},
					mock.MatchedBy(func(update bson.M) bool {
						set, ok := update["$set"].(bson.M)
						if !ok {
							return false
						}
						// So sánh repeat_until chính xác
						if !set["repeat_until"].(time.Time).Equal(repeatUntil) {
							return false
						}
						// So sánh updated_at gần với thời điểm hiện tại (trong 2s)
						updatedAt, ok := set["updated_at"].(time.Time)
						if !ok {
							return false
						}
						return time.Since(updatedAt) < 2*time.Second
					}),
				).Return(nil, nil)
			},
			wantErr: nil,
			isErr:   false,
		},
		"invalid id": {
			id:         "invalid-id",
			repeatTime: repeatUntil,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				// Không cần mock gì vì sẽ lỗi ngay từ đầu
			},
			wantErr: primitive.ErrInvalidHex,
			isErr:   true,
		},
		"invalid shop id in scope": {
			id:         mockID.Hex(),
			repeatTime: repeatUntil,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				deps.db.EXPECT().Collection(fmt.Sprintf("events_%d_%d", y, p)).Return(deps.col)
			},
			wantErr: primitive.ErrInvalidHex,
			isErr:   true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, fixedTime)
			// Sửa scope cho case invalid shop id
			testScope := scope
			if name == "invalid shop id in scope" {
				testScope = models.Scope{ShopID: "invalid-shop-id"}
			}
			tc.setupMocks(ctx, &deps)

			_, gotErr := repo.UpdateRepeatUntil(ctx, testScope, tc.id, tc.repeatTime)

			if tc.isErr {
				require.Error(t, gotErr)
				if tc.wantErr != nil {
					require.Equal(t, tc.wantErr, gotErr)
				}
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestEventRepository_ListByIDs(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockID1, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockID2, _ := primitive.ObjectIDFromHex("667788990011223344556678")
	mockID3, _ := primitive.ObjectIDFromHex("667788990011223344556679")

	// Create invalid ID for testing error case
	invalidID := "invalid-id"

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	tcs := map[string]struct {
		ids        []string
		setupMocks func(ctx context.Context, deps *mockDeps)
		wantEvents []*models.Event
		wantErr    bool
	}{
		"success with multiple collections": {
			ids: []string{mockID1.Hex(), mockID2.Hex(), mockID3.Hex()},
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				// Giả sử cả 3 ID cùng collection:
				p1, y1 := pkgmongo.GetPeriodAndYearFromObjectID(mockID1)
				colName1 := fmt.Sprintf("events_%d_%d", y1, p1)
				deps.db.EXPECT().Collection(colName1).Return(deps.col)

				deps.col.EXPECT().Find(
					ctx,
					mock.MatchedBy(func(filter bson.M) bool {
						ids, ok := filter["_id"].(bson.M)["$in"].([]primitive.ObjectID)
						return ok && len(ids) == 3
					}),
				).Return(deps.cur, nil)

				events := []models.Event{
					{ID: mockID1, ShopID: mockShopID, Title: "Event 1"},
					{ID: mockID2, ShopID: mockShopID, Title: "Event 2"},
					{ID: mockID3, ShopID: mockShopID, Title: "Event 3"},
				}

				deps.cur.EXPECT().All(ctx, mock.AnythingOfType("*[]models.Event")).
					Run(func(ctx context.Context, result interface{}) {
						evs := result.(*[]models.Event)
						*evs = events
					}).Return(nil)
			},
			wantEvents: []*models.Event{
				{ID: mockID1, ShopID: mockShopID, Title: "Event 1"},
				{ID: mockID2, ShopID: mockShopID, Title: "Event 2"},
				{ID: mockID3, ShopID: mockShopID, Title: "Event 3"},
			},
			wantErr: false,
		},
		"error with invalid ID": {
			ids: []string{invalidID},
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				// No mocks needed as it should fail on ObjectIDFromHex
			},
			wantEvents: nil,
			wantErr:    true,
		},
		"error with find operation": {
			ids: []string{mockID1.Hex()},
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID1)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)

				deps.col.EXPECT().Find(
					ctx,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["shop_id"] == mockShopID && filter["deleted_at"] == nil
					}),
				).Return(nil, fmt.Errorf("database error"))
			},
			wantEvents: nil,
			wantErr:    true,
		},
		"error with cursor all": {
			ids: []string{mockID1.Hex()},
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID1)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)

				deps.col.EXPECT().Find(
					ctx,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["shop_id"] == mockShopID && filter["deleted_at"] == nil
					}),
				).Return(deps.cur, nil)

				deps.cur.EXPECT().All(ctx, mock.AnythingOfType("*[]models.Event")).
					Return(fmt.Errorf("cursor error"))
			},
			wantEvents: nil,
			wantErr:    true,
		},
		"error with scope": {
			ids: []string{mockID1.Hex()},
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				scope.ShopID = "invalid-shop-id"
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID1)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
				deps.col.EXPECT().Find(
					ctx,
					mock.Anything, // hoặc mock.MatchedBy(...) nếu muốn chặt chẽ hơn
				).Return(nil, fmt.Errorf("invalid shop id"))
			},
			wantEvents: nil,
			wantErr:    true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, fixedTime)

			// Reset scope for each test
			testScope := models.Scope{
				ShopID: mockShopID.Hex(),
				UserID: "user123",
			}

			// Set up mocks
			tc.setupMocks(ctx, &deps)

			// Test the function
			gotEvents, gotErr := repo.ListByIDs(ctx, testScope, tc.ids)

			if tc.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			require.Equal(t, len(tc.wantEvents), len(gotEvents))

			// Check each event matches the expected one
			for i, want := range tc.wantEvents {
				require.Equal(t, want.ID, gotEvents[i].ID)
				require.Equal(t, want.ShopID, gotEvents[i].ShopID)
				require.Equal(t, want.Title, gotEvents[i].Title)
			}
		})
	}
}

func TestEventRepository_UpdateAttendance(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockID, _ := primitive.ObjectIDFromHex("667788990011223344556677")

	tcs := map[string]struct {
		eventID    string
		status     int
		setupMocks func(ctx context.Context, deps *mockDeps)
		wantErr    bool
	}{
		"success with accept status": {
			eventID: mockID.Hex(),
			status:  1, // Accept
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)

				deps.col.EXPECT().UpdateOne(
					ctx,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["shop_id"] == mockShopID &&
							filter["deleted_at"] == nil &&
							filter["_id"] == mockID
					}),
					mock.MatchedBy(func(update bson.M) bool {
						addToSet, ok1 := update["$addToSet"].(bson.M)
						pull, ok2 := update["$pull"].(bson.M)

						if !ok1 || !ok2 {
							return false
						}

						// Check if the right fields are updated
						return addToSet["accepted_ids"] == "user123" &&
							pull["declined_ids"] == "user123" &&
							update["updated_at"] != nil
					}),
				).Return(&mongo.UpdateResult{}, nil)
			},
			wantErr: false,
		},
		"success with decline status": {
			eventID: mockID.Hex(),
			status:  -1, // Decline
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)

				deps.col.EXPECT().UpdateOne(
					ctx,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["shop_id"] == mockShopID &&
							filter["deleted_at"] == nil &&
							filter["_id"] == mockID
					}),
					mock.MatchedBy(func(update bson.M) bool {
						addToSet, ok1 := update["$addToSet"].(bson.M)
						pull, ok2 := update["$pull"].(bson.M)

						if !ok1 || !ok2 {
							return false
						}

						// Check if the right fields are updated
						return addToSet["declined_ids"] == "user123" &&
							pull["accepted_ids"] == "user123" &&
							update["updated_at"] != nil
					}),
				).Return(&mongo.UpdateResult{}, nil)
			},
			wantErr: false,
		},
		"error with invalid event ID": {
			eventID: "invalid-id",
			status:  1,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				// No mocks needed as it should fail on ObjectIDFromHex
			},
			wantErr: true,
		},
		"error with update operation": {
			eventID: mockID.Hex(),
			status:  1,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)

				deps.col.EXPECT().UpdateOne(
					ctx,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["shop_id"] == mockShopID &&
							filter["deleted_at"] == nil &&
							filter["_id"] == mockID
					}),
					mock.Anything,
				).Return(nil, fmt.Errorf("database error"))
			},
			wantErr: true,
		},
		"error with scope": {
			eventID: mockID.Hex(),
			status:  1,
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
				deps.col.EXPECT().UpdateOne(
					ctx,
					mock.Anything,
					mock.Anything,
				).Return(nil, primitive.ErrInvalidHex)
			},
			wantErr: true,
		},
		"neutral status (no update)": {
			eventID: mockID.Hex(),
			status:  0, // Neutral - no update should happen
			setupMocks: func(ctx context.Context, deps *mockDeps) {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockID)
				colName := fmt.Sprintf("events_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)

				deps.col.EXPECT().UpdateOne(
					ctx,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["shop_id"] == mockShopID &&
							filter["deleted_at"] == nil &&
							filter["_id"] == mockID
					}),
					mock.MatchedBy(func(update bson.M) bool {
						// Should be empty update since status is 0
						return len(update) == 0
					}),
				).Return(&mongo.UpdateResult{}, nil)
			},
			wantErr: false,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, fixedTime)

			// Reset scope for each test
			testScope := models.Scope{
				ShopID: mockShopID.Hex(),
				UserID: "user123",
			}

			// Set up mocks
			tc.setupMocks(ctx, &deps)

			// Test the function
			gotErr := repo.UpdateAttendance(ctx, testScope, tc.eventID, tc.status)

			if tc.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}

func TestEventRepository_DeleteNextRecurringInstances(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
	specificTime := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	scope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     mockShopID.Hex(),
		ShopPrefix: "t",
	})

	invalidShopIDScope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     "invalid-shop-id",
		ShopPrefix: "t",
	})

	type mockDeleteMany struct {
		isCalled bool
		filter   bson.M
		result   int64
		err      error
	}

	tcs := map[string]struct {
		input      repository.DeleteNextRecurringInstancesOptions
		scope      models.Scope
		mockColl   mockDeleteMany
		shouldMock bool
		wantErr    error
	}{
		"success": {
			input: repository.DeleteNextRecurringInstancesOptions{
				EventID: mockEventID.Hex(),
				Date:    specificTime,
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: true,
				filter: bson.M{
					"event_id": mockEventID,
					"shop_id":  mockShopID,
					"start_time": bson.M{
						"$gt": specificTime,
					},
					"deleted_at": nil,
				},
				result: 1,
				err:    nil,
			},
			shouldMock: true,
			wantErr:    nil,
		},
		"error_invalid_eventid": {
			input: repository.DeleteNextRecurringInstancesOptions{
				EventID: "invalid-id",
				Date:    specificTime,
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: false,
			wantErr:    primitive.ErrInvalidHex,
		},
		"error_missing_date": {
			input: repository.DeleteNextRecurringInstancesOptions{
				EventID: mockEventID.Hex(),
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: true,
			wantErr:    errors.New("required field"),
		},
		"error_missing_eventid": {
			input: repository.DeleteNextRecurringInstancesOptions{
				Date: specificTime,
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: false,
			wantErr:    primitive.ErrInvalidHex,
		},
		"error_deleting": {
			input: repository.DeleteNextRecurringInstancesOptions{
				EventID: mockEventID.Hex(),
				Date:    specificTime,
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: true,
				filter: bson.M{
					"event_id": mockEventID,
					"shop_id":  mockShopID,
					"start_time": bson.M{
						"$gt": specificTime,
					},
					"deleted_at": nil,
				},
				result: 0,
				err:    pkgmongo.ErrNoDocuments,
			},
			shouldMock: true,
			wantErr:    pkgmongo.ErrNoDocuments,
		},
		"invalid_shop_id": {
			input: repository.DeleteNextRecurringInstancesOptions{
				EventID: mockEventID.Hex(),
				Date:    specificTime,
			},
			scope: invalidShopIDScope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: true,
			wantErr:    primitive.ErrInvalidHex,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			repo, deps := initRepo(t, mockTime)

			if tc.shouldMock {
				deps.db.EXPECT().Collection(mock.AnythingOfType("string")).Return(deps.col)
			}

			if tc.mockColl.isCalled {
				deps.col.EXPECT().DeleteSoftMany(ctx, tc.mockColl.filter).
					Return(tc.mockColl.result, tc.mockColl.err)
			}

			gotErr := repo.DeleteNextRecurringInstances(ctx, tc.scope, tc.input)

			if tc.wantErr != nil {
				require.Error(t, gotErr)
				require.Equal(t, tc.wantErr, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}

func TestEventRepository_DeleteRecurringInstance(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockInstanceID, _ := primitive.ObjectIDFromHex("112233445566778899001122")
	mockInstanceID2, _ := primitive.ObjectIDFromHex("112233445566778899001123")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     mockShopID.Hex(),
		ShopPrefix: "t",
	})

	invalidShopIDScope := jwt.NewScope(jwt.Payload{
		UserID:     "user-id",
		ShopID:     "invalid-shop-id",
		ShopPrefix: "t",
	})

	type mockDeleteMany struct {
		isCalled bool
		filter   bson.M
		result   int64
		err      error
	}

	tcs := map[string]struct {
		input      repository.DeleteRecurringInstanceOptions
		scope      models.Scope
		mockColl   mockDeleteMany
		shouldMock bool
		wantErr    error
	}{
		"success_with_ids": {
			input: repository.DeleteRecurringInstanceOptions{
				IDs:     []string{mockInstanceID.Hex(), mockInstanceID2.Hex()},
				EventID: mockEventID.Hex(),
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: true,
				filter: bson.M{
					"_id": bson.M{
						"$in": []primitive.ObjectID{mockInstanceID, mockInstanceID2},
					},
					"event_id":   mockEventID,
					"shop_id":    mockShopID,
					"deleted_at": nil,
				},
				result: 2,
				err:    nil,
			},
			shouldMock: true,
			wantErr:    nil,
		},
		"success_with_eventid_only": {
			input: repository.DeleteRecurringInstanceOptions{
				EventID: mockEventID.Hex(),
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: true,
				filter: bson.M{
					"event_id":   mockEventID,
					"shop_id":    mockShopID,
					"deleted_at": nil,
				},
				result: 3, // Giả sử xóa 3 instances
				err:    nil,
			},
			shouldMock: true,
			wantErr:    nil,
		},
		"error_invalid_eventid": {
			input: repository.DeleteRecurringInstanceOptions{
				EventID: "invalid-id",
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: false,
			wantErr:    primitive.ErrInvalidHex,
		},
		"error_invalid_instance_id": {
			input: repository.DeleteRecurringInstanceOptions{
				EventID: mockEventID.Hex(),
				IDs:     []string{"invalid-id"},
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: true,
			wantErr:    pkgmongo.ErrInvalidObjectID,
		},
		"error_deleting": {
			input: repository.DeleteRecurringInstanceOptions{
				EventID: mockEventID.Hex(),
			},
			scope: scope,
			mockColl: mockDeleteMany{
				isCalled: true,
				filter: bson.M{
					"event_id":   mockEventID,
					"shop_id":    mockShopID,
					"deleted_at": nil,
				},
				result: 0,
				err:    pkgmongo.ErrNoDocuments,
			},
			shouldMock: true,
			wantErr:    pkgmongo.ErrNoDocuments,
		},
		"invalid_shop_id": {
			input: repository.DeleteRecurringInstanceOptions{
				EventID: mockEventID.Hex(),
			},
			scope: invalidShopIDScope,
			mockColl: mockDeleteMany{
				isCalled: false,
			},
			shouldMock: true,
			wantErr:    primitive.ErrInvalidHex,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			repo, deps := initRepo(t, mockTime)

			if tc.shouldMock {
				deps.db.EXPECT().Collection(mock.AnythingOfType("string")).Return(deps.col)
			}

			if tc.mockColl.isCalled {
				deps.col.EXPECT().DeleteSoftMany(ctx, tc.mockColl.filter).
					Return(tc.mockColl.result, tc.mockColl.err)
			}

			gotErr := repo.DeleteRecurringInstance(ctx, tc.scope, tc.input)

			if tc.wantErr != nil {
				require.Error(t, gotErr)
				require.Equal(t, tc.wantErr, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}

func TestEventRepository_CreateRecurringTracking(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID, _ := primitive.ObjectIDFromHex("667788990011223344556688")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	type mockInsertOne struct {
		isCalled bool
		doc      interface{}
		err      error
	}

	tcs := map[string]struct {
		opt        repository.CreateRecurringTrackingOptions
		scope      models.Scope
		mockInsert mockInsertOne
		wantErr    bool
	}{
		"success": {
			opt: repository.CreateRecurringTrackingOptions{
				EventID: mockEventID.Hex(),
				Month:   5,
				Year:    2024,
				Repeat:  models.EventRepeatMonthly,
				StartEndTime: []repository.StartEndTime{
					{StartTime: mockTime, EndTime: mockTime.Add(2 * time.Hour)},
				},
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
		"error_invalid_eventid": {
			opt: repository.CreateRecurringTrackingOptions{
				EventID: "invalid-id",
				Month:   5,
				Year:    2024,
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_insert": {
			opt: repository.CreateRecurringTrackingOptions{
				EventID: mockEventID.Hex(),
				Month:   5,
				Year:    2024,
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: true,
				err:      fmt.Errorf("insert error"),
			},
			wantErr: true,
		},
		"success_with_repeat_until": {
			opt: repository.CreateRecurringTrackingOptions{
				EventID:     mockEventID.Hex(),
				Month:       5,
				Year:        2024,
				Repeat:      models.EventRepeatMonthly,
				RepeatUntil: ptrTime(mockTime.Add(24 * time.Hour)),
				StartEndTime: []repository.StartEndTime{
					{StartTime: mockTime, EndTime: mockTime.Add(2 * time.Hour)},
				},
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			deps.db.EXPECT().Collection("recurring_trackings").Return(deps.col)
			deps.db.EXPECT().NewObjectID().Return(primitive.NewObjectID())

			if tc.mockInsert.isCalled {
				deps.col.EXPECT().InsertOne(ctx, mock.AnythingOfType("models.RecurringTracking")).Return(nil, tc.mockInsert.err)
			}

			_, err := repo.CreateRecurringTracking(ctx, tc.scope, tc.opt)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventRepository_GetGenRTsInDateRange(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
	fromTime := mockTime
	toTime := mockTime.Add(24 * time.Hour)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	invalidScope := models.Scope{
		ShopID: "invalid-id",
		UserID: "user123",
	}

	type mockFindAll struct {
		findErr error
		allErr  error
		results []models.RecurringTracking
	}

	tcs := map[string]struct {
		scope      models.Scope
		from, to   time.Time
		mockFind   mockFindAll
		wantErr    bool
		wantResult []models.RecurringTracking
	}{
		"success": {
			scope: scope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: nil,
				allErr:  nil,
				results: []models.RecurringTracking{
					{ID: primitive.NewObjectID(), ShopID: mockShopID, Month: 5, Year: 2024},
				},
			},
			wantErr:    false,
			wantResult: []models.RecurringTracking{{ShopID: mockShopID, Month: 5, Year: 2024}},
		},
		"error_build_query": {
			scope: invalidScope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: nil,
			},
			wantErr: true,
		},
		"error_find": {
			scope: scope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: fmt.Errorf("find error"),
			},
			wantErr: true,
		},
		"error_all": {
			scope: scope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: nil,
				allErr:  fmt.Errorf("all error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			deps.db.EXPECT().Collection("recurring_trackings").Return(deps.col)

			if tc.mockFind.findErr != nil {
				deps.col.EXPECT().Find(ctx, mock.Anything).Return(nil, tc.mockFind.findErr)
			} else if tc.scope.ShopID != "invalid-id" {
				deps.col.EXPECT().Find(ctx, mock.Anything).Return(deps.cur, nil)
				if tc.mockFind.allErr != nil {
					deps.cur.EXPECT().All(ctx, mock.AnythingOfType("*[]models.RecurringTracking")).Return(tc.mockFind.allErr)
				} else {
					deps.cur.EXPECT().All(ctx, mock.AnythingOfType("*[]models.RecurringTracking")).
						Run(func(ctx context.Context, result interface{}) {
							ptr := result.(*[]models.RecurringTracking)
							*ptr = tc.mockFind.results
						}).Return(nil)
				}
			}

			got, err := repo.GetGenRTsInDateRange(ctx, tc.scope, tc.from, tc.to)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.wantResult), len(got))
				if len(got) > 0 {
					require.Equal(t, tc.wantResult[0].ShopID, got[0].ShopID)
					require.Equal(t, tc.wantResult[0].Month, got[0].Month)
					require.Equal(t, tc.wantResult[0].Year, got[0].Year)
				}
			}
		})
	}
}

func TestEventRepository_GetGenRTsNotInDateRange(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
	fromTime := mockTime
	toTime := mockTime.Add(24 * time.Hour)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}
	invalidScope := models.Scope{
		ShopID: "invalid-id",
		UserID: "user123",
	}

	type mockFindAll struct {
		findErr error
		allErr  error
		results []models.RecurringTracking
	}

	tcs := map[string]struct {
		scope      models.Scope
		from, to   time.Time
		mockFind   mockFindAll
		wantErr    bool
		wantResult []models.RecurringTracking
	}{
		"success": {
			scope: scope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: nil,
				allErr:  nil,
				results: []models.RecurringTracking{
					{ID: primitive.NewObjectID(), ShopID: mockShopID, Month: 5, Year: 2024},
				},
			},
			wantErr:    false,
			wantResult: []models.RecurringTracking{{ShopID: mockShopID, Month: 5, Year: 2024}},
		},
		"error_build_query": {
			scope: invalidScope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: nil,
			},
			wantErr: true,
		},
		"error_find": {
			scope: scope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: fmt.Errorf("find error"),
			},
			wantErr: true,
		},
		"error_all": {
			scope: scope,
			from:  fromTime,
			to:    toTime,
			mockFind: mockFindAll{
				findErr: nil,
				allErr:  fmt.Errorf("all error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			deps.db.EXPECT().Collection("recurring_trackings").Return(deps.col)

			if tc.mockFind.findErr != nil {
				deps.col.EXPECT().Find(ctx, mock.Anything).Return(nil, tc.mockFind.findErr)
			} else if tc.scope.ShopID != "invalid-id" {
				deps.col.EXPECT().Find(ctx, mock.Anything).Return(deps.cur, nil)
				if tc.mockFind.allErr != nil {
					deps.cur.EXPECT().All(ctx, mock.AnythingOfType("*[]models.RecurringTracking")).Return(tc.mockFind.allErr)
				} else {
					deps.cur.EXPECT().All(ctx, mock.AnythingOfType("*[]models.RecurringTracking")).
						Run(func(ctx context.Context, result interface{}) {
							ptr := result.(*[]models.RecurringTracking)
							*ptr = tc.mockFind.results
						}).Return(nil)
				}
			}

			got, err := repo.GetGenRTsNotInDateRange(ctx, tc.scope, tc.from, tc.to)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.wantResult), len(got))
				if len(got) > 0 {
					require.Equal(t, tc.wantResult[0].ShopID, got[0].ShopID)
					require.Equal(t, tc.wantResult[0].Month, got[0].Month)
					require.Equal(t, tc.wantResult[0].Year, got[0].Year)
				}
			}
		})
	}
}

func TestEventRepository_UpdateRepeatUntilRecurringTrackings(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID, _ := primitive.ObjectIDFromHex("667788990011223344556688")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)
	repeatUntil := mockTime.Add(30 * 24 * time.Hour)
	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}
	invalidScope := models.Scope{
		ShopID: "invalid-id",
		UserID: "user123",
	}

	type mockUpdateMany struct {
		isCalled bool
		filter   bson.M
		update   bson.M
		err      error
	}

	tcs := map[string]struct {
		opt        repository.UpdateRepeatUntilRecurringTrackingsOptions
		scope      models.Scope
		mockUpdate mockUpdateMany
		wantErr    bool
	}{
		"success": {
			opt: repository.UpdateRepeatUntilRecurringTrackingsOptions{
				EventID:     mockEventID.Hex(),
				RepeatUntil: &repeatUntil,
			},
			scope: scope,
			mockUpdate: mockUpdateMany{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
		"error_build_query": {
			opt: repository.UpdateRepeatUntilRecurringTrackingsOptions{
				EventID:     "invalid-id",
				RepeatUntil: &repeatUntil,
			},
			scope: scope,
			mockUpdate: mockUpdateMany{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_build_update": {
			opt: repository.UpdateRepeatUntilRecurringTrackingsOptions{
				EventID:     mockEventID.Hex(),
				RepeatUntil: nil, // Test with nil RepeatUntil
			},
			scope: scope,
			mockUpdate: mockUpdateMany{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
		"error_update_many": {
			opt: repository.UpdateRepeatUntilRecurringTrackingsOptions{
				EventID:     mockEventID.Hex(),
				RepeatUntil: &repeatUntil,
			},
			scope: scope,
			mockUpdate: mockUpdateMany{
				isCalled: true,
				err:      fmt.Errorf("update error"),
			},
			wantErr: true,
		},
		"error_scope": {
			opt: repository.UpdateRepeatUntilRecurringTrackingsOptions{
				EventID:     mockEventID.Hex(),
				RepeatUntil: &repeatUntil,
			},
			scope: invalidScope,
			mockUpdate: mockUpdateMany{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_empty_scope": {
			opt: repository.UpdateRepeatUntilRecurringTrackingsOptions{
				EventID:     mockEventID.Hex(),
				RepeatUntil: &repeatUntil,
			},
			scope: models.Scope{
				ShopID: "",
				UserID: "",
			},
			mockUpdate: mockUpdateMany{
				isCalled: true,                      // Changed to true since we expect the call
				err:      fmt.Errorf("empty scope"), // Add expected error
			},
			wantErr: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			if name == "error_nil_collection" {
				deps.db.EXPECT().Collection("recurring_trackings").Return(nil)
				// No UpdateMany expectation needed since we expect an error before that point
			} else {
				deps.db.EXPECT().Collection("recurring_trackings").Return(deps.col)
				if tc.mockUpdate.isCalled {
					deps.col.EXPECT().UpdateMany(ctx, mock.Anything, mock.Anything).Return(nil, tc.mockUpdate.err)
				}
			}

			err := repo.UpdateRepeatUntilRecurringTrackings(ctx, tc.scope, tc.opt)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventRepository_DeleteRecurringTracking(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID, _ := primitive.ObjectIDFromHex("667788990011223344556688")
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}
	invalidScope := models.Scope{
		ShopID: "invalid-id",
		UserID: "user123",
	}

	type mockDeleteMany struct {
		isCalled bool
		filter   bson.M
		result   int64
		err      error
	}

	tcs := map[string]struct {
		opt        repository.DeleteRecurringTrackingOptions
		scope      models.Scope
		mockDelete mockDeleteMany
		wantErr    bool
	}{
		"success": {
			opt: repository.DeleteRecurringTrackingOptions{
				EventID: mockEventID.Hex(),
			},
			scope: scope,
			mockDelete: mockDeleteMany{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
		"error_build_query": {
			opt: repository.DeleteRecurringTrackingOptions{
				EventID: "invalid-id",
			},
			scope: scope,
			mockDelete: mockDeleteMany{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_delete": {
			opt: repository.DeleteRecurringTrackingOptions{
				EventID: mockEventID.Hex(),
			},
			scope: scope,
			mockDelete: mockDeleteMany{
				isCalled: true,
				err:      fmt.Errorf("delete error"),
			},
			wantErr: true,
		},
		"error_scope": {
			opt: repository.DeleteRecurringTrackingOptions{
				EventID: mockEventID.Hex(),
			},
			scope: invalidScope,
			mockDelete: mockDeleteMany{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_eventid": {
			opt: repository.DeleteRecurringTrackingOptions{
				EventID: "invalid-object-id",
			},
			scope: scope,
			mockDelete: mockDeleteMany{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_ids": {
			opt: repository.DeleteRecurringTrackingOptions{
				IDs: []string{"invalid-object-id"},
			},
			scope: scope,
			mockDelete: mockDeleteMany{
				isCalled: false,
			},
			wantErr: true,
		},
		"success_with_ids": {
			opt: repository.DeleteRecurringTrackingOptions{
				IDs: []string{
					primitive.NewObjectID().Hex(),
					primitive.NewObjectID().Hex(),
				},
			},
			scope: scope,
			mockDelete: mockDeleteMany{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
		"success_with_all_filters": {
			opt: repository.DeleteRecurringTrackingOptions{
				IDs:     []string{primitive.NewObjectID().Hex()},
				EventID: primitive.NewObjectID().Hex(),
				Month:   ptrInt32(6),
				Year:    ptrInt32(2024),
			},
			scope: scope,
			mockDelete: mockDeleteMany{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			deps.db.EXPECT().Collection("recurring_trackings").Return(deps.col)

			if tc.mockDelete.isCalled {
				deps.col.EXPECT().DeleteSoftMany(ctx, mock.Anything).Return(tc.mockDelete.result, tc.mockDelete.err)
			}

			err := repo.DeleteRecurringTracking(ctx, tc.scope, tc.opt)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventRepository_CreateRecurringInstance(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID := primitive.NewObjectID()
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	type mockInsertOne struct {
		isCalled bool
		err      error
	}

	tcs := map[string]struct {
		opt        repository.CreateRecurringInstanceOptions
		scope      models.Scope
		mockInsert mockInsertOne
		wantErr    bool
	}{
		"success": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID:   mockEventID.Hex(),
				Title:     "Test",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
		},
		"error_invalid_eventid": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID: "invalid-id",
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_department_ids": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID:       mockEventID.Hex(),
				Title:         "Test",
				StartTime:     mockTime,
				EndTime:       mockTime.Add(1 * time.Hour),
				DepartmentIDs: []string{"invalid-dept-id"},
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_category_id": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID:    mockEventID.Hex(),
				Title:      "Test",
				StartTime:  mockTime,
				EndTime:    mockTime.Add(1 * time.Hour),
				CategoryID: "invalid-category-id",
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_room_ids": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID:   mockEventID.Hex(),
				Title:     "Test",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
				RoomIDs:   []string{"invalid-room-id"},
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_timezone_id": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID:    mockEventID.Hex(),
				Title:      "Test",
				StartTime:  mockTime,
				EndTime:    mockTime.Add(1 * time.Hour),
				TimezoneID: "invalid-timezone-id",
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_branch_ids": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID:   mockEventID.Hex(),
				Title:     "Test",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
				BranchIDs: []string{"invalid-branch-id"},
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_insert": {
			opt: repository.CreateRecurringInstanceOptions{
				EventID:   mockEventID.Hex(),
				Title:     "Test",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
			},
			scope: scope,
			mockInsert: mockInsertOne{
				isCalled: true,
				err:      fmt.Errorf("insert error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			if tc.opt.EventID == mockEventID.Hex() {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockEventID)
				colName := fmt.Sprintf("recurring_instances_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
				deps.db.EXPECT().NewObjectID().Return(primitive.NewObjectID())
			}

			if tc.mockInsert.isCalled {
				deps.col.EXPECT().InsertOne(ctx, mock.AnythingOfType("models.RecurringInstance")).Return(nil, tc.mockInsert.err)
			}

			got, err := repo.CreateRecurringInstance(ctx, tc.scope, tc.opt)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.opt.Title, got.Title)
			}
		})
	}
}

func TestEventRepository_CreateManyRecurringInstances(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID := primitive.NewObjectID()
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	type mockInsertMany struct {
		isCalled bool
		err      error
	}

	tcs := map[string]struct {
		opt        repository.CreateManyRecurringInstancesOptions
		scope      models.Scope
		mockInsert mockInsertMany
		wantErr    bool
		wantNil    bool
	}{
		"success": {
			opt: repository.CreateManyRecurringInstancesOptions{
				EventID: mockEventID.Hex(),
				RecurringInstances: []repository.CreateRecurringInstanceOptions{
					{
						EventID:   mockEventID.Hex(),
						Title:     "Test",
						StartTime: mockTime,
						EndTime:   mockTime.Add(1 * time.Hour),
					},
				},
			},
			scope: scope,
			mockInsert: mockInsertMany{
				isCalled: true,
				err:      nil,
			},
			wantErr: false,
			wantNil: false,
		},
		"empty_instances": {
			opt: repository.CreateManyRecurringInstancesOptions{
				EventID:            mockEventID.Hex(),
				RecurringInstances: []repository.CreateRecurringInstanceOptions{},
			},
			scope:   scope,
			wantErr: false,
			wantNil: true,
		},
		"error_invalid_eventid": {
			opt: repository.CreateManyRecurringInstancesOptions{
				EventID: "invalid-id",
				RecurringInstances: []repository.CreateRecurringInstanceOptions{
					{
						EventID:   "invalid-id",
						Title:     "Test",
						StartTime: mockTime,
						EndTime:   mockTime.Add(1 * time.Hour),
					},
				},
			},
			scope: scope,
			mockInsert: mockInsertMany{
				isCalled: false,
			},
			wantErr: true,
			wantNil: true,
		},
		"error_build_models": {
			opt: repository.CreateManyRecurringInstancesOptions{
				EventID: mockEventID.Hex(),
				RecurringInstances: []repository.CreateRecurringInstanceOptions{
					{
						EventID:   mockEventID.Hex(),
						Title:     "Test",
						StartTime: mockTime,
						EndTime:   mockTime.Add(1 * time.Hour),
						BranchIDs: []string{"invalid-branch-id"},
					},
				},
			},
			scope: scope,
			mockInsert: mockInsertMany{
				isCalled: false,
			},
			wantErr: true,
			wantNil: true,
		},
		"error_insert_many": {
			opt: repository.CreateManyRecurringInstancesOptions{
				EventID: mockEventID.Hex(),
				RecurringInstances: []repository.CreateRecurringInstanceOptions{
					{
						EventID:   mockEventID.Hex(),
						Title:     "Test",
						StartTime: mockTime,
						EndTime:   mockTime.Add(1 * time.Hour),
					},
				},
			},
			scope: scope,
			mockInsert: mockInsertMany{
				isCalled: true,
				err:      fmt.Errorf("insert many error"),
			},
			wantErr: true,
			wantNil: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			// Nếu có instance và EventID hợp lệ, luôn mock Collection và NewObjectID
			if len(tc.opt.RecurringInstances) > 0 && tc.opt.EventID == mockEventID.Hex() {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockEventID)
				colName := fmt.Sprintf("recurring_instances_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
				for i := 0; i < len(tc.opt.RecurringInstances); i++ {
					deps.db.EXPECT().NewObjectID().Return(primitive.NewObjectID())
				}
			}

			if tc.mockInsert.isCalled {
				deps.col.EXPECT().InsertMany(ctx, mock.Anything).Return(nil, tc.mockInsert.err)
			}

			got, err := repo.CreateManyRecurringInstances(ctx, tc.scope, tc.opt)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.wantNil {
					require.Nil(t, got)
				} else {
					require.NotNil(t, got)
					require.Equal(t, tc.opt.RecurringInstances[0].Title, got[0].Title)
				}
			}
		})
	}
}

func TestEventRepository_DetailRecurringInstance(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID := primitive.NewObjectID()
	mockID := primitive.NewObjectID()
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	type mockFindOne struct {
		isCalled bool
		err      error
		result   models.RecurringInstance
	}

	tcs := map[string]struct {
		id        string
		eventID   string
		scope     models.Scope
		mockFind  mockFindOne
		wantErr   bool
		wantTitle string
	}{
		"success": {
			id:      mockID.Hex(),
			eventID: mockEventID.Hex(),
			scope:   scope,
			mockFind: mockFindOne{
				isCalled: true,
				err:      nil,
				result:   models.RecurringInstance{ID: mockID, Title: "Test"},
			},
			wantErr:   false,
			wantTitle: "Test",
		},
		"error_invalid_eventid": {
			id:      mockID.Hex(),
			eventID: "invalid-id",
			scope:   scope,
			mockFind: mockFindOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_build_query": {
			id:      "invalid-id",
			eventID: mockEventID.Hex(),
			scope:   scope,
			mockFind: mockFindOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_findone": {
			id:      mockID.Hex(),
			eventID: mockEventID.Hex(),
			scope:   scope,
			mockFind: mockFindOne{
				isCalled: true,
				err:      fmt.Errorf("findone error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			if tc.eventID == mockEventID.Hex() {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockEventID)
				colName := fmt.Sprintf("recurring_instances_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
			}

			if tc.mockFind.isCalled {
				deps.col.EXPECT().FindOne(ctx, mock.Anything).Return(&dummySingleResult{
					result: tc.mockFind.result,
					err:    tc.mockFind.err,
				})
			}

			got, err := repo.DetailRecurringInstance(ctx, tc.scope, tc.id, tc.eventID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantTitle, got.Title)
			}
		})
	}
}

func TestEventRepository_UpdateRecurringInstance(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID := primitive.NewObjectID()
	mockID := primitive.NewObjectID()
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	type mockUpdateOne struct {
		isCalled bool
		err      error
	}

	tcs := map[string]struct {
		opt        repository.UpdateRecurringInstanceOptions
		scope      models.Scope
		mockUpdate mockUpdateOne
		wantErr    bool
		wantTitle  string
	}{
		"success": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:        mockID.Hex(),
				EventID:   mockEventID.Hex(),
				Title:     "Updated",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: true,
				err:      nil,
			},
			wantErr:   false,
			wantTitle: "Updated",
		},
		"error_invalid_eventid": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:        mockID.Hex(),
				EventID:   "invalid-id",
				Title:     "Updated",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_build_query": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:        "invalid-id",
				EventID:   mockEventID.Hex(),
				Title:     "Updated",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_build_update": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:        mockID.Hex(),
				EventID:   mockEventID.Hex(),
				Title:     "Updated",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
				BranchIDs: []string{"invalid-branch-id"},
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_update_one": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:        mockID.Hex(),
				EventID:   mockEventID.Hex(),
				Title:     "Updated",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: true,
				err:      fmt.Errorf("update error"),
			},
			wantErr: true,
		},
		"error_invalid_department_ids": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:            mockID.Hex(),
				EventID:       mockEventID.Hex(),
				Title:         "Updated",
				StartTime:     mockTime,
				EndTime:       mockTime.Add(1 * time.Hour),
				DepartmentIDs: []string{"invalid-dept-id"},
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_timezone_id": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:         mockID.Hex(),
				EventID:    mockEventID.Hex(),
				Title:      "Updated",
				StartTime:  mockTime,
				EndTime:    mockTime.Add(1 * time.Hour),
				TimezoneID: "invalid-timezone-id",
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_room_ids": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:        mockID.Hex(),
				EventID:   mockEventID.Hex(),
				Title:     "Updated",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
				RoomIDs:   []string{"invalid-room-id"},
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_category_id": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:         mockID.Hex(),
				EventID:    mockEventID.Hex(),
				Title:      "Updated",
				StartTime:  mockTime,
				EndTime:    mockTime.Add(1 * time.Hour),
				CategoryID: "invalid-category-id",
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
		"error_invalid_branch_ids": {
			opt: repository.UpdateRecurringInstanceOptions{
				ID:        mockID.Hex(),
				EventID:   mockEventID.Hex(),
				Title:     "Updated",
				StartTime: mockTime,
				EndTime:   mockTime.Add(1 * time.Hour),
				BranchIDs: []string{"invalid-branch-id"},
			},
			scope: scope,
			mockUpdate: mockUpdateOne{
				isCalled: false,
			},
			wantErr: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			if tc.opt.EventID == mockEventID.Hex() {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockEventID)
				colName := fmt.Sprintf("recurring_instances_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
			}

			if tc.mockUpdate.isCalled {
				deps.col.EXPECT().UpdateOne(ctx, mock.Anything, mock.Anything).Return(nil, tc.mockUpdate.err)
			}

			got, err := repo.UpdateRecurringInstance(ctx, tc.scope, tc.opt)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantTitle, got.Title)
			}
		})
	}
}

func TestEventRepository_DeleteRecurringInstancesByEventID(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID := primitive.NewObjectID()
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	tcs := map[string]struct {
		eventID    string
		scope      models.Scope
		mockDelete error
		wantErr    bool
	}{
		"success": {
			eventID:    mockEventID.Hex(),
			scope:      scope,
			mockDelete: nil,
			wantErr:    false,
		},
		"error_invalid_eventid": {
			eventID: "invalid-id",
			scope:   scope,
			wantErr: true,
		},
		"error_build_scope": {
			eventID: mockEventID.Hex(),
			scope:   models.Scope{ShopID: "invalid-id"},
			wantErr: true,
		},
		"error_delete": {
			eventID:    mockEventID.Hex(),
			scope:      scope,
			mockDelete: fmt.Errorf("delete error"),
			wantErr:    true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repoIface, deps := initRepo(t, mockTime)
			repo := repoIface.(*implRepository) // ép kiểu về struct để gọi method

			if tc.eventID == mockEventID.Hex() {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockEventID)
				colName := fmt.Sprintf("recurring_instances_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
				if tc.mockDelete != nil || !tc.wantErr {
					deps.col.EXPECT().DeleteSoftMany(ctx, mock.Anything).Return(int64(1), tc.mockDelete)
				}
			}

			err := repo.DeleteRecurringInstancesByEventID(ctx, tc.scope, tc.eventID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventRepository_UpdateAttendanceRecurringInstance(t *testing.T) {
	mockShopID, _ := primitive.ObjectIDFromHex("667788990011223344556677")
	mockEventID := primitive.NewObjectID()
	mockID := primitive.NewObjectID()
	mockTime := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)

	scope := models.Scope{
		ShopID: mockShopID.Hex(),
		UserID: "user123",
	}

	tcs := map[string]struct {
		opt        repository.UpdateAttendanceRecurringInstanceOptions
		scope      models.Scope
		mockUpdate error
		wantErr    bool
	}{
		"success_accept": {
			opt: repository.UpdateAttendanceRecurringInstanceOptions{
				ID:      mockID.Hex(),
				EventID: mockEventID.Hex(),
				Status:  1,
			},
			scope:      scope,
			mockUpdate: nil,
			wantErr:    false,
		},
		"success_decline": {
			opt: repository.UpdateAttendanceRecurringInstanceOptions{
				ID:      mockID.Hex(),
				EventID: mockEventID.Hex(),
				Status:  -1,
			},
			scope:      scope,
			mockUpdate: nil,
			wantErr:    false,
		},
		"error_invalid_eventid": {
			opt: repository.UpdateAttendanceRecurringInstanceOptions{
				ID:      mockID.Hex(),
				EventID: "invalid-id",
				Status:  1,
			},
			scope:   scope,
			wantErr: true,
		},
		"error_build_scope": {
			opt: repository.UpdateAttendanceRecurringInstanceOptions{
				ID:      mockID.Hex(),
				EventID: mockEventID.Hex(),
				Status:  1,
			},
			scope:   models.Scope{ShopID: "invalid-id"},
			wantErr: true,
		},
		"error_invalid_id": {
			opt: repository.UpdateAttendanceRecurringInstanceOptions{
				ID:      "invalid-id",
				EventID: mockEventID.Hex(),
				Status:  1,
			},
			scope:   scope,
			wantErr: true,
		},
		"error_update": {
			opt: repository.UpdateAttendanceRecurringInstanceOptions{
				ID:      mockID.Hex(),
				EventID: mockEventID.Hex(),
				Status:  1,
			},
			scope:      scope,
			mockUpdate: fmt.Errorf("update error"),
			wantErr:    true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, deps := initRepo(t, mockTime)

			shouldMockCollection := tc.opt.EventID == mockEventID.Hex()
			if shouldMockCollection {
				p, y := pkgmongo.GetPeriodAndYearFromObjectID(mockEventID)
				colName := fmt.Sprintf("recurring_instances_%d_%d", y, p)
				deps.db.EXPECT().Collection(colName).Return(deps.col)
			}

			if shouldMockCollection && tc.opt.ID != "invalid-id" && tc.scope.ShopID != "invalid-id" {
				deps.col.EXPECT().UpdateMany(ctx, mock.Anything, mock.Anything).Return(nil, tc.mockUpdate)
			}

			err := repo.UpdateAttendanceRecurringInstance(ctx, tc.scope, tc.opt)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
