package mongo

import (
	"context"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
)

const (
	recurringTrackingCollection = "recurring_trackings"
)

func (repo implRepository) getRecurringTrackingCollection() mongo.Collection {
	return repo.db.Collection(recurringTrackingCollection)
}

func (repo implRepository) CreateRecurringTracking(ctx context.Context, sc models.Scope, opt repository.CreateRecurringTrackingOptions) (models.RecurringTracking, error) {
	col := repo.getRecurringTrackingCollection()

	m, err := repo.buildRecurringTrackingModel(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.CreateRecurringTracking.buildRecurringTrackingModel: %v", err)
		return models.RecurringTracking{}, err
	}

	_, err = col.InsertOne(ctx, m)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.CreateRecurringTracking.InsertOne: %v", err)
		return models.RecurringTracking{}, err
	}

	return m, nil
}

func (repo implRepository) GetGenRTsInDateRange(ctx context.Context, sc models.Scope, fromTime, toTime time.Time) ([]models.RecurringTracking, error) {
	col := repo.getRecurringTrackingCollection()

	filter, err := repo.buildGenRTsInDateRangeQuery(ctx, sc, fromTime, toTime)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetGenRTsInDateRange.buildGenRTsInDateRangeQuery: %v", err)
		return nil, err
	}

	cur, err := col.Find(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetGenRTsInDateRange.Find: %v", err)
		return nil, err
	}

	var results []models.RecurringTracking
	if err := cur.All(ctx, &results); err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetGenRTsInDateRange.All: %v", err)
		return nil, err
	}

	return results, nil
}

func (repo implRepository) GetGenRTsNotInDateRange(ctx context.Context, sc models.Scope, fromTime, toTime time.Time) ([]models.RecurringTracking, error) {
	col := repo.getRecurringTrackingCollection()

	filter, err := repo.buildGenRTsNotInDateRangeQuery(ctx, sc, fromTime, toTime)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetGenRTsNotInDateRange.buildGenRTsNotInDateRangeQuery: %v", err)
		return nil, err
	}

	cur, err := col.Find(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetGenRTsNotInDateRange.Find: %v", err)
		return nil, err
	}

	var results []models.RecurringTracking
	if err := cur.All(ctx, &results); err != nil {
		repo.l.Errorf(ctx, "event.mongo.GetGenRTsNotInDateRange.All: %v", err)
		return nil, err
	}

	return results, nil
}

func (repo implRepository) UpdateRepeatUntilRecurringTrackings(ctx context.Context, sc models.Scope, opt repository.UpdateRepeatUntilRecurringTrackingsOptions) error {
	col := repo.getRecurringTrackingCollection()

	filter, err := repo.buildUpdateRepeatUntilRecurringTrackingsQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRepeatUntilRecurringTrackings.buildUpdateRepeatUntilRecurringTrackingsQuery: %v", err)
		return err
	}

	update, err := repo.buildUpdateRepeatUntilRecurringTrackingsUpdate(opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRepeatUntilRecurringTrackings.buildUpdateRepeatUntilRecurringTrackingsUpdate: %v", err)
		return err
	}

	_, err = col.UpdateMany(ctx, filter, update)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.UpdateRepeatUntilRecurringTrackings.UpdateMany: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) DeleteRecurringTracking(ctx context.Context, sc models.Scope, opt repository.DeleteRecurringTrackingOptions) error {
	col := repo.getRecurringTrackingCollection()

	filter, err := repo.buildDeleteRecurringTrackingQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringTracking.buildDeleteRecurringTrackingQuery: %v", err)
		return err
	}

	_, err = col.DeleteSoftMany(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.DeleteRecurringTracking.DeleteMany: %v", err)
		return err
	}

	return nil
}
