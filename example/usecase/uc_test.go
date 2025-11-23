package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/gma-vietnam/tanca-connect/internal/device"
	"gitlab.com/gma-vietnam/tanca-connect/internal/element"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq/producer"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/eventcategory"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/internal/resources/notification"
	"gitlab.com/gma-vietnam/tanca-connect/internal/room"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mockDeps struct {
	repo            *repository.MockRepository
	elementUC       *element.MockUseCase
	roomUC          *room.MockUseCase
	shopUC          *microservice.MockShopUseCase
	eventCategoryUC *eventcategory.MockUseCase
	prod            *producer.MockProducer
}

func initUseCase(t *testing.T) (event.UseCase, mockDeps) {
	t.Helper()

	l := log.InitializeTestZapLogger()

	repo := repository.NewMockRepository(t)
	deviceUC := device.NewMockUseCase(t)
	elementUC := element.NewMockUseCase(t)
	roomUC := room.NewMockUseCase(t)
	shopUC := microservice.NewMockShopUseCase(t)
	eventCategoryUC := eventcategory.NewMockUseCase(t)
	prod := producer.NewMockProducer(t)

	return New(l, repo, deviceUC, elementUC, roomUC, shopUC, eventCategoryUC, prod), mockDeps{
		repo:            repo,
		elementUC:       elementUC,
		roomUC:          roomUC,
		shopUC:          shopUC,
		eventCategoryUC: eventCategoryUC,
		prod:            prod,
	}
}

// func TestCreate(t *testing.T) {
// 	scope := models.Scope{
// 		Suffix: "test",
// 		ShopID: primitive.NewObjectID().Hex(),
// 		UserID: "test",
// 	}
// 	ctx := context.Background()

// 	type mockRepoCreate struct {
// 		isCalled bool
// 		input    event.CreateOptions
// 		output   models.Event
// 		err      error
// 	}

// 	type mockRepo struct {
// 		createEvent mockRepoCreate
// 	}

// 	type mock struct {
// 		repo mockRepo
// 	}

// 	expectedErr := errors.New("error")

