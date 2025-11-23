package mongo

import (
	"context"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildRecurringInstanceModel(ctx context.Context, sc models.Scope, opt repository.CreateRecurringInstanceOptions) (models.RecurringInstance, error) {
	now := repo.clock()

	m := models.RecurringInstance{
		ID:          repo.db.NewObjectID(),
		ShopID:      mongo.ObjectIDFromHexOrNil(sc.ShopID),
		Title:       opt.Title,
		StartTime:   opt.StartTime,
		EndTime:     opt.EndTime,
		AllDay:      opt.AllDay,
		Repeat:      opt.Repeat,
		Description: opt.Description,
		Notify:      opt.Notify,
		System:      opt.System,
		Public:      opt.Public,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if !opt.System {
		m.CreatedByID = sc.UserID
	}

	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildRecurringInstanceModel.ObjectIDFromHex: %v", err)
		return models.RecurringInstance{}, err
	}
	m.EventID = eventID

	if len(opt.DepartmentIDs) > 0 {
		departmentIDs, err := mongo.ObjectIDsFromHexs(opt.DepartmentIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildRecurringInstanceModel.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, err
		}
		m.DepartmentIDs = departmentIDs
	}

	if opt.CategoryID != "" {
		categoryID, err := primitive.ObjectIDFromHex(opt.CategoryID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildRecurringInstanceModel.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, err
		}
		m.CategoryID = &categoryID
	}

	if len(opt.AssignIDs) > 0 {
		m.AssignIDs = opt.AssignIDs
	}

	if len(opt.RoomIDs) > 0 {
		roomIDs, err := mongo.ObjectIDsFromHexs(opt.RoomIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildRecurringInstanceModel.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, err
		}
		m.RoomIDs = roomIDs
	}

	if opt.TimezoneID != "" {
		timezoneID, err := primitive.ObjectIDFromHex(opt.TimezoneID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildRecurringInstanceModel.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, err
		}
		m.TimezoneID = timezoneID
	}

	if opt.RepeatUntil != nil {
		m.RepeatUntil = opt.RepeatUntil
	}

	if opt.NotifyTime != nil {
		m.NotifyTime = opt.NotifyTime
	}

	if opt.Alert != nil {
		alert := models.DateConfig{
			Num:     opt.Alert.Num,
			Unit:    opt.Alert.Unit,
			Hour:    opt.Alert.Hour,
			Instant: opt.Alert.Instant,
		}
		m.Alert = &alert
	}

	if len(opt.BranchIDs) > 0 {
		branchIDs, err := mongo.ObjectIDsFromHexs(opt.BranchIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildRecurringInstanceModel.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, err
		}
		m.BranchIDs = branchIDs
	}

	return m, nil
}

func (repo implRepository) buildRecurringInstanceModels(ctx context.Context, sc models.Scope, opts []repository.CreateRecurringInstanceOptions) ([]models.RecurringInstance, error) {
	ms := make([]models.RecurringInstance, len(opts))
	for i, opt := range opts {
		m, err := repo.buildRecurringInstanceModel(ctx, sc, opt)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildRecurringInstanceModels.buildRecurringInstanceModel: %v", err)
			return nil, err
		}

		ms[i] = m
	}

	return ms, nil
}

func (repo implRepository) buildUpdateRecurringInstanceUpdate(ctx context.Context, sc models.Scope, opt repository.UpdateRecurringInstanceOptions) (models.RecurringInstance, bson.M, error) {
	unset := bson.M{}
	set := bson.M{}

	if opt.Title != "" {
		set["title"] = opt.Title
		opt.Model.Title = opt.Title
	}

	if opt.Description != "" {
		set["description"] = opt.Description
		opt.Model.Description = opt.Description
	} else {
		unset["description"] = 1
		opt.Model.Description = ""
	}

	if len(opt.AssignIDs) > 0 {
		set["assign_ids"] = opt.AssignIDs
		opt.Model.AssignIDs = opt.AssignIDs
	} else {
		unset["assign_ids"] = 1
		opt.Model.AssignIDs = []string{}
	}

	if len(opt.DepartmentIDs) > 0 {
		departmentIDs, err := mongo.ObjectIDsFromHexs(opt.DepartmentIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceUpdate.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, bson.M{}, err
		}
		set["department_ids"] = departmentIDs
		opt.Model.DepartmentIDs = departmentIDs
	} else {
		unset["department_ids"] = 1
		opt.Model.DepartmentIDs = []primitive.ObjectID{}
	}

	if opt.TimezoneID != "" {
		timezoneID, err := primitive.ObjectIDFromHex(opt.TimezoneID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceUpdate.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, bson.M{}, err
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
	set["public"] = opt.Public
	opt.Model.Public = opt.Public

	if len(opt.RoomIDs) > 0 {
		roomIDs, err := mongo.ObjectIDsFromHexs(opt.RoomIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceUpdate.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, bson.M{}, err
		}
		set["room_ids"] = roomIDs
		opt.Model.RoomIDs = roomIDs
	} else {
		unset["room_ids"] = 1
		opt.Model.RoomIDs = []primitive.ObjectID{}
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
			repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceUpdate.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, bson.M{}, err
		}
		set["category_id"] = categoryID
		opt.Model.CategoryID = &categoryID
	} else {
		unset["category_id"] = 1
		opt.Model.CategoryID = nil
	}

	if opt.NotifyTime != nil {
		set["notify_time"] = opt.NotifyTime
		opt.Model.NotifyTime = opt.NotifyTime
	}

	if opt.Alert != nil {
		set["alert"] = opt.Alert
		opt.Model.Alert = opt.Alert
	} else {
		unset["alert"] = 1
		opt.Model.Alert = nil
	}

	if len(opt.BranchIDs) > 0 {
		branchIDs, err := mongo.ObjectIDsFromHexs(opt.BranchIDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceUpdate.ObjectIDFromHex: %v", err)
			return models.RecurringInstance{}, bson.M{}, err
		}
		set["branch_ids"] = branchIDs
		opt.Model.BranchIDs = branchIDs
	} else {
		unset["branch_ids"] = 1
		opt.Model.BranchIDs = []primitive.ObjectID{}
	}

	set["updated_at"] = repo.clock()
	opt.Model.UpdatedAt = repo.clock()

	update := bson.M{}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	update["$set"] = set

	return opt.Model, update, nil
}
