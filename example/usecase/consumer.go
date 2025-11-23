package usecase

import (
	"context"

	"gitlab.com/gma-vietnam/tanca-connect/internal/element"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	rabb "gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
)

func (uc implUseCase) CreateSystemEvent(ctx context.Context, sc models.Scope, input event.CreateSystemEventInput) error {
	startTime, endTime, err := uc.parseEventTimes(ctx, input.StartTime, input.EndTime, input.NeedParseTimezone)
	if err != nil {
		return err
	}

	cat, err := uc.elementUC.Detail(ctx, sc, input.CategoryID)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Create.GetCategory: %v", err)
		return err
	}

	if cat.Key == BirthdayCategoryKey {
		err := uc.repo.Delete(ctx, sc, input.ObjectID)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Create.Delete: %v", err)
			return err
		}
	}

	opts := repository.CreateOptions{
		Title:         input.Title,
		AssignIDs:     input.AssignIDs,
		DepartmentIDs: input.DepartmentIDs,
		TimezoneID:    input.TimezoneID,
		StartTime:     startTime,
		EndTime:       endTime,
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
	}

	tz, err := uc.elementUC.Detail(ctx, sc, input.TimezoneID)
	if err != nil {
		if err == element.ErrElementNotFound {
			uc.l.Warnf(ctx, "event.usecase.Create.GetTimezone: %v", err)
			return event.ErrTimezoneNotFound
		}
		uc.l.Errorf(ctx, "event.usecase.Create.GetTimezone: %v", err)
		return err
	}

	if input.Notify {
		opts.NotifyTime = uc.calculateNotifyTimeForEvent(startTime, input.AllDay, input.Alert, *tz.Offset)
	}

	e, err := uc.repo.Create(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Create.Create: %v", err)
		return err
	}

	_, err = uc.generateRecurringInstances(ctx, sc, e, e.StartTime, &tz)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Create.generateRecurringInstances: %v", err)
		return err
	}

	if cat.Key == BusinessTripCategoryKey || cat.Key == OnLeaveCategoryKey {
		err = uc.producer.PublishUpdateRequestEventIDMsg(ctx, rabb.UpdateRequestEventIDMsg{
			RequestID: input.ObjectID,
			EventID:   e.ID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Create.PublishUpdateRequestEventIDMsg: %v", err)
			return err
		}
	}

	if cat.Key == InterviewCategoryKey {
		err = uc.producer.PublishUpdateTaskEventIDMsg(ctx, rabb.UpdateTaskEventIDMsg{
			TaskID:  input.ObjectID,
			EventID: e.ID.Hex(),
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Create.PublishUpdateTaskEventIDMsg: %v", err)
			return err
		}
	}

	return nil
}

func (uc implUseCase) UpdateSystemEvent(ctx context.Context, sc models.Scope, input event.UpdateSystemEventInput) error {
	e, err := uc.repo.Detail(ctx, sc, input.EventID)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Update.Detail: %v", err)
		return err
	}

	startTime, endTime, err := uc.parseEventTimes(ctx, input.StartTime, input.EndTime, input.NeedParseTimezone)
	if err != nil {
		return err
	}

	opts := repository.UpdateOptions{
		ID:            e.ID.Hex(),
		Model:         e,
		Title:         input.Title,
		AssignIDs:     input.AssignIDs,
		DepartmentIDs: input.DepartmentIDs,
		TimezoneID:    input.TimezoneID,
		StartTime:     startTime,
		EndTime:       endTime,
		AllDay:        input.AllDay,
		Repeat:        input.Repeat,
		RoomIDs:       input.RoomIDs,
		Description:   input.Description,
		CategoryID:    input.CategoryID,
		RepeatUntil:   input.RepeatUntil,
		Notify:        input.Notify,
		Alert:         input.Alert,
		ObjectID:      input.ObjectID,
	}

	tz, err := uc.elementUC.Detail(ctx, sc, input.TimezoneID)
	if err != nil {
		if err == element.ErrElementNotFound {
			uc.l.Warnf(ctx, "event.usecase.Update.GetTimezone: %v", err)
			return event.ErrTimezoneNotFound
		}
		uc.l.Errorf(ctx, "event.usecase.Update.GetTimezone: %v", err)
		return err
	}

	if input.Notify {
		opts.NotifyTime = uc.calculateNotifyTimeForEvent(startTime, input.AllDay, input.Alert, *tz.Offset)
	}

	_, err = uc.repo.Update(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.Update.Update: %v", err)
		return err
	}

	return nil
}