// 	tcs := map[string]struct {
// 		input  event.CreateInput
// 		mock   mock
// 		output models.Event
// 		err    error
// 	}{
// 		"success": {
// 			input: event.CreateInput{
// 				Title: "test event",
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					createEvent: mockRepoCreate{
// 						isCalled: true,
// 						input: event.CreateOptions{
// 							Title: "test event",
// 						},
// 						output: models.Event{},
// 					},
// 				},
// 			},
// 			output: models.Event{},
// 		},
// 		"err repo Create": {
// 			input: event.CreateInput{
// 				Title: "test event",
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					createEvent: mockRepoCreate{
// 						isCalled: true,
// 						input: event.CreateOptions{
// 							Title: "test event",
// 						},
// 						err: expectedErr,
// 					},
// 				},
// 			},
// 			err: expectedErr,
// 		},
// 		"err required field": {
// 			input: event.CreateInput{
// 				Title: "",
// 			},
// 			err: event.ErrRequiredField,
// 		},
// 	}

// 	for n, tc := range tcs {
// 		t.Run(n, func(t *testing.T) {
// 			uc, deps := initUseCase(t)

// 			if tc.mock.repo.createEvent.isCalled {
// 				deps.repo.EXPECT().Create(ctx, scope, tc.mock.repo.createEvent.input).
// 					Return(
// 						tc.mock.repo.createEvent.output,
// 						tc.mock.repo.createEvent.err,
// 					)
// 			}

// 			res, err := uc.Create(ctx, scope, tc.input)
// 			if err != nil {
// 				require.EqualError(t, err, tc.err.Error())
// 			} else {
// 				require.Equal(t, tc.output, res)
// 			}
// 		})
// 	}
// }

// func TestDetail(t *testing.T) {
// 	scope := models.Scope{
// 		Suffix: "test",
// 		ShopID: primitive.NewObjectID().Hex(),
// 		UserID: "test",
// 	}
// 	ctx := context.Background()

// 	type mockRepoDetail struct {
// 		isCalled bool
// 		input    string
// 		output   models.Event
// 		err      error
// 	}

// 	type mockRepo struct {
// 		detailEvent mockRepoDetail
// 	}

// 	type mock struct {
// 		repo mockRepo
// 	}

// 	expectedID := primitive.NewObjectID()
// 	expectedErr := errors.New("error")

// 	tcs := map[string]struct {
// 		input  string
// 		mock   mock
// 		output models.Event
// 		err    error
// 	}{
// 		"success": {
// 			input: expectedID.Hex(),
// 			mock: mock{
// 				repo: mockRepo{
// 					detailEvent: mockRepoDetail{
// 						isCalled: true,
// 						input:    expectedID.Hex(),
// 						output:   models.Event{},
// 					},
// 				},
// 			},
// 			output: models.Event{},
// 		},
// 		"err repo Detail": {
// 			input: expectedID.Hex(),
// 			mock: mock{
// 				repo: mockRepo{
// 					detailEvent: mockRepoDetail{
// 						isCalled: true,
// 						input:    expectedID.Hex(),
// 						err:      expectedErr,
// 					},
// 				},
// 			},
// 			err: expectedErr,
// 		},
// 		"err event.ErrEventNotFound": {
// 			input: expectedID.Hex(),
// 			mock: mock{
// 				repo: mockRepo{
// 					detailEvent: mockRepoDetail{
// 						isCalled: true,
// 						input:    expectedID.Hex(),
// 						err:      mongo.ErrNoDocuments,
// 					},
// 				},
// 			},
// 			err: event.ErrEventNotFound,
// 		},
// 	}

// 	for n, tc := range tcs {
// 		t.Run(n, func(t *testing.T) {
// 			uc, deps := initUseCase(t)

// 			if tc.mock.repo.detailEvent.isCalled {
// 				deps.repo.EXPECT().Detail(ctx, scope, tc.mock.repo.detailEvent.input).
// 					Return(
// 						tc.mock.repo.detailEvent.output,
// 						tc.mock.repo.detailEvent.err,
// 					)
// 			}

// 			res, err := uc.Detail(ctx, scope, tc.input, "")
// 			if err != nil {
// 				require.EqualError(t, err, tc.err.Error())
// 			} else {
// 				require.Equal(t, tc.output, res)
// 			}
// 		})
// 	}
// }

// func TestUpdate(t *testing.T) {
// 	scope := models.Scope{
// 		Suffix: "test",
// 		ShopID: primitive.NewObjectID().Hex(),
// 		UserID: "test",
// 	}
// 	ctx := context.Background()

// 	type mockRepoDetail struct {
// 		isCalled bool
// 		input    string
// 		output   models.Event
// 		err      error
// 	}

// 	type mockRepoUpdate struct {
// 		isCalled bool
// 		input    event.UpdateOptions
// 		output   models.Event
// 		err      error
// 	}

// 	type mockRepo struct {
// 		detailEvent mockRepoDetail
// 		updateEvent mockRepoUpdate
// 	}

// 	type mock struct {
// 		repo mockRepo
// 	}

// 	expectedID := primitive.NewObjectID()
// 	expectedErr := errors.New("error")

// 	tcs := map[string]struct {
// 		input  event.UpdateInput
// 		mock   mock
// 		output models.Event
// 		err    error
// 	}{
// 		"success": {
// 			input: event.UpdateInput{
// 				ID:    expectedID.Hex(),
// 				Title: "new name",
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					detailEvent: mockRepoDetail{
// 						isCalled: true,
// 						input:    expectedID.Hex(),
// 						output:   models.Event{},
// 					},
// 					updateEvent: mockRepoUpdate{
// 						isCalled: true,
// 						input: event.UpdateOptions{
// 							Model: models.Event{},
// 							Title: "new name",
// 						},
// 						output: models.Event{},
// 					},
// 				},
// 			},
// 			output: models.Event{},
// 		},
// 		"err repo Update": {
// 			input: event.UpdateInput{
// 				ID:    expectedID.Hex(),
// 				Title: "new name",
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					detailEvent: mockRepoDetail{
// 						isCalled: true,
// 						input:    expectedID.Hex(),
// 						output:   models.Event{},
// 					},
// 					updateEvent: mockRepoUpdate{
// 						isCalled: true,
// 						input: event.UpdateOptions{
// 							Model: models.Event{},
// 							Title: "new name",
// 						},
// 						err: expectedErr,
// 					},
// 				},
// 			},
// 			err: expectedErr,
// 		},
// 		"err repo Detail": {
// 			input: event.UpdateInput{
// 				ID:    expectedID.Hex(),
// 				Title: "new name",
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					detailEvent: mockRepoDetail{
// 						isCalled: true,
// 						input:    expectedID.Hex(),
// 						err:      expectedErr,
// 					},
// 				},
// 			},
// 			err: expectedErr,
// 		},
// 		"err event.ErrEventNotFound": {
// 			input: event.UpdateInput{
// 				ID:    expectedID.Hex(),
// 				Title: "new name",
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					detailEvent: mockRepoDetail{
// 						isCalled: true,
// 						input:    expectedID.Hex(),
// 						err:      mongo.ErrNoDocuments,
// 					},
// 				},
// 			},
// 			err: event.ErrEventNotFound,
// 		},
// 		"err required field": {
// 			input: event.UpdateInput{
// 				ID:    expectedID.Hex(),
// 				Title: "",
// 			},
// 			err: event.ErrRequiredField,
// 		},
// 	}

// 	for n, tc := range tcs {
// 		t.Run(n, func(t *testing.T) {
// 			uc, deps := initUseCase(t)

// 			if tc.mock.repo.updateEvent.isCalled {
// 				deps.repo.EXPECT().Update(ctx, scope, tc.mock.repo.updateEvent.input).
// 					Return(
// 						tc.mock.repo.updateEvent.output,
// 						tc.mock.repo.updateEvent.err,
// 					)
// 			}

// 			if tc.mock.repo.detailEvent.isCalled {
// 				deps.repo.EXPECT().Detail(ctx, scope, tc.mock.repo.detailEvent.input).
// 					Return(
// 						tc.mock.repo.detailEvent.output,
// 						tc.mock.repo.detailEvent.err,
// 					)
// 			}

// 			err := uc.Update(ctx, scope, tc.input)
// 			if err != nil {
// 				require.EqualError(t, err, tc.err.Error())
// 			}
// 		})
// 	}
// }

func TestDelete(t *testing.T) {
	scope := models.Scope{
		Suffix: "test",
		ShopID: "68584e3a898df90f470fdd30",
		UserID: "test-user-id",
	}
	ctx := context.Background()

	type mockRepoDeleteSingle struct {
		isCalled bool
		input    string
		err      error
	}

	type mockRepoDetailEvent struct {
		isCalled bool
		input    string
		output   models.Event
		err      error
	}

	type mockRepoDetailRecurringInstance struct {
		isCalled     bool
		inputID      string
		inputEventID string
		output       models.RecurringInstance
		err          error
	}

	type mockRepoDeleteRecurringInstance struct {
		isCalled bool
		input    repository.DeleteRecurringInstanceOptions
		err      error
	}

	type mockRepoDeleteNextRecurringInstances struct {
		isCalled bool
		input    repository.DeleteNextRecurringInstancesOptions
		err      error
	}

	type mockRepoUpdateRepeatUntilRecurringTrackings struct {
		isCalled bool
		input    repository.UpdateRepeatUntilRecurringTrackingsOptions
		err      error
	}

	type mockRepoDeleteRecurringTracking struct {
		isCalled bool
		input    repository.DeleteRecurringTrackingOptions
		err      error
	}

	type mockDetailShop struct {
		isCalled bool
		input    string
		output   microservice.Shop
		err      error
	}

	type mockPublishPushNotiMsg struct {
		isCalled bool
		input    rabbitmq.PushNotiMsg
		err      error
	}

	type mockGetNotiHeading struct {
		isCalled bool
		wantVi   string
		wantEn   string
		wantJa   string
		errVi    error
		errEn    error
		errJa    error
	}

	type mockGetNotiContent struct {
		isCalled bool
		wantVi   string
		wantEn   string
		wantJa   string
		errVi    error
		errEn    error
		errJa    error
	}

	type mockListAllUsers struct {
		isCalled bool
		input    microservice.GetUsersFilter
		output   []microservice.User
		err      error
	}

	type mockUpdateTrackingException struct {
		isCalled bool
		input    room.UpdateTrackingExceptionInput
		err      error
	}

	type mockUpdateTrackingRepeatUntil struct {
		isCalled bool
		input    room.UpdateTrackingRepeatUntilInput
		err      error
	}

	type mockDeleteTracking struct {
		isCalled bool
		input    room.DeleteTrackingInput
		err      error
	}

	type mockUpdateRepeatUntil struct {
		isCalled  bool
		inputID   string
		inputTime time.Time
		output    models.Event
		err       error
	}

	type mockRepo struct {
		deleteEvent                         mockRepoDeleteSingle
		detailEvent                         mockRepoDetailEvent
		detailRecurringInstance             mockRepoDetailRecurringInstance
		deleteRecurringInstance             mockRepoDeleteRecurringInstance
		deleteNextRecurringInstances        mockRepoDeleteNextRecurringInstances
		updateRepeatUntilRecurringTrackings mockRepoUpdateRepeatUntilRecurringTrackings
		deleteRecurringTracking             mockRepoDeleteRecurringTracking
		updateRepeatUntil                   mockUpdateRepeatUntil
	}

	type mockShop struct {
		detailShop   mockDetailShop
		listAllUsers mockListAllUsers
	}

	type mockRoom struct {
		updateTrackingException   mockUpdateTrackingException
		updateTrackingRepeatUntil mockUpdateTrackingRepeatUntil
		deleteTracking            mockDeleteTracking
	}

	type mock struct {
		repo mockRepo
		room mockRoom
	}

	type mockProducer struct {
		publishPushNotiMsg mockPublishPushNotiMsg
	}

	expectedID := primitive.NewObjectID()
	expectedEventID := primitive.NewObjectID()
	expectedErr := errors.New("repository error")
	recurringInstanceTime := time.Now().UTC()
	expectedNoneRecurringEvent := models.Event{
		ID:          expectedID,
		CreatedByID: scope.UserID,
		StartTime:   recurringInstanceTime,
		EndTime:     recurringInstanceTime.Add(time.Hour),
		RepeatUntil: &recurringInstanceTime,
	}
	prevTime, _ := time.Parse(time.RFC3339, "2023-08-01T10:00:00Z")

	tcs := map[string]struct {
		input          event.DeleteInput
		mock           mock
		shop           mockShop
		prod           mockProducer
		getNotiHeading mockGetNotiHeading
		getNotiContent mockGetNotiContent
		err            error
	}{
		"success delete single event": {
			input: event.DeleteInput{
				ID: expectedID.Hex(),
			},
			mock: mock{
				repo: mockRepo{
					detailEvent: mockRepoDetailEvent{
						isCalled: true,
						input:    expectedID.Hex(),
						output:   expectedNoneRecurringEvent,
					},
					deleteEvent: mockRepoDeleteSingle{
						isCalled: true,
						input:    expectedID.Hex(),
					},
				},
			},
			getNotiHeading: mockGetNotiHeading{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			getNotiContent: mockGetNotiContent{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			shop: mockShop{
				detailShop: mockDetailShop{
					isCalled: true,
					input:    scope.ShopID,
					output: microservice.Shop{
						ID:         scope.ShopID,
						DateFormat: "DD/MM/YYYY",
						TimeFormat: "24hour",
					},
				},
				listAllUsers: mockListAllUsers{
					isCalled: true,
					input:    microservice.GetUsersFilter{},
					output:   []microservice.User{},
				},
			},
			prod: mockProducer{
				publishPushNotiMsg: mockPublishPushNotiMsg{
					isCalled: true,
					input: rabbitmq.PushNotiMsg{
						ShopScope: rabbitmq.ShopScope{
							ID:     "68584e3a898df90f470fdd30",
							Suffix: "test",
						},
						Content:       "test",
						Heading:       "test",
						UserIDs:       []string{"test-user-id"},
						CreatedUserID: "test-user-id",
						En: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Ja: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Data: rabbitmq.NotiData{
							Data: event.PublishNotiEventInput{
								ID:          expectedID.Hex(),
								EventID:     expectedID.Hex(),
								CreatedByID: scope.UserID,
								StartTime:   util.DateTimeToStr(recurringInstanceTime, nil),
							},
							Activity: "EVENT_DETAIL",
						},
						Source: "event",
					},
				},
			},
		},
		"error getting event detail": {
			input: event.DeleteInput{
				ID: expectedID.Hex(),
			},
			mock: mock{
				repo: mockRepo{
					detailEvent: mockRepoDetailEvent{
						isCalled: true,
						input:    expectedID.Hex(),
						err:      expectedErr,
					},
				},
			},
			err: expectedErr,
		},
		// "error event not found": {
		// 	input: event.DeleteInput{
		// 		ID: expectedID.Hex(),
		// 	},
		// 	mock: mock{
		// 		repo: mockRepo{
		// 			detailEvent: mockRepoDetailEvent{
		// 				isCalled: true,
		// 				input:    expectedID.Hex(),
		// 				err:      mongo.ErrNoDocuments,
		// 			},
		// 		},
		// 	},
		// 	err: mongo.ErrNoDocuments,
		// },
		"error unauthorized to delete event": {
			input: event.DeleteInput{
				ID: expectedID.Hex(),
			},
			mock: mock{
				repo: mockRepo{
					detailEvent: mockRepoDetailEvent{
						isCalled: true,
						input:    expectedID.Hex(),
						output: models.Event{
							ID:          expectedID,
							CreatedByID: "different-user-id",
						},
					},
				},
			},
			err: event.ErrEventEditNotAllowed,
		},
		"error in delete repository call": {
			input: event.DeleteInput{
				ID: expectedID.Hex(),
			},
			mock: mock{
				repo: mockRepo{
					detailEvent: mockRepoDetailEvent{
						isCalled: true,
						input:    expectedID.Hex(),
						output: models.Event{
							ID:          expectedID,
							CreatedByID: scope.UserID,
						},
					},
					deleteEvent: mockRepoDeleteSingle{
						isCalled: true,
						input:    expectedID.Hex(),
						err:      expectedErr,
					},
				},
			},
			shop: mockShop{
				detailShop: mockDetailShop{
					isCalled: true,
					input:    scope.ShopID,
					output: microservice.Shop{
						ID:         scope.ShopID,
						DateFormat: "DD/MM/YYYY",
						TimeFormat: "24hour",
					},
				},
				listAllUsers: mockListAllUsers{
					isCalled: true,
					input:    microservice.GetUsersFilter{},
					output:   []microservice.User{},
				},
			},
			getNotiHeading: mockGetNotiHeading{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			getNotiContent: mockGetNotiContent{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			prod: mockProducer{
				publishPushNotiMsg: mockPublishPushNotiMsg{
					isCalled: true,
					input: rabbitmq.PushNotiMsg{
						ShopScope: rabbitmq.ShopScope{
							ID:     scope.ShopID,
							Suffix: scope.Suffix,
						},
						Content:       "test",
						Heading:       "test",
						UserIDs:       []string{scope.UserID},
						CreatedUserID: scope.UserID,
						En: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Ja: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Data: rabbitmq.NotiData{
							Data: event.PublishNotiEventInput{
								ID:          expectedID.Hex(),
								EventID:     expectedID.Hex(),
								CreatedByID: scope.UserID,
								StartTime:   util.DateTimeToStr(time.Time{}, nil),
							},
							Activity: "EVENT_DETAIL",
						},
						Source: "event",
					},
				},
			},
			err: expectedErr,
		},
		"success delete one recurring instance": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionOne,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
						},
					},
					deleteRecurringInstance: mockRepoDeleteRecurringInstance{
						isCalled: true,
						input: repository.DeleteRecurringInstanceOptions{
							IDs:     []string{expectedID.Hex()},
							EventID: expectedEventID.Hex(),
						},
					},
				},
			},
			shop: mockShop{
				detailShop: mockDetailShop{
					isCalled: true,
					input:    scope.ShopID,
					output: microservice.Shop{
						ID:         scope.ShopID,
						DateFormat: "DD/MM/YYYY",
						TimeFormat: "24hour",
					},
				},
				listAllUsers: mockListAllUsers{
					isCalled: true,
					input:    microservice.GetUsersFilter{},
					output:   []microservice.User{},
				},
			},
			getNotiHeading: mockGetNotiHeading{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			getNotiContent: mockGetNotiContent{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			prod: mockProducer{
				publishPushNotiMsg: mockPublishPushNotiMsg{
					isCalled: true,
					input: rabbitmq.PushNotiMsg{
						ShopScope: rabbitmq.ShopScope{
							ID:     scope.ShopID,
							Suffix: scope.Suffix,
						},
						Content:       "test",
						Heading:       "test",
						UserIDs:       []string{scope.UserID},
						CreatedUserID: scope.UserID,
						En: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Ja: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Data: rabbitmq.NotiData{
							Data: event.PublishNotiEventInput{
								ID:          expectedID.Hex(),
								EventID:     expectedEventID.Hex(),
								CreatedByID: scope.UserID,
								StartTime:   util.DateTimeToStr(time.Time{}, nil),
							},
							Activity: "EVENT_DETAIL",
						},
						Source: "event",
					},
				},
			},
		},
		"error in delete recurring instance": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionOne,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
						},
					},
					deleteRecurringInstance: mockRepoDeleteRecurringInstance{
						isCalled: true,
						input: repository.DeleteRecurringInstanceOptions{
							IDs:     []string{expectedID.Hex()},
							EventID: expectedEventID.Hex(),
						},
						err: expectedErr,
					},
				},
			},
			err: expectedErr,
		},
		"success delete from date": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionFrom,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
							StartTime:   recurringInstanceTime,
							EndTime:     recurringInstanceTime.Add(time.Hour),
							Repeat:      models.EventRepeatDaily,
						},
					},
					deleteNextRecurringInstances: mockRepoDeleteNextRecurringInstances{
						isCalled: true,
						input: repository.DeleteNextRecurringInstancesOptions{
							EventID: expectedEventID.Hex(),
							Date:    prevTime,
						},
					},
					updateRepeatUntilRecurringTrackings: mockRepoUpdateRepeatUntilRecurringTrackings{
						isCalled: true,
						input: repository.UpdateRepeatUntilRecurringTrackingsOptions{
							EventID:     expectedEventID.Hex(),
							RepeatUntil: &prevTime,
						},
					},
					updateRepeatUntil: mockUpdateRepeatUntil{
						isCalled:  true,
						inputID:   expectedEventID.Hex(),
						inputTime: prevTime,
						output:    models.Event{},
					},
				},
				room: mockRoom{
					updateTrackingRepeatUntil: mockUpdateTrackingRepeatUntil{
						isCalled: true,
						input: room.UpdateTrackingRepeatUntilInput{
							RoomIDs:     []string{},
							EventID:     expectedEventID.Hex(),
							RepeatUntil: prevTime,
						},
					},
				},
			},
			shop: mockShop{
				detailShop: mockDetailShop{
					isCalled: true,
					input:    scope.ShopID,
					output: microservice.Shop{
						ID:         scope.ShopID,
						DateFormat: "DD/MM/YYYY",
						TimeFormat: "24hour",
					},
				},
				listAllUsers: mockListAllUsers{
					isCalled: true,
					input:    microservice.GetUsersFilter{},
					output:   []microservice.User{},
				},
			},
			getNotiHeading: mockGetNotiHeading{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			getNotiContent: mockGetNotiContent{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			prod: mockProducer{
				publishPushNotiMsg: mockPublishPushNotiMsg{
					isCalled: true,
					input: rabbitmq.PushNotiMsg{
						ShopScope: rabbitmq.ShopScope{
							ID:     scope.ShopID,
							Suffix: scope.Suffix,
						},
						Content:       "test",
						Heading:       "test",
						UserIDs:       []string{scope.UserID},
						CreatedUserID: scope.UserID,
						En: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Ja: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Data: rabbitmq.NotiData{
							Data: event.PublishNotiEventInput{
								ID:          expectedID.Hex(),
								EventID:     expectedEventID.Hex(),
								CreatedByID: scope.UserID,
								StartTime:   util.DateTimeToStr(recurringInstanceTime, nil),
							},
							Activity: "EVENT_DETAIL",
						},
						Source: "event",
					},
				},
			},
		},
		"success delete all instances": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionAll,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
						},
					},
					deleteRecurringInstance: mockRepoDeleteRecurringInstance{
						isCalled: true,
						input: repository.DeleteRecurringInstanceOptions{
							EventID: expectedEventID.Hex(),
						},
					},
					deleteEvent: mockRepoDeleteSingle{
						isCalled: true,
						input:    expectedEventID.Hex(),
					},
					deleteRecurringTracking: mockRepoDeleteRecurringTracking{
						isCalled: true,
						input: repository.DeleteRecurringTrackingOptions{
							EventID: expectedEventID.Hex(),
						},
					},
				},
				room: mockRoom{
					deleteTracking: mockDeleteTracking{
						isCalled: true,
						input: room.DeleteTrackingInput{
							RoomIDs: []string{},
							EventID: expectedEventID.Hex(),
						},
					},
				},
			},
			shop: mockShop{
				detailShop: mockDetailShop{
					isCalled: true,
					input:    scope.ShopID,
					output: microservice.Shop{
						ID:         scope.ShopID,
						DateFormat: "DD/MM/YYYY",
						TimeFormat: "24hour",
					},
				},
				listAllUsers: mockListAllUsers{
					isCalled: true,
					input:    microservice.GetUsersFilter{},
					output:   []microservice.User{},
				},
			},
			getNotiHeading: mockGetNotiHeading{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			getNotiContent: mockGetNotiContent{
				isCalled: true,
				wantVi:   "test",
				wantEn:   "test",
				wantJa:   "test",
			},
			prod: mockProducer{
				publishPushNotiMsg: mockPublishPushNotiMsg{
					isCalled: true,
					input: rabbitmq.PushNotiMsg{
						ShopScope: rabbitmq.ShopScope{
							ID:     scope.ShopID,
							Suffix: scope.Suffix,
						},
						Content:       "test",
						Heading:       "test",
						UserIDs:       []string{scope.UserID},
						CreatedUserID: scope.UserID,
						En: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Ja: rabbitmq.MultiLangObj{
							Heading: "test",
							Content: "test",
						},
						Data: rabbitmq.NotiData{
							Data: event.PublishNotiEventInput{
								ID:          expectedID.Hex(),
								EventID:     expectedEventID.Hex(),
								CreatedByID: scope.UserID,
								StartTime:   util.DateTimeToStr(time.Time{}, nil),
							},
							Activity: "EVENT_DETAIL",
						},
						Source: "event",
					},
				},
			},
		},
		"error getting recurring instance detail": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionOne,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						err:          expectedErr,
					},
				},
			},
			err: expectedErr,
		},
		"error unauthorized to delete recurring instance": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionOne,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: "different-user-id",
						},
					},
				},
			},
			err: event.ErrEventEditNotAllowed,
		},
		"error delete from date - fail on DeleteNextRecurringInstances": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionFrom,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
							StartTime:   recurringInstanceTime,
							EndTime:     recurringInstanceTime.Add(time.Hour),
							Repeat:      models.EventRepeatDaily,
						},
					},
					deleteNextRecurringInstances: mockRepoDeleteNextRecurringInstances{
						isCalled: true,
						input: repository.DeleteNextRecurringInstancesOptions{
							EventID: expectedEventID.Hex(),
							Date:    prevTime,
						},
						err: expectedErr,
					},
					updateRepeatUntilRecurringTrackings: mockRepoUpdateRepeatUntilRecurringTrackings{
						isCalled: true,
						input: repository.UpdateRepeatUntilRecurringTrackingsOptions{
							EventID:     expectedEventID.Hex(),
							RepeatUntil: &prevTime,
						},
					},
					updateRepeatUntil: mockUpdateRepeatUntil{
						isCalled:  true,
						inputID:   expectedEventID.Hex(),
						inputTime: prevTime,
						output:    models.Event{},
					},
				},
				room: mockRoom{
					updateTrackingRepeatUntil: mockUpdateTrackingRepeatUntil{
						isCalled: true,
						input: room.UpdateTrackingRepeatUntilInput{
							RoomIDs:     []string{},
							EventID:     expectedEventID.Hex(),
							RepeatUntil: prevTime,
						},
					},
				},
			},
			err: expectedErr,
		},
		"error delete from date - fail on UpdateRepeatUntilRecurringTrackings": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionFrom,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
							StartTime:   recurringInstanceTime,
							EndTime:     recurringInstanceTime.Add(time.Hour),
							Repeat:      models.EventRepeatDaily,
						},
					},
					deleteNextRecurringInstances: mockRepoDeleteNextRecurringInstances{
						isCalled: true,
						input: repository.DeleteNextRecurringInstancesOptions{
							EventID: expectedEventID.Hex(),
							Date:    prevTime,
						},
					},
					updateRepeatUntilRecurringTrackings: mockRepoUpdateRepeatUntilRecurringTrackings{
						isCalled: true,
						input: repository.UpdateRepeatUntilRecurringTrackingsOptions{
							EventID:     expectedEventID.Hex(),
							RepeatUntil: &prevTime,
						},
						err: expectedErr,
					},
					updateRepeatUntil: mockUpdateRepeatUntil{
						isCalled:  true,
						inputID:   expectedEventID.Hex(),
						inputTime: prevTime,
						output:    models.Event{},
					},
				},
				room: mockRoom{
					updateTrackingRepeatUntil: mockUpdateTrackingRepeatUntil{
						isCalled: true,
						input: room.UpdateTrackingRepeatUntilInput{
							RoomIDs:     []string{},
							EventID:     expectedEventID.Hex(),
							RepeatUntil: prevTime,
						},
					},
				},
			},
			err: expectedErr,
		},
		"error delete all instances - fail on DeleteRecurringInstance": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionAll,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
						},
					},
					deleteRecurringInstance: mockRepoDeleteRecurringInstance{
						isCalled: true,
						input: repository.DeleteRecurringInstanceOptions{
							EventID: expectedEventID.Hex(),
						},
						err: expectedErr,
					},
					deleteEvent: mockRepoDeleteSingle{
						isCalled: true,
						input:    expectedEventID.Hex(),
					},
					deleteRecurringTracking: mockRepoDeleteRecurringTracking{
						isCalled: true,
						input: repository.DeleteRecurringTrackingOptions{
							EventID: expectedEventID.Hex(),
						},
					},
				},
				room: mockRoom{
					deleteTracking: mockDeleteTracking{
						isCalled: true,
						input: room.DeleteTrackingInput{
							RoomIDs: []string{},
							EventID: expectedEventID.Hex(),
						},
					},
				},
			},
			err: expectedErr,
		},
		"error delete all instances - fail on Delete": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionAll,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
						},
					},
					deleteRecurringInstance: mockRepoDeleteRecurringInstance{
						isCalled: true,
						input: repository.DeleteRecurringInstanceOptions{
							EventID: expectedEventID.Hex(),
						},
					},
					deleteEvent: mockRepoDeleteSingle{
						isCalled: true,
						input:    expectedEventID.Hex(),
						err:      expectedErr,
					},
					deleteRecurringTracking: mockRepoDeleteRecurringTracking{
						isCalled: true,
						input: repository.DeleteRecurringTrackingOptions{
							EventID: expectedEventID.Hex(),
						},
					},
				},
				room: mockRoom{
					deleteTracking: mockDeleteTracking{
						isCalled: true,
						input: room.DeleteTrackingInput{
							RoomIDs: []string{},
							EventID: expectedEventID.Hex(),
						},
					},
				},
			},
			err: expectedErr,
		},
		"error delete all instances - fail on DeleteRecurringTracking": {
			input: event.DeleteInput{
				ID:      expectedID.Hex(),
				EventID: expectedEventID.Hex(),
				Type:    models.EventActionAll,
			},
			mock: mock{
				repo: mockRepo{
					detailRecurringInstance: mockRepoDetailRecurringInstance{
						isCalled:     true,
						inputID:      expectedID.Hex(),
						inputEventID: expectedEventID.Hex(),
						output: models.RecurringInstance{
							ID:          expectedID,
							EventID:     expectedEventID,
							CreatedByID: scope.UserID,
						},
					},
					deleteRecurringInstance: mockRepoDeleteRecurringInstance{
						isCalled: true,
						input: repository.DeleteRecurringInstanceOptions{
							EventID: expectedEventID.Hex(),
						},
					},
					deleteEvent: mockRepoDeleteSingle{
						isCalled: true,
						input:    expectedEventID.Hex(),
					},
					deleteRecurringTracking: mockRepoDeleteRecurringTracking{
						isCalled: true,
						input: repository.DeleteRecurringTrackingOptions{
							EventID: expectedEventID.Hex(),
						},
						err: expectedErr,
					},
				},
				room: mockRoom{
					deleteTracking: mockDeleteTracking{
						isCalled: true,
						input: room.DeleteTrackingInput{
							RoomIDs: []string{},
							EventID: expectedEventID.Hex(),
						},
					},
				},
			},
			err: expectedErr,
		},
		// "error recurring instance not found": {
		// 	input: event.DeleteInput{
		// 		ID:      expectedID.Hex(),
		// 		EventID: expectedEventID.Hex(),
		// 		Type:    models.EventActionOne,
		// 	},
		// 	mock: mock{
		// 		repo: mockRepo{
		// 			detailRecurringInstance: mockRepoDetailRecurringInstance{
		// 				isCalled:     true,
		// 				inputID:      expectedID.Hex(),
		// 				inputEventID: expectedEventID.Hex(),
		// 				err:          mongo.ErrNoDocuments,
		// 			},
		// 		},
		// 	},
		// 	err: mongo.ErrNoDocuments,
		// },
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.repo.detailEvent.isCalled {
				deps.repo.EXPECT().Detail(ctx, scope, tc.mock.repo.detailEvent.input).
					Return(tc.mock.repo.detailEvent.output, tc.mock.repo.detailEvent.err)
			}

			if tc.mock.repo.deleteEvent.isCalled {
				deps.repo.EXPECT().Delete(ctx, scope, tc.mock.repo.deleteEvent.input).
					Return(tc.mock.repo.deleteEvent.err)
			}

			if tc.mock.repo.detailRecurringInstance.isCalled {
				deps.repo.EXPECT().DetailRecurringInstance(ctx, scope,
					tc.mock.repo.detailRecurringInstance.inputID,
					tc.mock.repo.detailRecurringInstance.inputEventID).
					Return(tc.mock.repo.detailRecurringInstance.output,
						tc.mock.repo.detailRecurringInstance.err)
			}

			if tc.mock.repo.deleteRecurringInstance.isCalled {
				deps.repo.EXPECT().DeleteRecurringInstance(ctx, scope, testifyMock.MatchedBy(func(options repository.DeleteRecurringInstanceOptions) bool {
					if len(options.IDs) > 0 {
						return options.IDs[0] == tc.mock.repo.deleteRecurringInstance.input.IDs[0] &&
							options.EventID == tc.mock.repo.deleteRecurringInstance.input.EventID
					}
					return options.EventID == tc.mock.repo.deleteRecurringInstance.input.EventID
				})).Return(tc.mock.repo.deleteRecurringInstance.err)
			}

			if tc.mock.repo.deleteNextRecurringInstances.isCalled {
				deps.repo.EXPECT().DeleteNextRecurringInstances(ctx, scope, testifyMock.Anything).
					Return(tc.mock.repo.deleteNextRecurringInstances.err)
			}

			if tc.mock.repo.updateRepeatUntilRecurringTrackings.isCalled {
				deps.repo.EXPECT().UpdateRepeatUntilRecurringTrackings(ctx, scope, testifyMock.Anything).
					Return(tc.mock.repo.updateRepeatUntilRecurringTrackings.err)
			}

			if tc.mock.repo.deleteRecurringTracking.isCalled {
				deps.repo.EXPECT().DeleteRecurringTracking(ctx, scope, testifyMock.Anything).
					Return(tc.mock.repo.deleteRecurringTracking.err)
			}

			if tc.mock.repo.updateRepeatUntil.isCalled {
				deps.repo.EXPECT().UpdateRepeatUntil(ctx, scope, tc.mock.repo.updateRepeatUntil.inputID, testifyMock.Anything).
					Return(tc.mock.repo.updateRepeatUntil.output, tc.mock.repo.updateRepeatUntil.err)
			}

			if tc.mock.room.updateTrackingException.isCalled {
				deps.roomUC.EXPECT().UpdateTrackingException(ctx, scope, tc.mock.room.updateTrackingException.input).
					Return(tc.mock.room.updateTrackingException.err)
			}

			if tc.mock.room.updateTrackingRepeatUntil.isCalled {
				deps.roomUC.EXPECT().UpdateTrackingRepeatUntil(ctx, scope, testifyMock.Anything).
					Return(tc.mock.room.updateTrackingRepeatUntil.err)
			}

			if tc.mock.room.deleteTracking.isCalled {
				deps.roomUC.EXPECT().DeleteTracking(ctx, scope, tc.mock.room.deleteTracking.input).
					Return(tc.mock.room.deleteTracking.err)
			}

			if tc.shop.detailShop.isCalled {
				deps.shopUC.EXPECT().DetailShop(ctx, scope, tc.shop.detailShop.input).
					Return(tc.shop.detailShop.output, tc.shop.detailShop.err)
			}

			if tc.shop.listAllUsers.isCalled {
				deps.shopUC.EXPECT().ListAllUsers(ctx, scope, testifyMock.Anything).
					Return(tc.shop.listAllUsers.output, tc.shop.listAllUsers.err).Maybe()
			}

			if tc.prod.publishPushNotiMsg.isCalled {
				deps.prod.EXPECT().PublishPushNotiMsg(ctx, tc.prod.publishPushNotiMsg.input).
					Return(tc.prod.publishPushNotiMsg.err)
			}

			if tc.getNotiHeading.isCalled {
				originalGetNotiHeading := notification.GetNotiHeading
				notification.GetNotiHeading = func(ctx context.Context, input notification.GetNotiHeadingInput) (string, error) {
					switch input.Lang {
					case "vi":
						return tc.getNotiHeading.wantVi, tc.getNotiHeading.errVi
					case "en":
						return tc.getNotiHeading.wantEn, tc.getNotiHeading.errEn
					case "ja":
						return tc.getNotiHeading.wantJa, tc.getNotiHeading.errJa
					default:
						return "", fmt.Errorf("unsupported language")
					}
				}
				defer func() { notification.GetNotiHeading = originalGetNotiHeading }()
			}

			if tc.getNotiContent.isCalled {
				originalGetNotiContent := notification.GetNotiContent
				notification.GetNotiContent = func(ctx context.Context, input notification.GetNotiContentInput) (string, error) {
					switch input.Lang {
					case "vi":
						return tc.getNotiContent.wantVi, tc.getNotiContent.errVi
					case "en":
						return tc.getNotiContent.wantEn, tc.getNotiContent.errEn
					case "ja":
						return tc.getNotiContent.wantJa, tc.getNotiContent.errJa
					default:
						return "", fmt.Errorf("unsupported language")
					}
				}
				defer func() { notification.GetNotiContent = originalGetNotiContent }()
			}

			err := uc.Delete(ctx, scope, tc.input)
			if tc.err != nil {
				require.Error(t, err)
				if errors.Is(tc.err, mongo.ErrNoDocuments) {
					require.True(t, errors.Is(err, mongo.ErrNoDocuments))
				} else {
					require.Equal(t, tc.err.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// func TestList(t *testing.T) {
// 	scope := models.Scope{
// 		Suffix: "test",
// 		ShopID: primitive.NewObjectID().Hex(),
// 		UserID: "test",
// 	}
// 	ctx := context.Background()

// 	type mockRepoList struct {
// 		isCalled bool
// 		input    event.ListOptions
// 		output   []models.Event
// 		err      error
// 	}

// 	type mockRepo struct {
// 		listEvents mockRepoList
// 	}

// 	type mock struct {
// 		repo mockRepo
// 	}

// 	expectedErr := errors.New("error")

// 	tcs := map[string]struct {
// 		input  event.ListInput
// 		mock   mock
// 		output event.ListOutput
// 		err    error
// 	}{
// 		"success": {
// 			input: event.ListInput{
// 				Filter: event.Filter{
// 					IDs: []string{"111111111111111111111111"},
// 				},
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					listEvents: mockRepoList{
// 						isCalled: true,
// 						input: event.ListOptions{
// 							Filter: event.Filter{
// 								IDs: []string{"111111111111111111111111"},
// 							},
// 						},
// 						output: []models.Event{
// 							{
// 								ID:    mongo.ObjectIDFromHexOrNil("111111111111111111111111"),
// 								Title: "test event",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			output: event.ListOutput{
// 				EventInstances: []event.EventInstance{
// 					{
// 						ID:    mongo.ObjectIDFromHexOrNil("111111111111111111111111"),
// 						Title: "test event",
// 					},
// 				},
// 			},
// 		},
// 		"err repo List": {
// 			input: event.ListInput{
// 				Filter: event.Filter{
// 					IDs: []string{"111111111111111111111111"},
// 				},
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					listEvents: mockRepoList{
// 						isCalled: true,
// 						input: event.ListOptions{
// 							Filter: event.Filter{
// 								IDs: []string{"111111111111111111111111"},
// 							},
// 						},
// 						err: expectedErr,
// 					},
// 				},
// 			},
// 			err: expectedErr,
// 		},
// 	}

// 	for n, tc := range tcs {
// 		t.Run(n, func(t *testing.T) {
// 			uc, deps := initUseCase(t)

// 			if tc.mock.repo.listEvents.isCalled {
// 				deps.repo.EXPECT().List(ctx, scope, tc.mock.repo.listEvents.input).
// 					Return(
// 						tc.mock.repo.listEvents.output,
// 						tc.mock.repo.listEvents.err,
// 					)
// 			}

// 			res, err := uc.List(ctx, scope, tc.input)
// 			if err != nil {
// 				require.EqualError(t, err, tc.err.Error())
// 			} else {
// 				require.Equal(t, tc.output, res)
// 			}
// 		})
// 	}
// }

// func TestGetOne(t *testing.T) {
// 	scope := models.Scope{
// 		Suffix: "test",
// 		ShopID: primitive.NewObjectID().Hex(),
// 		UserID: "test",
// 	}
// 	ctx := context.Background()

// 	expectedErr := errors.New("some error")

// 	type mockRepoGetOne struct {
// 		isCalled bool
// 		input    event.GetOneOptions
// 		output   models.Event
// 		err      error
// 	}

// 	type mockRepo struct {
// 		getOneEvent mockRepoGetOne
// 	}

// 	type mock struct {
// 		repo mockRepo
// 	}

// 	tcs := map[string]struct {
// 		input  event.GetOneInput
// 		mock   mock
// 		output models.Event
// 		err    error
// 	}{
// 		"success": {
// 			input: event.GetOneInput{
// 				Filter: event.Filter{
// 					ID: "111111111111111111111111",
// 				},
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					getOneEvent: mockRepoGetOne{
// 						isCalled: true,
// 						input: event.GetOneOptions{
// 							Filter: event.Filter{
// 								ID: "111111111111111111111111",
// 							},
// 						},
// 						output: models.Event{
// 							ID: mongo.ObjectIDFromHexOrNil("111111111111111111111111"),
// 						},
// 					},
// 				},
// 			},
// 			output: models.Event{
// 				ID: mongo.ObjectIDFromHexOrNil("111111111111111111111111"),
// 			},
// 		},
// 		"err repo GetOne": {
// 			input: event.GetOneInput{
// 				Filter: event.Filter{
// 					ID: "111111111111111111111111",
// 				},
// 			},
// 			mock: mock{
// 				repo: mockRepo{
// 					getOneEvent: mockRepoGetOne{
// 						isCalled: true,
// 						input: event.GetOneOptions{
// 							Filter: event.Filter{
// 								ID: "111111111111111111111111",
// 							},
// 						},
// 						err: expectedErr,
// 					},
// 				},
// 			},
// 			err: expectedErr,
// 		},
// 	}

// 	for n, tc := range tcs {
// 		t.Run(n, func(t *testing.T) {
// 			uc, deps := initUseCase(t)

// 			if tc.mock.repo.getOneEvent.isCalled {
// 				deps.repo.EXPECT().GetOne(ctx, scope, tc.mock.repo.getOneEvent.input).
// 					Return(
// 						tc.mock.repo.getOneEvent.output,
// 						tc.mock.repo.getOneEvent.err,
// 					)
// 			}

// 			res, err := uc.GetOne(ctx, scope, tc.input)
// 			if err != nil {
// 				require.EqualError(t, err, tc.err.Error())
// 			} else {
// 				require.Equal(t, tc.output, res)
// 			}
// 		})
// 	}
// }
