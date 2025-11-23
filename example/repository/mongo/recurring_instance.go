package mongo

import (
	"context"
	"fmt"
	"sync"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

const (
	recurringInstanceCollection = "recurring_instances"
)

func (repo implRepository) getRecurringInstanceCollection(p int32, y int32) mongo.Collection {
	colName := fmt.Sprintf("%s_%d_%d", recurringInstanceCollection, y, p)
	return repo.db.Collection(colName)
}

func (repo implRepository) CreateRecurringInstance(ctx context.Context, sc models.Scope, opt repository.CreateRecurringInstanceOptions) (models.RecurringInstance, error) {
	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.CreateRecurringInstance.ObjectIDFromHex: %v", err)
		return models.RecurringInstance{}, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eventID)

	col := repo.getRecurringInstanceCollection(p, y)

	m, err := repo.buildRecurringInstanceModel(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.CreateRecurringInstance.buildRecurringInstanceModel: %v", err)
		return models.RecurringInstance{}, err
	}

	_, err = col.InsertOne(ctx, m)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Create.InsertOne: %v", err)
		return models.RecurringInstance{}, err
	}

	return m, nil
}

func (repo implRepository) CreateManyRecurringInstances(ctx context.Context, sc models.Scope, opt repository.CreateManyRecurringInstancesOptions) ([]models.RecurringInstance, error) {
	if len(opt.RecurringInstances) == 0 {
		return nil, nil
	}

	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.CreateManyRecurringInstances.ObjectIDFromHex: %v", err)
		return nil, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eventID)
	col := repo.getRecurringInstanceCollection(p, y)

	ms, err := repo.buildRecurringInstanceModels(ctx, sc, opt.RecurringInstances)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.CreateManyRecurringInstances.buildRecurringInstanceModels: %v", err)
		return nil, err
	}

	docs := make([]interface{}, len(ms))
	for i, m := range ms {
		docs[i] = m
	}

	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.CreateManyRecurringInstances.InsertMany: %v", err)
		return nil, err
	}

	return ms, nil
}

func (repo implRepository) DetailRecurringInstance(ctx context.Context, sc models.Scope, id string, eventID string) (models.RecurringInstance, error) {
	eID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DetailRecurringInstance.ObjectIDFromHex: %v", err)
		return models.RecurringInstance{}, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eID)
	col := repo.getRecurringInstanceCollection(p, y)

	filter, err := repo.buildDetailQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Detail.buildDetailQuery: %v", err)
		return models.RecurringInstance{}, err
	}

	var m models.RecurringInstance
	err = col.FindOne(ctx, filter).Decode(&m)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Detail.Decode: %v", err)
		return models.RecurringInstance{}, err
	}

	return m, nil
}

func (repo implRepository) UpdateRecurringInstance(ctx context.Context, sc models.Scope, opt repository.UpdateRecurringInstanceOptions) (models.RecurringInstance, error) {
	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRecurringInstance.ObjectIDFromHex: %v", err)
		return models.RecurringInstance{}, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eventID)
	col := repo.getRecurringInstanceCollection(p, y)
	filter, err := repo.buildUpdateRecurringInstanceQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRecurringInstance.buildUpdateQuery: %v", err)
		return models.RecurringInstance{}, err
	}

	m, update, err := repo.buildUpdateRecurringInstanceUpdate(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRecurringInstance.buildUpdate: %v", err)
		return models.RecurringInstance{}, err
	}

	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRecurringInstance.UpdateOne: %v", err)
		return models.RecurringInstance{}, err
	}

	return m, nil
}

func (repo implRepository) DeleteRecurringInstance(ctx context.Context, sc models.Scope, opt repository.DeleteRecurringInstanceOptions) error {
	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringInstance.ObjectIDFromHex: %v", err)
		return err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eventID)
	col := repo.getRecurringInstanceCollection(p, y)

	filter, err := repo.buildDeleteRecurringInstanceQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringInstance.buildDeleteQuery: %v", err)
		return err
	}

	_, err = col.DeleteSoftMany(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringInstance.DeleteSoftMany: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) ListRecurringInstances(ctx context.Context, sc models.Scope, opt repository.ListRecurringInstancesOptions) ([]models.RecurringInstance, error) {
	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.ListRecurringInstances.ObjectIDFromHex: %v", err)
		return nil, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eventID)
	col := repo.getRecurringInstanceCollection(p, y)

	filter, err := repo.buildListRecurringInstancesQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.ListRecurringInstance.buildListRecurringInstancesQuery: %v", err)
		return nil, err
	}

	cur, err := col.Find(ctx, filter, options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: -1},
			{Key: "_id", Value: -1},
		}),
	)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.ListRecurringInstance.Find: %v", err)
		return nil, err
	}

	var ms []models.RecurringInstance
	err = cur.All(ctx, &ms)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.ListRecurringInstance.All: %v", err)
		return nil, err
	}

	return ms, nil
}

