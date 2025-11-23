package usecase

import (
	"context"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/internal/room"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"golang.org/x/sync/errgroup"
)

// handleUpdateSingleEventInstance handles updating a single event instance (EventActionOne)
func (uc implUseCase) handleUpdateSingleEventInstance(ctx context.Context, sc models.Scope, input event.UpdateInput, ri models.RecurringInstance) error {
	uOpts := repository.UpdateRecurringInstanceOptions{
		ID:            input.ID,
		Model:         ri,
		EventID:       input.EventID,
		Title:         input.Title,
		BranchIDs:     input.BranchIDs,
		AssignIDs:     input.AssignIDs,
		DepartmentIDs: input.DepartmentIDs,
		TimezoneID:    input.TimezoneID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		AllDay:        input.AllDay,
		RoomIDs:       input.RoomIDs,
		CategoryID:    input.CategoryID,
		Description:   input.Description,
		Notify:        input.Notify,
		Alert:         input.Alert,
		Public:        input.Public,
	}

	if input.Notify && input.Alert != nil {
		tz, err := uc.elementUC.Detail(ctx, sc, input.TimezoneID)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.Detail: %v", err)
			return err
		}
		uOpts.NotifyTime = uc.calculateNotifyTimeForEvent(input.StartTime, input.AllDay, input.Alert, *tz.Offset)
	}

	nri, err := uc.repo.UpdateRecurringInstance(ctx, sc, uOpts)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Update.UpdateRecurringInstance: %v", err)
		return err
	}

	if len(input.RoomIDs) > 0 && (input.StartTime != nri.StartTime || input.EndTime != nri.EndTime) {
		eg := new(errgroup.Group)

		// Delete existing tracking
		eg.Go(func() error {
			err := uc.roomUC.DeleteTrackingByInstanceID(ctx, sc, ri.ID.Hex())
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.DeleteTrackingByInstanceID: %v", err)
				return err
			}
			return nil
		})

		createdRoomIDs := []string{}
		updatedAndRemovedRoomIDs := []string{}

		// Find rooms that need to be created (in nri but not in ri)
		for _, id := range nri.RoomIDs {
			if !util.Contains(ri.RoomIDs, id) {
				createdRoomIDs = append(createdRoomIDs, id.Hex())
			}
		}

		// Find rooms that need to be updated (in both nri and ri)
		for _, id := range nri.RoomIDs {
			if util.Contains(ri.RoomIDs, id) {
				updatedAndRemovedRoomIDs = append(updatedAndRemovedRoomIDs, id.Hex())
			}
		}

		for _, id := range ri.RoomIDs {
			if !util.Contains(nri.RoomIDs, id) {
				updatedAndRemovedRoomIDs = append(updatedAndRemovedRoomIDs, id.Hex())
			}
		}

		if len(createdRoomIDs) > 0 {
			eg.Go(func() error {
				_, err = uc.roomUC.CreateTrackings(ctx, sc, room.CreateTrackingsInput{
					RoomIDs:     createdRoomIDs,
					EventID:     nri.EventID.Hex(),
					InstanceID:  nri.ID.Hex(),
					StartTime:   nri.StartTime,
					EndTime:     nri.EndTime,
					Repeat:      models.EventRepeatNone,
					RepeatUntil: nri.RepeatUntil,
				})
				if err != nil {
					uc.l.Errorf(ctx, "event.usecase.Update.CreateTrackings: %v", err)
					return err
				}
				return nil
			})
		}

		if len(updatedAndRemovedRoomIDs) > 0 {
			eg.Go(func() error {
				err = uc.roomUC.UpdateTrackingException(ctx, sc, room.UpdateTrackingExceptionInput{
					RoomIDs:   updatedAndRemovedRoomIDs,
					EventID:   nri.EventID.Hex(),
					StartTime: nri.StartTime,
					EndTime:   nri.EndTime,
				})
				if err != nil {
					uc.l.Errorf(ctx, "event.usecase.Update.UpdateTrackingRepeatUntil: %v", err)
					return err
				}
				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			return err
		}
	}

	shop, err := uc.shopUC.DetailShop(ctx, sc, sc.ShopID)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Update.DetailShop: %v", err)
		return err
	}

	err = uc.handleUpdateEventNotifications(ctx, sc, event.RecurringInstanceToEventInstance(sc, ri), input, shop)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Update.handleUpdateEventNotifications: %v", err)
		return err
	}

	return nil
}

