package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

const (
	eventCollection = "events"
)

func (repo implRepository) getEventCollection(p int32, y int32) mongo.Collection {
	colName := fmt.Sprintf("%s_%d_%d", eventCollection, y, p)
	return repo.db.Collection(colName)
}

func (repo implRepository) Create(ctx context.Context, sc models.Scope, opt repository.CreateOptions) (models.Event, error) {
	event, err := repo.buildEventModel(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Create.buildEventModel: %v", err)
		return models.Event{}, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(event.ID)
	col := repo.getEventCollection(p, y)

	_, err = col.InsertOne(ctx, event)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Create.InsertOne: %v", err)
		return models.Event{}, err
	}

	return event, nil
}

func (repo implRepository) Detail(ctx context.Context, sc models.Scope, id string) (event models.Event, err error) {
	i, _ := primitive.ObjectIDFromHex(id)

	p, y := mongo.GetPeriodAndYearFromObjectID(i)
	col := repo.getEventCollection(p, y)

	filter, err := repo.buildDetailQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Detail.buildDetailQuery: %v", err)
		return models.Event{}, err
	}

	err = col.FindOne(ctx, filter).Decode(&event)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Detail.Decode: %v", err)
		return models.Event{}, err
	}

	return event, nil
}

func (repo implRepository) Update(ctx context.Context, sc models.Scope, opt repository.UpdateOptions) (models.Event, error) {
	id, _ := primitive.ObjectIDFromHex(opt.ID)

	p, y := mongo.GetPeriodAndYearFromObjectID(id)
	col := repo.getEventCollection(p, y)

	filter, err := repo.buildDetailQuery(ctx, sc, opt.ID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Update.buildDetailQuery: %v", err)
		return models.Event{}, err
	}

	m, update, err := repo.buildUpdate(ctx, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Update.buildUpdate: %v", err)
		return models.Event{}, err
	}

	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Update.UpdateOne: %v", err)
		return models.Event{}, err
	}

	return m, nil
}

func (repo implRepository) Delete(ctx context.Context, sc models.Scope, id string) error {
	mID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Delete.ObjectIDFromHex: %v", err)
		return mongo.ErrInvalidObjectID
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(mID)
	col := repo.getEventCollection(p, y)

	filter, err := repo.buildDetailQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Delete.buildDetailQuery: %v", err)
		return err
	}

	_, err = col.DeleteSoftOne(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.Delete.DeleteSoftOne: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) List(ctx context.Context, sc models.Scope, opt repository.ListOptions) ([]models.Event, error) {
	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.List.buildFilter: %v", err)
		return nil, err
	}

	periodRange := util.GetPeriodAndYearRange(opt.StartTime, opt.EndTime)

	g, ctx := errgroup.WithContext(ctx)
	var events []models.Event
	var mu sync.Mutex

	for _, period := range periodRange {
		g.Go(func() error {
			p, y := util.GetPeriodAndYear(period)
			col := repo.getEventCollection(p, y)
			cur, err := col.Find(ctx, filter, options.Find().
				SetSort(bson.D{
					{Key: "created_at", Value: -1},
					{Key: "_id", Value: -1},
				}),
			)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.List.Find: %v", err)
				return err
			}

			var evs []models.Event
			err = cur.All(ctx, &evs)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.List.All: %v", err)
				return err
			}

			mu.Lock()
			events = append(events, evs...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return events, nil
}

func (repo implRepository) GetOne(ctx context.Context, sc models.Scope, opt repository.GetOneOptions) (models.Event, error) {
	i, err := primitive.ObjectIDFromHex(opt.ID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetOne.ObjectIDFromHex: %v", err)
		return models.Event{}, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(i)
	col := repo.getEventCollection(p, y)

	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetOne.buildFilter: %v", err)
		return models.Event{}, err
	}

	var event models.Event
	err = col.FindOne(ctx, filter).Decode(&event)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetOne.Decode: %v", err)
		return models.Event{}, err
	}

	return event, nil
}