func (repo implRepository) GetOneRecurringInstance(ctx context.Context, sc models.Scope, opt repository.GetOneRecurringInstanceOptions) (models.RecurringInstance, error) {
	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetOneRecurringInstance.ObjectIDFromHex: %v", err)
		return models.RecurringInstance{}, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eventID)
	col := repo.getRecurringInstanceCollection(p, y)

	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetOneRecurringInstance.buildFilter: %v", err)
		return models.RecurringInstance{}, err
	}

	var m models.RecurringInstance
	err = col.FindOne(ctx, filter).Decode(&m)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetOneRecurringInstance.Decode: %v", err)
		return models.RecurringInstance{}, err
	}

	return m, nil
}

func (repo implRepository) DeleteRecurringInstancesByEventID(ctx context.Context, sc models.Scope, eventID string) error {
	eID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringInstancesByEventID.ObjectIDFromHex: %v", err)
		return err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eID)
	col := repo.getRecurringInstanceCollection(p, y)

	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringInstancesByEventID.BuildScopeQuery: %v", err)
		return err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["event_id"] = eID

	_, err = col.DeleteSoftMany(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringInstancesByEventID.DeleteSoftMany: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) UpdateAttendanceRecurringInstance(ctx context.Context, sc models.Scope, opt repository.UpdateAttendanceRecurringInstanceOptions) error {
	eID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateAttendanceRecurringInstance.ObjectIDFromHex: %v", err)
		return err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eID)
	col := repo.getRecurringInstanceCollection(p, y)

	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateAttendanceRecurringInstance.BuildScopeQuery: %v", err)
		return err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	i, err := primitive.ObjectIDFromHex(opt.ID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateAttendanceRecurringInstance.ObjectIDFromHex: %v", err)
		return err
	}
	filter["_id"] = i
	filter["event_id"] = eID

	update := bson.M{}
	if opt.Status == 1 {
		update = bson.M{
			"$push": bson.M{"accepted_ids": sc.UserID},
			"$set":  bson.M{"updated_at": repo.clock()},
		}
	} else if opt.Status == -1 {
		update = bson.M{
			"$push": bson.M{"declined_ids": sc.UserID},
			"$set":  bson.M{"updated_at": repo.clock()},
		}
	}

	_, err = col.UpdateMany(ctx, filter, update)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateAttendanceRecurringInstance.UpdateMany: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) ListRecurringInstancesByEventIDs(ctx context.Context, sc models.Scope, opt repository.ListEventInstancesByEventIDsOptions) ([]models.RecurringInstance, error) {
	mapColName := make(map[string][]primitive.ObjectID)
	eg := new(errgroup.Group)
	var mu sync.Mutex
	ris := make([]models.RecurringInstance, 0)

	for _, id := range opt.EventIDs {
		i, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.ListByIDs.ObjectIDFromHex: %v", err)
			return nil, err
		}

		p, y := mongo.GetPeriodAndYearFromObjectID(i)
		colName := fmt.Sprintf("%s_%d_%d", recurringInstanceCollection, y, p)

		if _, ok := mapColName[colName]; !ok {
			mapColName[colName] = []primitive.ObjectID{}
		}

		mapColName[colName] = append(mapColName[colName], i)
	}

	for colName, ids := range mapColName {
		colName, ids := colName, ids // Create new variables for goroutine
		eg.Go(func() error {
			col := repo.db.Collection(colName)
			filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.ListRecurringInstancesByEventIDs.BuildScopeQuery: %v", err)
				return err
			}

			filter = mongo.BuildQueryWithSoftDelete(filter)

			filter["event_id"] = bson.M{"$in": ids}

			if !opt.StartTime.IsZero() && !opt.EndTime.IsZero() {
				filter["$and"] = []bson.M{
					{"end_time": bson.M{"$gt": opt.StartTime}},
					{"start_time": bson.M{"$lt": opt.EndTime}},
				}
			}

			if opt.NotifyTime != nil {
				filter["notify_time"] = opt.NotifyTime
			}

			cur, err := col.Find(ctx, filter)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.ListRecurringInstancesByEventIDs.Find: %v", err)
				return err
			}

			var ms []models.RecurringInstance
			err = cur.All(ctx, &ms)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.ListRecurringInstancesByEventIDs.All: %v", err)
				return err
			}

			mu.Lock()
			ris = append(ris, ms...)
			mu.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return ris, nil
}

func (repo implRepository) DeleteNextRecurringInstances(ctx context.Context, sc models.Scope, opt repository.DeleteNextRecurringInstancesOptions) error {
	id, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteNextRecurringInstances.ObjectIDFromHex: %v", err)
		return err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(id)
	col := repo.getRecurringInstanceCollection(p, y)

	filter, err := repo.buildDeleteNextRecurringInstancesQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteNextRecurringInstances.buildDeleteNextRecurringInstancesQuery: %v", err)
		return err
	}

	_, err = col.DeleteSoftMany(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteNextRecurringInstances.DeleteMany: %v", err)
		return err
	}

	return nil
}
