package mongo

import (
	"context"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildDetailQuery(ctx context.Context, sc models.Scope, id string) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildDetailQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"], err = primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildDetailQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	return filter, nil
}

func (repo implRepository) buildFilter(ctx context.Context, sc models.Scope, f repository.Filter) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildFilter.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if len(f.IDs) > 0 {
		mIDs := make([]primitive.ObjectID, 0, len(f.IDs))
		for _, id := range f.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.buildFilter.ObjectIDFromHex: %v", err)
				return bson.M{}, mongo.ErrNoDocuments
			}
			mIDs = append(mIDs, mID)
		}

		filter["_id"] = bson.M{
			"$in": mIDs,
		}
	}

	if f.ID != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(f.ID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildFilter.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	if f.NeedRepeat != nil {
		if *f.NeedRepeat {
			filter["repeat"] = bson.M{
				"$ne": models.EventRepeatNone,
			}
		} else {
			filter["repeat"] = models.EventRepeatNone
		}
	}

	if !f.StartTime.IsZero() && !f.EndTime.IsZero() {
		filter["$and"] = []bson.M{
			{"end_time": bson.M{"$gte": f.StartTime}},
			{"start_time": bson.M{"$lte": f.EndTime}},
		}
	}

	branchFilter := bson.M{}
	if len(f.BranchIDs) > 0 {
		bIDs := make([]primitive.ObjectID, 0, len(f.BranchIDs))
		for _, id := range f.BranchIDs {
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
		if len(f.DepartmentIDs) > 0 {
			dIDs := make([]primitive.ObjectID, 0, len(f.DepartmentIDs))
			for _, id := range f.DepartmentIDs {
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

	if len(f.ExcludeCategoryIDs) > 0 {
		catIDs := make([]primitive.ObjectID, 0, len(f.ExcludeCategoryIDs))
		for _, id := range f.ExcludeCategoryIDs {
			catID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.buildFilter.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			catIDs = append(catIDs, catID)
		}
		filter["category_id"] = bson.M{"$nin": catIDs}
	}

	return filter, nil
}

func (repo implRepository) buildSystemListQuery(ctx context.Context, sc models.Scope, opt repository.SystemListOptions) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildSystemListQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opt.NotifyTime != nil {
		filter["notify_time"] = opt.NotifyTime
	}

	if opt.NeedRepeat != nil {
		if *opt.NeedRepeat {
			filter["repeat"] = bson.M{
				"$ne": models.EventRepeatNone,
			}
		} else {
			filter["repeat"] = models.EventRepeatNone
		}
	}

	if !opt.StartTime.IsZero() && !opt.EndTime.IsZero() {
		filter["$and"] = []bson.M{
			{"end_time": bson.M{"$gte": opt.StartTime}},
			{"start_time": bson.M{"$lte": opt.EndTime}},
		}
	}

	return filter, nil
}