// handleUpdateEventFromDate handles updating event from a specific date (EventActionFrom)
func (uc implUseCase) handleUpdateEventFromDate(ctx context.Context, sc models.Scope, input event.UpdateInput, ri models.RecurringInstance) error {
	// Fetch event details
	e, err := uc.repo.Detail(ctx, sc, ri.EventID.Hex())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "event.usecase.Update.Detail: %v", err)
			return event.ErrEventNotFound
		}
		uc.l.Errorf(ctx, "event.usecase.Update.Detail: %v", err)
		return err
	}

	// Create new event
	cOpts := repository.CreateOptions{
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
		CategoryID:    input.CategoryID,
		RepeatUntil:   input.RepeatUntil,
		Description:   input.Description,
		Notify:        input.Notify,
		Public:        input.Public,
	}

	ne, err := uc.repo.Create(ctx, sc, cOpts)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Update.Create: %v", err)
		return err
	}

	tz, err := uc.elementUC.Detail(ctx, sc, e.TimezoneID.Hex())
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Update.GetTimezone: %v", err)
		return err
	}

	if input.Notify && input.Alert != nil {
		cOpts.NotifyTime = uc.calculateNotifyTimeForEvent(input.StartTime, input.AllDay, input.Alert, *tz.Offset)
	}

	prevStartTime, _ := uc.getPreviousOccurrence(ri.StartTime, ri.EndTime, ri.Repeat)

	eg := new(errgroup.Group)

	// Delete next recurring instances
	eg.Go(func() error {
		err := uc.repo.DeleteNextRecurringInstances(ctx, sc, repository.DeleteNextRecurringInstancesOptions{
			EventID: e.ID.Hex(),
			Date:    prevStartTime,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.DeleteNextRecurringInstances: %v", err)
			return err
		}
		return nil
	})

	// Generate recurring instances if needed
	if input.Repeat != models.EventRepeatNone {
		eg.Go(func() error {
			_, err := uc.generateRecurringInstances(ctx, sc, ne, input.StartTime, &tz)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.generateRecurringInstances: %v", err)
				return err
			}
			return nil
		})
	}

	// Update repeat until recurring trackings
	eg.Go(func() error {
		err := uc.repo.UpdateRepeatUntilRecurringTrackings(ctx, sc, repository.UpdateRepeatUntilRecurringTrackingsOptions{
			EventID:     ri.EventID.Hex(),
			RepeatUntil: &prevStartTime,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.UpdateRepeatUntilRecurringTrackings: %v", err)
			return err
		}
		return nil
	})

	// Patch the event
	eg.Go(func() error {
		_, err := uc.repo.UpdateRepeatUntil(ctx, sc, ri.EventID.Hex(), prevStartTime)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.Patch: %v", err)
			return err
		}
		return nil
	})

	// Wait for all operations to complete
	if err := eg.Wait(); err != nil {
		return err
	}

	// Handle room tracking updates
	eg = new(errgroup.Group)

	eg.Go(func() error {
		err = uc.roomUC.UpdateTrackingRepeatUntil(ctx, sc, room.UpdateTrackingRepeatUntilInput{
			RoomIDs:     mongo.HexFromObjectIDsOrNil(ri.RoomIDs),
			EventID:     ri.EventID.Hex(),
			RepeatUntil: prevStartTime,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.UpdateTrackingRepeatUntil: %v", err)
			return err
		}
		return nil
	})

	if len(input.RoomIDs) > 0 {
		eg.Go(func() error {
			_, err = uc.roomUC.CreateTrackings(ctx, sc, room.CreateTrackingsInput{
				RoomIDs:     input.RoomIDs,
				EventID:     ne.ID.Hex(),
				InstanceID:  ne.ID.Hex(),
				StartTime:   ne.StartTime,
				EndTime:     ne.EndTime,
				Repeat:      input.Repeat,
				RepeatUntil: input.RepeatUntil,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.CreateTrackings: %v", err)
				return err
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
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

	return nil
}

// handleUpdateAllEventInstances handles updating all instances of an event (EventActionAll)
func (uc implUseCase) handleUpdateAllEventInstances(ctx context.Context, sc models.Scope, input event.UpdateInput, ri models.RecurringInstance) error {
	e, err := uc.repo.Detail(ctx, sc, ri.EventID.Hex())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "event.usecase.Update.Detail: %v", err)
			return event.ErrEventNotFound
		}
		uc.l.Errorf(ctx, "event.usecase.Update.Detail: %v", err)
		return err
	}

	updateOpts := repository.UpdateOptions{
		ID:            ri.EventID.Hex(),
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
		updateOpts.NotifyTime = uc.calculateNotifyTimeForEvent(input.StartTime, input.AllDay, input.Alert, *tz.Offset)
	}

	eg := new(errgroup.Group)

	var updatedEvent models.Event
	eg.Go(func() error {
		var err error
		updatedEvent, err = uc.repo.Update(ctx, sc, updateOpts)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.Update: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := uc.repo.DeleteRecurringInstance(ctx, sc, repository.DeleteRecurringInstanceOptions{
			EventID: ri.EventID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.DeleteRecurringInstance: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := uc.repo.DeleteRecurringTracking(ctx, sc, repository.DeleteRecurringTrackingOptions{
			EventID: ri.EventID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.DeleteRecurringTracking: %v", err)
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	if input.Repeat != models.EventRepeatNone {
		_, err = uc.generateRecurringInstances(ctx, sc, updatedEvent, updatedEvent.StartTime, &tz)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.generateRecurringInstances: %v", err)
			return err
		}
	}

	// Handle room tracking updates
	eg = new(errgroup.Group)

	eg.Go(func() error {
		err = uc.roomUC.DeleteTracking(ctx, sc, room.DeleteTrackingInput{
			RoomIDs: mongo.HexFromObjectIDsOrNil(ri.RoomIDs),
			EventID: ri.EventID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Update.DeleteTracking: %v", err)
			return err
		}
		return nil
	})

	if len(input.RoomIDs) > 0 {
		eg.Go(func() error {
			_, err = uc.roomUC.CreateTrackings(ctx, sc, room.CreateTrackingsInput{
				RoomIDs:     input.RoomIDs,
				EventID:     updatedEvent.ID.Hex(),
				StartTime:   updatedEvent.StartTime,
				EndTime:     updatedEvent.EndTime,
				Repeat:      input.Repeat,
				RepeatUntil: input.RepeatUntil,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.Update.CreateTrackings: %v", err)
				return err
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
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

	return nil
}

// handleDeleteSingleEventInstance handles deletion of a single event instance (EventActionOne)
func (uc implUseCase) handleDeleteSingleEventInstance(ctx context.Context, sc models.Scope, input event.DeleteInput, ri models.RecurringInstance) error {
	err := uc.repo.DeleteRecurringInstance(ctx, sc, repository.DeleteRecurringInstanceOptions{
		IDs:     []string{input.ID},
		EventID: input.EventID,
	})
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Delete.DeleteRecurringInstance: %v", err)
		return err
	}

	if len(ri.RoomIDs) > 0 {
		err = uc.roomUC.UpdateTrackingException(ctx, sc, room.UpdateTrackingExceptionInput{
			RoomIDs:   mongo.HexFromObjectIDsOrNil(ri.RoomIDs),
			EventID:   input.EventID,
			StartTime: ri.StartTime,
			EndTime:   ri.EndTime,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.UpdateTrackingException: %v", err)
			return err
		}
	}

	shop, err := uc.shopUC.DetailShop(ctx, sc, sc.ShopID)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Delete.DetailShop: %v", err)
		return err
	}

	err = uc.handleEventDeleteNotification(ctx, sc, event.RecurringInstanceToEventInstance(sc, ri), shop, models.EventActionOne)
	if err != nil {
		uc.l.Debugf(ctx, "event.usecase.Delete.handleEventDeleteNotification: %v", err)
		return err
	}

	return nil
}

// handleDeleteEventFromDate handles deletion of events from a specific date (EventActionFrom)
func (uc implUseCase) handleDeleteEventFromDate(ctx context.Context, sc models.Scope, input event.DeleteInput, ri models.RecurringInstance) error {
	timeToDelete := ri.StartTime

	if input.Type == models.EventActionFrom {
		prevStartTime, _ := uc.getPreviousOccurrence(ri.StartTime, ri.EndTime, ri.Repeat)
		timeToDelete = prevStartTime
	}

	eg := new(errgroup.Group)

	eg.Go(func() error {
		err := uc.repo.DeleteNextRecurringInstances(ctx, sc, repository.DeleteNextRecurringInstancesOptions{
			EventID: ri.EventID.Hex(),
			Date:    timeToDelete,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.DeleteNextRecurringInstances: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		_, err := uc.repo.UpdateRepeatUntil(ctx, sc, ri.EventID.Hex(), timeToDelete)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.UpdateRepeatUntil: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := uc.repo.UpdateRepeatUntilRecurringTrackings(ctx, sc, repository.UpdateRepeatUntilRecurringTrackingsOptions{
			EventID:     ri.EventID.Hex(),
			RepeatUntil: &timeToDelete,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.UpdateRepeatUntilRecurringTrackings: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := uc.roomUC.UpdateTrackingRepeatUntil(ctx, sc, room.UpdateTrackingRepeatUntilInput{
			RoomIDs:     mongo.HexFromObjectIDsOrNil(ri.RoomIDs),
			EventID:     ri.EventID.Hex(),
			RepeatUntil: timeToDelete,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.UpdateTrackingRepeatUntil: %v", err)
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	shop, err := uc.shopUC.DetailShop(ctx, sc, sc.ShopID)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Delete.DetailShop: %v", err)
		return err
	}

	err = uc.handleEventDeleteNotification(ctx, sc, event.RecurringInstanceToEventInstance(sc, ri), shop, input.Type)
	if err != nil {
		uc.l.Debugf(ctx, "event.usecase.Delete.handleEventDeleteNotification: %v", err)
		return err
	}

	return nil
}

// handleDeleteAllEventInstances handles deletion of all instances of an event (EventActionAll)
func (uc implUseCase) handleDeleteAllEventInstances(ctx context.Context, sc models.Scope, input event.DeleteInput, ri models.RecurringInstance) error {
	eg := new(errgroup.Group)

	eg.Go(func() error {
		err := uc.repo.DeleteRecurringInstance(ctx, sc, repository.DeleteRecurringInstanceOptions{
			EventID: ri.EventID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.DeleteRecurringInstance: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := uc.repo.Delete(ctx, sc, input.EventID)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.Delete: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := uc.repo.DeleteRecurringTracking(ctx, sc, repository.DeleteRecurringTrackingOptions{
			EventID: ri.EventID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.DeleteRecurringTracking: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := uc.roomUC.DeleteTracking(ctx, sc, room.DeleteTrackingInput{
			RoomIDs: mongo.HexFromObjectIDsOrNil(ri.RoomIDs),
			EventID: ri.EventID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Delete.DeleteTracking: %v", err)
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	shop, err := uc.shopUC.DetailShop(ctx, sc, sc.ShopID)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Delete.DetailShop: %v", err)
		return err
	}

	err = uc.handleEventDeleteNotification(ctx, sc, event.RecurringInstanceToEventInstance(sc, ri), shop, models.EventActionAll)
	if err != nil {
		uc.l.Debugf(ctx, "event.usecase.Delete.handleEventDeleteNotification: %v", err)
		return err
	}

	return nil
}
