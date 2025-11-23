package usecase

import (
	"context"
	"slices"
	"sort"

	"gitlab.com/gma-vietnam/tanca-connect/internal/element"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/eventcategory"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/internal/resources/notification"
	"gitlab.com/gma-vietnam/tanca-connect/internal/room"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"golang.org/x/sync/errgroup"
)

func (uc implUseCase) Create(ctx context.Context, sc models.Scope, input event.CreateInput) (event.CreateEventOutput, error) {
	if len(input.DepartmentIDs) > 0 || len(input.AssignIDs) > 0 && len(input.BranchIDs) == 0 {
		sessUser, err := uc.shopUC.GetSessionUser(ctx, sc)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Create.GetSessionUser: %v", err)
			return event.CreateEventOutput{}, err
		}

		input.BranchIDs = append(input.BranchIDs, sessUser.BranchID)
		if len(sessUser.BranchPlusIds) > 0 {
			input.BranchIDs = append(input.BranchIDs, sessUser.BranchPlusIds...)
		}
	}

	err := uc.validateAssign(ctx, sc, input.AssignIDs, input.BranchIDs, input.DepartmentIDs)
	if err != nil {
		uc.l.Warnf(ctx, "event.usecase.Create.validateAssign: %v", err)
		return event.CreateEventOutput{}, err
	}

	if len(input.RoomIDs) > 0 {
		allUnvalidRoomIDs, err := uc.roomUC.GetUnavailableRooms(ctx, sc, room.GetUnavailableRoomsInput{
			StartTime:   input.StartTime,
			EndTime:     input.EndTime,
			Repeat:      input.Repeat,
			RepeatUntil: input.RepeatUntil,
			TimezoneID:  input.TimezoneID,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Create.GetUnavailableRooms: %v", err)
			return event.CreateEventOutput{}, err
		}

		unavailableRoomIDs := make([]string, 0, len(allUnvalidRoomIDs))
		for _, roomID := range input.RoomIDs {
			if slices.Contains(allUnvalidRoomIDs, roomID) {
				unavailableRoomIDs = append(unavailableRoomIDs, roomID)
			}
		}

		if len(unavailableRoomIDs) > 0 {
			unavailableRooms, err := uc.roomUC.List(ctx, sc, room.ListInput{
				Filter: room.Filter{
					IDs: unavailableRoomIDs,
				},
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Create.ListRooms: %v", err)
				return event.CreateEventOutput{}, err
			}

			return event.CreateEventOutput{
				UnavailableRooms: unavailableRooms,
			}, event.ErrRoomUnavailable

		}
	}

	tz, err := uc.elementUC.Detail(ctx, sc, input.TimezoneID)
	if err != nil {
		if err == element.ErrElementNotFound {
			uc.l.Warnf(ctx, "event.usecase.Create.GetTimezone: %v", err)
			return event.CreateEventOutput{}, event.ErrTimezoneNotFound
		}
		uc.l.Errorf(ctx, "event.usecase.Create.GetTimezone: %v", err)
		return event.CreateEventOutput{}, err
	}

	opts := repository.CreateOptions{
		Title:         input.Title,
		BranchIDs:     input.BranchIDs,
		AssignIDs:     input.AssignIDs,
		DepartmentIDs: input.DepartmentIDs,
		TimezoneID:    input.TimezoneID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		AllDay:        input.AllDay,
		Repeat:        input.Repeat,
		RoomIDs:       input.RoomIDs,
		Description:   input.Description,
		CategoryID:    input.CategoryID,
		RepeatUntil:   input.RepeatUntil,
		Notify:        input.Notify,
		System:        input.System,
		Alert:         input.Alert,
		ObjectID:      input.ObjectID,
		Public:        input.Public,
	}

	if input.Notify {
		opts.NotifyTime = uc.calculateNotifyTimeForEvent(input.StartTime, input.AllDay, input.Alert, *tz.Offset)
	}

	e, err := uc.repo.Create(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Create.Create: %v", err)
		return event.CreateEventOutput{}, err
	}

	ei := event.EventToEventInstance(sc, e)

	eg := new(errgroup.Group)

	// Handle recurring instances generation
	if e.Repeat != models.EventRepeatNone {
		eg.Go(func() error {
			_, err := uc.generateRecurringInstances(ctx, sc, e, e.StartTime, &tz)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Create.generateRecurringInstances: %v", err)
				return err
			}
			return nil
		})
	}

	// Handle notifications
	if input.Notify {
		eg.Go(func() error {
			depIDs := mongo.HexFromObjectIDsOrNil(e.DepartmentIDs)
			branchIDs := mongo.HexFromObjectIDsOrNil(e.BranchIDs)
			uIDs, err := uc.getAssignUserIDs(ctx, sc, depIDs, branchIDs, e.AssignIDs)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Create.getAssignUserIDs: %v", err)
				return err
			}
			uIDs = append(uIDs, e.CreatedByID)

			shop, err := uc.shopUC.DetailShop(ctx, sc, sc.ShopID)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Create.DetailShop: %v", err)
				return err
			}

			timeText, dateText := uc.formatEventDateTime(ctx, ei, shop)

			if len(uIDs) > 0 {
				uIDs = util.RemoveDuplicates(uIDs)
				noti, err := uc.getEventNoti(ctx, sc, getEventNotiInput{
					EI:       ei,
					Type:     notification.SourceEventAssign,
					UserIDs:  uIDs,
					TimeText: timeText,
					DateText: dateText,
				})
				if err != nil {
					uc.l.Errorf(ctx, "event.usecase.Create.getEventNoti: %v", err)
					return err
				}

				err = uc.publishPushNotiMsg(ctx, noti)
				if err != nil {
					uc.l.Errorf(ctx, "event.usecase.Create.publishPushNotiMsg: %v", err)
					return err
				}
			}
			return nil
		})
	}

	// Handle room trackings
	if len(input.RoomIDs) > 0 {
		eg.Go(func() error {
			_, err := uc.roomUC.CreateTrackings(ctx, sc, room.CreateTrackingsInput{
				RoomIDs:     input.RoomIDs,
				EventID:     e.ID.Hex(),
				StartTime:   util.ConvertToLocalTimezone(e.StartTime, *tz.Offset),
				EndTime:     util.ConvertToLocalTimezone(e.EndTime, *tz.Offset),
				Repeat:      e.Repeat,
				RepeatUntil: e.RepeatUntil,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Create.CreateTrackings: %v", err)
				return err
			}
			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := eg.Wait(); err != nil {
		return event.CreateEventOutput{}, err
	}

	return event.CreateEventOutput{
		EventInstance: ei,
	}, nil
}

func (uc implUseCase) Detail(ctx context.Context, sc models.Scope, id string, eventID string) (event.DetailOutput, error) {
	var (
		ei   event.EventInstance
		base event.EventInstance
	)

	eg := new(errgroup.Group)
	var e, baseE models.Event
	var ri models.RecurringInstance

	// Try to get main event details
	eg.Go(func() error {
		var err error
		e, err = uc.repo.Detail(ctx, sc, id)
		if err != nil && err != mongo.ErrNoDocuments {
			uc.l.Errorf(ctx, "event.usecase.Detail.Detail: %v", err)
			return err
		}
		return nil
	})

	// Try to get base event details
	eg.Go(func() error {
		var err error
		baseE, err = uc.repo.Detail(ctx, sc, eventID)
		if err != nil && err != mongo.ErrNoDocuments {
			uc.l.Errorf(ctx, "event.usecase.Detail.Detail: %v", err)
			return err
		}
		return nil
	})

	// Try to get recurring instance details
	eg.Go(func() error {
		var err error
		ri, err = uc.repo.DetailRecurringInstance(ctx, sc, id, eventID)
		if err != nil && err != mongo.ErrNoDocuments {
			uc.l.Errorf(ctx, "event.usecase.Detail.DetailRecurringInstance: %v", err)
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return event.DetailOutput{}, err
	}

	// Process results
	if !e.ID.IsZero() {
		// Main event exists
		ei = event.EventToEventInstance(sc, e)
	} else if !baseE.ID.IsZero() && !ri.ID.IsZero() {
		// Base event and recurring instance exist
		base = event.EventToEventInstance(sc, baseE)
		ei = event.RecurringInstanceToEventInstance(sc, ri)
	} else {
		// Neither main event nor recurring instance found
		uc.l.Errorf(ctx, "event.usecase.Detail: event not found")
		return event.DetailOutput{}, event.ErrEventNotFound
	}

	branchIDs := mongo.HexFromObjectIDsOrNil(ei.BranchIDs)
	deptIDs := mongo.HexFromObjectIDsOrNil(ei.DepartmentIDs)

	assignIDs, err := uc.getAssignUserIDs(ctx, sc, deptIDs, branchIDs, ei.AssignIDs)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Detail.getAssignUserIDs: %v", err)
		return event.DetailOutput{}, err
	}

	if !util.Contains(assignIDs, sc.UserID) && sc.UserID != ei.CreatedByID && !ei.Public {
		uc.l.Warnf(ctx, "event.usecase.Detail.getAssignUserIDs: %v", event.ErrCanNotViewEvent)
		return event.DetailOutput{}, event.ErrCanNotViewEvent
	}

	var (
		users      []microservice.User
		depts      []microservice.Department
		rooms      []models.Room
		categories []models.EventCategory
		timezones  []models.Element
		branches   []microservice.Branch
	)

	eg = new(errgroup.Group)

	eg.Go(func() error {
		userIDs := append(ei.AssignIDs, ei.CreatedByID)

		if !base.ID.IsZero() {
			userIDs = append(userIDs, base.AssignIDs...)
			userIDs = append(userIDs, base.CreatedByID)
		}

		userIDs = util.RemoveDuplicates(userIDs)

		if len(userIDs) > 0 {
			var err error
			users, err = uc.shopUC.ListAllUsers(ctx, sc, microservice.GetUsersFilter{
				IDs: userIDs,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Detail.ListAllUsers: %v", err)
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		deptIDs := make([]string, 0, len(ei.DepartmentIDs))

		for _, id := range ei.DepartmentIDs {
			deptIDs = append(deptIDs, id.Hex())
		}

		if !base.ID.IsZero() {
			for _, id := range base.DepartmentIDs {
				deptIDs = append(deptIDs, id.Hex())
			}
		}

		deptIDs = util.RemoveDuplicates(deptIDs)
		if len(deptIDs) > 0 {
			var err error
			depts, err = uc.shopUC.GetDepartments(ctx, sc, microservice.GetDepartmentsFilter{IDs: deptIDs})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Detail.GetDepartments: %v", err)
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		roomIDs := make([]string, 0, len(ei.RoomIDs))

		for _, id := range ei.RoomIDs {
			roomIDs = append(roomIDs, id.Hex())
		}

		if !base.ID.IsZero() {
			for _, id := range base.RoomIDs {
				roomIDs = append(roomIDs, id.Hex())
			}
		}

		roomIDs = util.RemoveDuplicates(roomIDs)
		if len(roomIDs) > 0 {
			var err error
			rooms, err = uc.roomUC.List(ctx, sc, room.ListInput{
				Filter: room.Filter{
					IDs: roomIDs,
				},
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Detail.ListRooms: %v", err)
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		cateIDs := []string{}

		if ei.CategoryID != nil {
			cateIDs = append(cateIDs, ei.CategoryID.Hex())
		}

		if base.CategoryID != nil {
			cateIDs = append(cateIDs, base.CategoryID.Hex())
		}

		cateIDs = util.RemoveDuplicates(cateIDs)

		if len(cateIDs) > 0 {
			var err error
			categories, err = uc.eventcategoryUC.List(ctx, sc, eventcategory.ListInput{
				Filter: eventcategory.Filter{
					IDs: cateIDs,
				},
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Detail.ListCategories: %v", err)
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		timezoneIDs := []string{}

		if !ei.TimezoneID.IsZero() {
			timezoneIDs = append(timezoneIDs, ei.TimezoneID.Hex())
		}

		if !base.TimezoneID.IsZero() {
			timezoneIDs = append(timezoneIDs, base.TimezoneID.Hex())
		}

		timezoneIDs = util.RemoveDuplicates(timezoneIDs)

		if len(timezoneIDs) > 0 {
			var err error
			timezones, err = uc.elementUC.List(ctx, sc, element.ListInput{Filter: element.Filter{IDs: timezoneIDs}})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Detail.ListTimezones: %v", err)
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		branchIDs := make([]string, 0, len(ei.BranchIDs))
		for _, id := range ei.BranchIDs {
			branchIDs = append(branchIDs, id.Hex())
		}

		if !base.ID.IsZero() {
			for _, id := range base.BranchIDs {
				branchIDs = append(branchIDs, id.Hex())
			}
		}

		branchIDs = util.RemoveDuplicates(branchIDs)

		if len(branchIDs) > 0 {
			var err error
			branches, err = uc.shopUC.GetBranches(ctx, sc, microservice.GetBranchesFilter{IDs: branchIDs})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Detail.ListBranches: %v", err)
				return err
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return event.DetailOutput{}, err
	}

	return event.DetailOutput{
		EventInstance: ei,
		EventCategory: categories,
		Timezones:     timezones,
		Users:         users,
		Departments:   depts,
		Rooms:         rooms,
		BaseEvent:     base,
		Branches:      branches,
	}, nil
}

func (uc implUseCase) Update(ctx context.Context, sc models.Scope, input event.UpdateInput) error {
	err := uc.validateAssign(ctx, sc, input.AssignIDs, input.BranchIDs, input.DepartmentIDs)
	if err != nil {
		uc.l.Warnf(ctx, "event.usecase.Update.validateAssign: %v", err)
		return err
	}

	if input.CategoryID != "" {
		_, err := uc.eventcategoryUC.Detail(ctx, sc, input.CategoryID)
		if err != nil {
			if err == eventcategory.ErrEventCategoryNotFound {
				uc.l.Warnf(ctx, "event.usecase.Update.Detail: %v", err)
				return err
			}
			uc.l.Errorf(ctx, "event.usecase.Update.Detail: %v", err)
			return err
		}
	}

	if input.Type != "" {
		ri, err := uc.repo.DetailRecurringInstance(ctx, sc, input.ID, input.EventID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				uc.l.Warnf(ctx, "event.usecase.Update.DetailRecurringInstance: %v", err)
				return event.ErrEventNotFound
			}
			uc.l.Errorf(ctx, "event.usecase.Update.DetailRecurringInstance: %v", err)
			return err
		}

		if ri.CreatedByID != sc.UserID {
			uc.l.Warnf(ctx, "event.usecase.Update.Update: %v", event.ErrEventEditNotAllowed)
			return event.ErrEventEditNotAllowed
		}

		switch input.Type {
		case models.EventActionOne:
			err = uc.handleUpdateSingleEventInstance(ctx, sc, input, ri)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.handleUpdateSingleEventInstance: %v", err)
				return err
			}
		case models.EventActionFrom:
			err = uc.handleUpdateEventFromDate(ctx, sc, input, ri)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.handleUpdateEventFromDate: %v", err)
				return err
			}
		case models.EventActionAll:
			err = uc.handleUpdateAllEventInstances(ctx, sc, input, ri)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.handleUpdateAllEventInstances: %v", err)
				return err
			}
		}
	} else {
		e, err := uc.repo.Detail(ctx, sc, input.ID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				uc.l.Warnf(ctx, "event.usecase.Update.Detail: %v", err)
				return event.ErrEventNotFound
			}
			return err
		}

		if e.CreatedByID != sc.UserID {
			uc.l.Warnf(ctx, "event.usecase.Update.Update: %v", event.ErrEventEditNotAllowed)
			return event.ErrEventEditNotAllowed
		}

		cOpts := repository.UpdateOptions{
			ID:            input.ID,
			Model:         e,
			Title:         input.Title,
			BranchIDs:     input.BranchIDs,
			AssignIDs:     input.AssignIDs,
			DepartmentIDs: input.DepartmentIDs,
			TimezoneID:    input.TimezoneID,
			StartTime:     input.StartTime,
			EndTime:       input.EndTime,
			AllDay:        input.AllDay,
			Repeat:        input.Repeat,
			RoomIDs:       input.RoomIDs,
			Description:   input.Description,
			CategoryID:    input.CategoryID,
			RepeatUntil:   input.RepeatUntil,
			Notify:        input.Notify,
			Alert:         input.Alert,
			Public:        input.Public,
		}

		tz, err := uc.elementUC.Detail(ctx, sc, input.TimezoneID)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.GetTimezone: %v", err)
			return err
		}

		if input.Notify && input.Alert != nil {
			cOpts.NotifyTime = uc.calculateNotifyTimeForEvent(input.StartTime, input.AllDay, input.Alert, *tz.Offset)
		}

		ue, err := uc.repo.Update(ctx, sc, cOpts)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.Update: %v", err)
			return err
		}

		shop, err := uc.shopUC.DetailShop(ctx, sc, sc.ShopID)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.DetailShop: %v", err)
			return err
		}

		err = uc.handleUpdateEventNotifications(ctx, sc, event.EventToEventInstance(sc, e), input, shop)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.handleUpdateEventNotifications: %v", err)
			return err
		}

		if ue.Repeat != models.EventRepeatNone {
			_, err = uc.generateRecurringInstances(ctx, sc, ue, ue.StartTime, &tz)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.generateRecurringInstances: %v", err)
				return err
			}
		}

		if len(e.RoomIDs) > 0 {
			err = uc.roomUC.DeleteTracking(ctx, sc, room.DeleteTrackingInput{
				RoomIDs: mongo.HexFromObjectIDsOrNil(e.RoomIDs),
				EventID: e.ID.Hex(),
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.DeleteTracking: %v", err)
				return err
			}
		}

		if len(ue.RoomIDs) > 0 {
			_, err = uc.roomUC.CreateTrackings(ctx, sc, room.CreateTrackingsInput{
				RoomIDs:     mongo.HexFromObjectIDsOrNil(ue.RoomIDs),
				EventID:     ue.ID.Hex(),
				StartTime:   ue.StartTime,
				EndTime:     ue.EndTime,
				Repeat:      ue.Repeat,
				RepeatUntil: ue.RepeatUntil,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.CreateTrackings: %v", err)
				return err
			}
		}

		return nil
	}

	return nil
}

func (uc implUseCase) Delete(ctx context.Context, sc models.Scope, input event.DeleteInput) error {
	if input.Type != "" {
		ri, err := uc.repo.DetailRecurringInstance(ctx, sc, input.ID, input.EventID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				uc.l.Warnf(ctx, "event.usecase.Delete.DetailRecurringInstance: %v", err)
				return event.ErrEventNotFound
			}
			uc.l.Errorf(ctx, "event.usecase.Delete.DetailRecurringInstance: %v", err)
			return err
		}

		if ri.CreatedByID != sc.UserID {
			uc.l.Warnf(ctx, "event.usecase.Delete.Delete: %v", event.ErrEventEditNotAllowed)
			return event.ErrEventEditNotAllowed
		}

		switch input.Type {
		case models.EventActionOne:
			err = uc.handleDeleteSingleEventInstance(ctx, sc, input, ri)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Delete.handleDeleteSingleEventInstance: %v", err)
				return err
			}
		case models.EventActionFrom, models.EventActionFollowing:
			err = uc.handleDeleteEventFromDate(ctx, sc, input, ri)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Delete.handleDeleteEventFromDate: %v", err)
				return err
			}
		case models.EventActionAll:
			err = uc.handleDeleteAllEventInstances(ctx, sc, input, ri)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Delete.handleDeleteAllEventInstances: %v", err)
				return err
			}
		}
	} else {
		e, err := uc.repo.Detail(ctx, sc, input.ID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				uc.l.Warnf(ctx, "event.usecase.Delete.Detail: %v", err)
				return event.ErrEventNotFound
			}
			uc.l.Errorf(ctx, "event.usecase.Delete.Detail: %v", err)
			return err
		}

		if e.CreatedByID != sc.UserID {
			uc.l.Warnf(ctx, "event.usecase.Delete.Delete: %v", event.ErrEventEditNotAllowed)
			return event.ErrEventEditNotAllowed
		}

		eg := new(errgroup.Group)

		eg.Go(func() error {
			err = uc.repo.Delete(ctx, sc, input.ID)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Delete.Delete: %v", err)
				return err
			}
			return nil
		})

		if len(e.RoomIDs) > 0 {
			eg.Go(func() error {
				err = uc.roomUC.DeleteTracking(ctx, sc, room.DeleteTrackingInput{
					RoomIDs: mongo.HexFromObjectIDsOrNil(e.RoomIDs),
					EventID: e.ID.Hex(),
				})
				if err != nil {
					uc.l.Errorf(ctx, "event.usecase.Delete.DeleteTracking: %v", err)
					return err
				}
				return nil
			})
		}

		eg.Go(func() error {
			shop, err := uc.shopUC.DetailShop(ctx, sc, sc.ShopID)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Delete.DetailShop: %v", err)
				return err
			}

			err = uc.handleEventDeleteNotification(ctx, sc, event.EventToEventInstance(sc, e), shop, input.Type)
			if err != nil {
				return err
			}
			return nil
		})

		if err := eg.Wait(); err != nil {
			return err
		}
	}

	return nil
}

func (uc implUseCase) Calendar(ctx context.Context, sc models.Scope, input event.CalendarInput) (event.CalendarOutput, error) {
	sessUser, err := uc.shopUC.GetSessionUser(ctx, sc)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Calendar.GetSessionUser: %v", err)
		return event.CalendarOutput{}, err
	}

	exCat, err := uc.elementUC.ListShopElement(ctx, sc, element.ListShopElementInput{
		ShopElementFilter: element.ShopElementFilter{
			Type:   models.ElementTypeEventCategory,
			Status: util.ToPointer(false),
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Calendar.ListShopElement: %v", err)
		return event.CalendarOutput{}, err
	}
	exCatIDs := make([]string, 0, len(exCat))
	for _, cat := range exCat {
		exCatIDs = append(exCatIDs, cat.ElementID.Hex())
	}

	branchIDs := []string{}
	if sessUser.BranchID != "" {
		branchIDs = append(branchIDs, sessUser.BranchID)
	}
	if len(sessUser.BranchPlusIds) > 0 {
		branchIDs = append(branchIDs, sessUser.BranchPlusIds...)
	}

	departmentIDs := []string{}
	if len(sessUser.DepartmentPlusIds) > 0 {
		departmentIDs = append(departmentIDs, sessUser.DepartmentPlusIds...)
	}
	if sessUser.DepartmentID != "" {
		departmentIDs = append(departmentIDs, sessUser.DepartmentID)
	}

	nres, err := uc.repo.List(ctx, sc, repository.ListOptions{
		Filter: repository.Filter{
			IDs:                input.IDs,
			NeedRepeat:         util.ToPointer(false),
			StartTime:          input.StartTime,
			EndTime:            input.EndTime,
			BranchIDs:          branchIDs,
			DepartmentIDs:      departmentIDs,
			ExcludeCategoryIDs: exCatIDs,
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Calendar.List: %v", err)
		return event.CalendarOutput{}, err
	}

	eis := []event.EventInstance{}
	for _, e := range nres {
		eis = append(eis, event.EventToEventInstance(sc, e))
	}

	riis, err := uc.getRecurringInstanceInDateRange(ctx, sc, input.StartTime, input.EndTime)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Calendar.getRecurringInstanceInDateRange: %v", err)
		return event.CalendarOutput{}, err
	}

	riis = uc.filterInstances(riis, sc, departmentIDs, branchIDs, exCatIDs)

	eeis := make([]event.EventInstance, 0, len(riis))
	for _, e := range riis {
		eeis = append(eeis, event.RecurringInstanceToEventInstance(sc, e))
	}

	eis = append(eis, eeis...)

	// sort by start time
	sort.Slice(eis, func(i, j int) bool {
		return eis[i].StartTime.Before(eis[j].StartTime)
	})

	var (
		users []microservice.User
		ecs   []models.EventCategory
		tzs   []models.Element
	)

	ecIDs := []string{}
	for _, e := range eis {
		if e.CategoryID != nil {
			ecIDs = append(ecIDs, e.CategoryID.Hex())
		}
	}
	ecIDs = util.RemoveDuplicates(ecIDs)

	userIDs := []string{}
	for _, e := range eis {
		userIDs = append(userIDs, e.AssignIDs...)
		userIDs = append(userIDs, e.CreatedByID)
	}
	userIDs = util.RemoveDuplicates(userIDs)

	eg := new(errgroup.Group)

	// Fetch event categories concurrently
	eg.Go(func() error {
		var err error
		ecs, err = uc.eventcategoryUC.List(ctx, sc, eventcategory.ListInput{
			Filter: eventcategory.Filter{
				IDs: ecIDs,
			},
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.List.ListEventCategories: %v", err)
			return err
		}
		return nil
	})

	tzIDs := []string{}
	for _, e := range eis {
		tzIDs = append(tzIDs, e.TimezoneID.Hex())
	}
	tzIDs = util.RemoveDuplicates(tzIDs)

	eg.Go(func() error {
		var err error
		tzs, err = uc.elementUC.List(ctx, sc, element.ListInput{
			Filter: element.Filter{
				IDs:  tzIDs,
				Type: models.ElementTypeTimezone,
			},
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.List.ListTimezones: %v", err)
			return err
		}
		return nil
	})

	// Fetch users concurrently if there are any userIDs
	if len(userIDs) > 0 {
		eg.Go(func() error {
			var err error
			users, err = uc.shopUC.ListAllUsers(ctx, sc, microservice.GetUsersFilter{
				IDs: userIDs,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.List.ListAllUsers: %v", err)
				return err
			}
			return nil
		})
	}

	// Wait for all goroutines to complete and check for errors
	if err := eg.Wait(); err != nil {
		return event.CalendarOutput{}, err
	}

	return event.CalendarOutput{
		EventInstances:  eis,
		EventCategories: ecs,
		Users:           users,
		Timezones:       tzs,
	}, nil
}

func (uc implUseCase) GetOne(ctx context.Context, sc models.Scope, input event.GetOneInput) (event.EventInstance, error) {
	e, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{
		Filter: repository.Filter{
			ID: input.ID,
		},
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ri, err := uc.repo.DetailRecurringInstance(ctx, sc, input.ID, input.EventID)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.GetOne.DetailRecurringInstance: %v", err)
				return event.EventInstance{}, err
			}
			return event.RecurringInstanceToEventInstance(sc, ri), nil
		}
		uc.l.Errorf(ctx, "event.usecase.GetOne.GetOne: %v", err)
		return event.EventInstance{}, err
	}

	return event.EventToEventInstance(sc, e), nil
}

func (uc implUseCase) UpdateAttendance(ctx context.Context, sc models.Scope, input event.UpdateAttendanceInput) error {
	_, err := uc.repo.Detail(ctx, sc, input.ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			_, err := uc.repo.DetailRecurringInstance(ctx, sc, input.ID, input.EventID)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.UpdateAttendance.DetailRecurringInstance: %v", err)
				return err
			}
			err = uc.repo.UpdateAttendanceRecurringInstance(ctx, sc, repository.UpdateAttendanceRecurringInstanceOptions{
				ID:      input.ID,
				EventID: input.EventID,
				Status:  input.Status,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.UpdateAttendance.UpdateAttendanceRecurringInstance: %v", err)
				return err
			}
			return nil
		}
		uc.l.Errorf(ctx, "event.usecase.UpdateAttendance.Detail: %v", err)
		return err
	}

	err = uc.repo.UpdateAttendance(ctx, sc, input.EventID, input.Status)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.UpdateAttendance.UpdateAttendance: %v", err)
		return err
	}
	return nil
}

func (uc implUseCase) ListByIDs(ctx context.Context, sc models.Scope, ids []string) ([]models.Event, error) {
	es, err := uc.repo.ListByIDs(ctx, sc, ids)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "event.usecase.ListByIDs.ListByIDs: %v", err)
			return nil, event.ErrEventNotFound
		}
		uc.l.Errorf(ctx, "event.usecase.ListByIDs.ListByIDs: %v", err)
		return nil, err
	}
	return es, nil
}
