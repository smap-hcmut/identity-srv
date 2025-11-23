package mongo

import (
	"context"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildRecurringTrackingModel(ctx context.Context, sc models.Scope, opt repository.CreateRecurringTrackingOptions) (models.RecurringTracking, error) {
	now := repo.clock()

	m := models.RecurringTracking{
		ID:        repo.db.NewObjectID(),
		ShopID:    mongo.ObjectIDFromHexOrNil(sc.ShopID),
		Month:     opt.Month,
		Year:      opt.Year,
		Repeat:    opt.Repeat,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if opt.RepeatUntil != nil {
		m.RepeatUntil = opt.RepeatUntil
	}

	for _, startEndTime := range opt.StartEndTime {
		m.StartEndTime = append(m.StartEndTime, models.StartEndTime{
			StartTime: startEndTime.StartTime,
			EndTime:   startEndTime.EndTime,
		})
	}

	eventID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildRecurringTrackingModel.ObjectIDFromHex: %v", err)
		return models.RecurringTracking{}, err
	}
	m.EventID = eventID

	return m, nil
}

func (repo implRepository) buildUpdateRepeatUntilRecurringTrackingsUpdate(opt repository.UpdateRepeatUntilRecurringTrackingsOptions) (bson.M, error) {
	update := bson.M{}

	if opt.RepeatUntil != nil {
		update["repeat_until"] = opt.RepeatUntil
	}

	return bson.M{"$set": update}, nil
}