func (repo implRepository) UpdateRepeatUntil(ctx context.Context, sc models.Scope, id string, repeatUntil time.Time) (models.Event, error) {
	i, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRepeatUntil.ObjectIDFromHex: %v", err)
		return models.Event{}, err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(i)
	col := repo.getEventCollection(p, y)

	filter, err := repo.buildUpdateRepeatUntilQuery(ctx, sc, id, repeatUntil)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRepeatUntil.buildUpdateRepeatUntilQuery: %v", err)
		return models.Event{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"repeat_until": repeatUntil,
			"updated_at":   time.Now(),
		},
	}

	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRepeatUntil.UpdateOne: %v", err)
		return models.Event{}, err
	}

	return models.Event{}, nil
}

func (repo implRepository) ListByIDs(ctx context.Context, sc models.Scope, ids []string) ([]models.Event, error) {
	mapColName := make(map[string][]primitive.ObjectID)
	events := make([]models.Event, 0, len(ids))
	var mu sync.Mutex

	for _, id := range ids {
		i, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.ListByIDs.ObjectIDFromHex: %v", err)
			return nil, err
		}

		p, y := mongo.GetPeriodAndYearFromObjectID(i)
		colName := fmt.Sprintf("%s_%d_%d", eventCollection, y, p)

		if _, ok := mapColName[colName]; !ok {
			mapColName[colName] = []primitive.ObjectID{}
		}

		mapColName[colName] = append(mapColName[colName], i)
	}

	for colName, ids := range mapColName {
		col := repo.db.Collection(colName)
		filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.ListByIDs.BuildScopeQuery: %v", err)
			return nil, err
		}

		filter = mongo.BuildQueryWithSoftDelete(filter)

		filter["_id"] = bson.M{"$in": ids}

		cur, err := col.Find(ctx, filter)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.ListByIDs.Find: %v", err)
			return nil, err
		}

		var evs []models.Event
		err = cur.All(ctx, &evs)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.ListByIDs.All: %v", err)
			return nil, err
		}

		mu.Lock()
		events = append(events, evs...)
		mu.Unlock()
	}

	return events, nil
}

func (repo implRepository) UpdateAttendance(ctx context.Context, sc models.Scope, eventID string, status int) error {
	eID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateAttendance.ObjectIDFromHex: %v", err)
		return err
	}

	p, y := mongo.GetPeriodAndYearFromObjectID(eID)
	col := repo.getEventCollection(p, y)

	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateAttendance.BuildScopeQuery: %v", err)
		return err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"] = eID

	update := bson.M{}
	if status == 1 {
		update["$addToSet"] = bson.M{"accepted_ids": sc.UserID}
		update["$pull"] = bson.M{"declined_ids": sc.UserID}
		update["updated_at"] = time.Now()
	} else if status == -1 {
		update["$addToSet"] = bson.M{"declined_ids": sc.UserID}
		update["$pull"] = bson.M{"accepted_ids": sc.UserID}
		update["updated_at"] = time.Now()
	}

	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateAttendance.UpdateOne: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) SystemList(ctx context.Context, sc models.Scope, opt repository.SystemListOptions) ([]models.Event, error) {
	filter, err := repo.buildSystemListQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.SystemList.buildSystemListQuery: %v", err)
		return nil, err
	}

	periodRange := util.GetPeriodAndYearRange(opt.StartTime, opt.EndTime)

	g, ctx := errgroup.WithContext(ctx)
	var events []models.Event
	var mu sync.Mutex

	for _, period := range periodRange {
		g.Go(func() error {
			p, y := util.GetPeriodAndYear(period)
			col := repo.getEventCollection(p, y)
			cur, err := col.Find(ctx, filter, options.Find().
				SetSort(bson.D{
					{Key: "created_at", Value: -1},
					{Key: "_id", Value: -1},
				}),
			)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.List.Find: %v", err)
				return err
			}

			var evs []models.Event
			err = cur.All(ctx, &evs)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.List.All: %v", err)
				return err
			}

			mu.Lock()
			events = append(events, evs...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return events, nil
}
