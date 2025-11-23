package mongo

import (
	"context"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildUpdateRecurringInstanceQuery(ctx context.Context, sc models.Scope, opt repository.UpdateRecurringInstanceOptions) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"], err = primitive.ObjectIDFromHex(opt.ID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	return filter, nil
}

func (repo implRepository) buildDeleteRecurringInstanceQuery(ctx context.Context, sc models.Scope, opt repository.DeleteRecurringInstanceOptions) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildDeleteRecurringInstanceQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if len(opt.IDs) > 0 {
		ids, err := mongo.ObjectIDsFromHexs(opt.IDs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildDeleteRecurringInstanceQuery.ObjectIDsFromHexs: %v", err)
			return bson.M{}, err
		}

		filter["_id"] = bson.M{
			"$in": ids,
		}
	}

	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildDeleteRecurringInstanceQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	filter["event_id"] = eventID

	return filter, nil
}

func (repo implRepository) buildDeleteNextRecurringInstancesQuery(ctx context.Context, sc models.Scope, opt repository.DeleteNextRecurringInstancesOptions) (bson.M, error) {
	if opt.EventID == "" || opt.Date.IsZero() {
		repo.l.Errorf(ctx, "event.mongo.buildDeleteNextRecurringInstancesQuery.RequiredField: %v", event.ErrRequiredField)
		return bson.M{}, event.ErrRequiredField
	}

	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildDeleteNextRecurringInstancesQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildDeleteNextRecurringInstancesQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	filter["event_id"] = eventID

	filter["start_time"] = bson.M{
		"$gt": opt.Date,
	}

	return filter, nil
}

func (repo implRepository) buildUpdateRepeatUntilQuery(ctx context.Context, sc models.Scope, id string, repeatUntil time.Time) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildUpdateRepeatUntilQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"], err = primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildUpdateRepeatUntilQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	return filter, nil
}

func (repo implRepository) buildListRecurringInstancesQuery(ctx context.Context, sc models.Scope, opt repository.ListRecurringInstancesOptions) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildListRecurringInstancesQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildListRecurringInstancesQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	filter["event_id"] = eventID

	if !opt.StartTime.IsZero() && !opt.EndTime.IsZero() {
		filter["$and"] = []bson.M{
			{"start_time": bson.M{"$lte": opt.EndTime}}, // Event starts before or at range end
			{"end_time": bson.M{"$gte": opt.StartTime}}, // Event ends after or at range start
		}
	} else if !opt.StartTime.IsZero() {
		filter["end_time"] = bson.M{"$gte": opt.StartTime}
	} else if !opt.EndTime.IsZero() {
		filter["start_time"] = bson.M{"$lte": opt.EndTime}
	}

	branchFilter := bson.M{}
	if len(opt.BranchIDs) > 0 {
		bIDs := make([]primitive.ObjectID, 0, len(opt.BranchIDs))
		for _, id := range opt.BranchIDs {
			bID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.buildFilter.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			bIDs = append(bIDs, bID)
		}

		orConditions := bson.A{
			// First condition: branch matches AND no department/assign IDs
			bson.M{
				"$and": bson.A{
					bson.M{
						"branch_ids": bson.M{
							"$elemMatch": bson.M{
								"$in": bIDs,
							},
						},
					},
					bson.M{"department_ids": bson.M{"$exists": false}},
					bson.M{"assign_ids": bson.M{"$exists": false}},
				},
			},
		}

		// Add department_ids condition if exists
		if len(opt.DepartmentIDs) > 0 {
			dIDs := make([]primitive.ObjectID, 0, len(opt.DepartmentIDs))
			for _, id := range opt.DepartmentIDs {
				dID, err := primitive.ObjectIDFromHex(id)
				if err != nil {
					repo.l.Errorf(ctx, "event.mongo.buildFilter.ObjectIDFromHex: %v", err)
					return bson.M{}, err
				}
				dIDs = append(dIDs, dID)
			}

			orConditions = append(orConditions, bson.M{
				"department_ids": bson.M{
					"$elemMatch": bson.M{
						"$in": dIDs,
					},
				},
			})
		}

		// Add assign_ids condition separately
		orConditions = append(orConditions, bson.M{
			"assign_ids": bson.M{
				"$elemMatch": bson.M{
					"$eq": sc.UserID,
				},
			},
		})

		// Add created_by condition
		orConditions = append(orConditions, bson.M{
			"created_by_id": sc.UserID,
		})

		orConditions = append(orConditions, bson.M{
			"system": true,
		})

		orConditions = append(orConditions, bson.M{
			"public": true,
		})

		branchFilter = bson.M{
			"$or": orConditions,
		}
	}

	if len(branchFilter) > 0 {
		if filter["$and"] == nil {
			filter["$and"] = []bson.M{}
		}
		filter["$and"] = append(filter["$and"].([]bson.M), branchFilter)
	} else {
		if filter["$or"] == nil {
			filter["$or"] = []bson.M{}
		}
		filter["$or"] = append(filter["$or"].([]bson.M), bson.M{
			"created_by_id": sc.UserID,
		})
		filter["$or"] = append(filter["$or"].([]bson.M), bson.M{
			"system": true,
		})
		filter["$or"] = append(filter["$or"].([]bson.M), bson.M{
			"public": true,
		})
	}

	return filter, nil
}
