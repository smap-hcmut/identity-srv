package mongo

import (
	"context"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildEventModel(ctx context.Context, sc models.Scope, opt repository.CreateOptions) (models.Event, error) {
	now := repo.clock()

	event := models.Event{
		ID:          primitive.NewObjectIDFromTimestamp(opt.StartTime),
		ShopID:      mongo.ObjectIDFromHexOrNil(sc.ShopID),
		Title:       opt.Title,
		AssignIDs:   opt.AssignIDs,
		AllDay:      opt.AllDay,
		Description: opt.Description,
		Notify:      opt.Notify,
		System:      opt.System,
		Public:      opt.Public,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if !opt.System {
		event.CreatedByID = sc.UserID
	}

	timezoneID, err := primitive.ObjectIDFromHex(opt.TimezoneID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildEventModel.ObjectIDFromHex: %v", err)
		return models.Event{}, err
	}
	event.TimezoneID = timezoneID

	if len(opt.DepartmentIDs) > 0 {
		departmentIDs, err := mongo.ObjectIDsFromHexs(opt.DepartmentIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildEventModel.ObjectIDFromHex: %v", err)
			return models.Event{}, err
		}
		event.DepartmentIDs = departmentIDs
	}

	if len(opt.RoomIDs) > 0 {
		roomIDs, err := mongo.ObjectIDsFromHexs(opt.RoomIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildEventModel.ObjectIDFromHex: %v", err)
			return models.Event{}, err
		}
		event.RoomIDs = roomIDs
	}

	if opt.Repeat != "" {
		event.Repeat = opt.Repeat
	} else {
		event.Repeat = models.EventRepeatNone
	}

	if opt.AllDay {
		event.StartTime = util.StartOfDay(opt.StartTime)
		event.EndTime = util.EndOfDay(opt.StartTime)
	} else {
		event.StartTime = opt.StartTime
		event.EndTime = opt.EndTime
	}

	if opt.CategoryID != "" {
		categoryID, err := primitive.ObjectIDFromHex(opt.CategoryID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildEventModel.ObjectIDFromHex: %v", err)
			return models.Event{}, err
		}
		event.CategoryID = &categoryID
	}

	if opt.RepeatUntil != nil {
		event.RepeatUntil = opt.RepeatUntil
	}

	if opt.NotifyTime != nil {
		event.NotifyTime = opt.NotifyTime
	}

	if opt.Alert != nil {
		alert := models.DateConfig{
			Num:     opt.Alert.Num,
			Unit:    opt.Alert.Unit,
			Hour:    opt.Alert.Hour,
			Instant: opt.Alert.Instant,
		}
		event.Alert = &alert
	}

	if opt.ObjectID != "" {
		objectID, err := primitive.ObjectIDFromHex(opt.ObjectID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildEventModel.ObjectIDFromHex: %v", err)
			return models.Event{}, err
		}
		event.ObjectID = &objectID
	}

	if len(opt.BranchIDs) > 0 {
		branchIDs, err := mongo.ObjectIDsFromHexs(opt.BranchIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildEventModel.ObjectIDFromHex: %v", err)
			return models.Event{}, err
		}
		event.BranchIDs = branchIDs
	}

	return event, nil
}

func (repo implRepository) buildUpdate(ctx context.Context, opt repository.UpdateOptions) (models.Event, bson.M, error) {
	unset := bson.M{}
	set := bson.M{}

	if opt.Title != "" {
		set["title"] = opt.Title
		opt.Model.Title = opt.Title
	}

	if len(opt.AssignIDs) > 0 {
		set["assign_ids"] = opt.AssignIDs
		opt.Model.AssignIDs = opt.AssignIDs
	} else {
		unset["assign_ids"] = 1
		opt.Model.AssignIDs = nil
	}

	if len(opt.DepartmentIDs) > 0 {
		departmentIDs, err := mongo.ObjectIDsFromHexs(opt.DepartmentIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdate.ObjectIDFromHex: %v", err)
			return models.Event{}, bson.M{}, err
		}
		set["department_ids"] = departmentIDs
		opt.Model.DepartmentIDs = departmentIDs
	} else {
		unset["department_ids"] = 1
		opt.Model.DepartmentIDs = nil
	}

	if opt.TimezoneID != "" {
		timezoneID, err := primitive.ObjectIDFromHex(opt.TimezoneID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdate.ObjectIDFromHex: %v", err)
			return models.Event{}, bson.M{}, err
		}
		set["timezone_id"] = timezoneID
		opt.Model.TimezoneID = timezoneID
	}

	set["start_time"] = opt.StartTime
	opt.Model.StartTime = opt.StartTime
	set["end_time"] = opt.EndTime
	opt.Model.EndTime = opt.EndTime
	set["all_day"] = opt.AllDay
	opt.Model.AllDay = opt.AllDay
	set["repeat"] = opt.Repeat
	opt.Model.Repeat = opt.Repeat

	if len(opt.RoomIDs) > 0 {
		roomIDs, err := mongo.ObjectIDsFromHexs(opt.RoomIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdate.ObjectIDFromHex: %v", err)
			return models.Event{}, bson.M{}, err
		}
		set["room_ids"] = roomIDs
		opt.Model.RoomIDs = roomIDs
	} else {
		unset["room_ids"] = 1
		opt.Model.RoomIDs = nil
	}

	if opt.Description != "" {
		set["description"] = opt.Description
		opt.Model.Description = opt.Description
	} else {
		unset["description"] = 1
		opt.Model.Description = ""
	}

	if opt.CategoryID != "" {
		categoryID, err := primitive.ObjectIDFromHex(opt.CategoryID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdate.ObjectIDFromHex: %v", err)
			return models.Event{}, bson.M{}, err
		}
		set["category_id"] = categoryID
		opt.Model.CategoryID = &categoryID
	} else {
		unset["category_id"] = 1
		opt.Model.CategoryID = nil
	}

	if opt.Alert != nil {
		set["alert"] = opt.Alert
		opt.Model.Alert = opt.Alert
	} else {
		unset["alert"] = 1
		opt.Model.Alert = nil
	}

	if opt.ObjectID != "" {
		objectID, err := primitive.ObjectIDFromHex(opt.ObjectID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdate.ObjectIDFromHex: %v", err)
			return models.Event{}, bson.M{}, err
		}
		set["object_id"] = objectID
		opt.Model.ObjectID = &objectID
	}

	if len(opt.BranchIDs) > 0 {
		branchIDs, err := mongo.ObjectIDsFromHexs(opt.BranchIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdate.ObjectIDFromHex: %v", err)
			return models.Event{}, bson.M{}, err
		}
		set["branch_ids"] = branchIDs
		opt.Model.BranchIDs = branchIDs
	} else {
		unset["branch_ids"] = 1
		opt.Model.BranchIDs = nil
	}

	if opt.NotifyTime != nil {
		set["notify_time"] = opt.NotifyTime
		opt.Model.NotifyTime = opt.NotifyTime
	} else {
		unset["notify_time"] = 1
		opt.Model.NotifyTime = nil
	}

	set["public"] = opt.Public
	opt.Model.Public = opt.Public

	if opt.RepeatUntil != nil {
		set["repeat_until"] = opt.RepeatUntil
		opt.Model.RepeatUntil = opt.RepeatUntil
	} else {
		unset["repeat_until"] = 1
		opt.Model.RepeatUntil = nil
	}

	set["updated_at"] = repo.clock()

	update := bson.M{}
	update["$set"] = set

	if len(unset) > 0 {
		update["$unset"] = unset
	}

	return opt.Model, update, nil
}
