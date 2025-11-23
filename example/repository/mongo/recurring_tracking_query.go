package mongo

import (
	"context"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildGenRTsInDateRangeQuery(ctx context.Context, sc models.Scope, fromTime, toTime time.Time) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildUpdateRecurringInstanceQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	// Common time constraints that apply to all patterns
	timeConstraints := bson.M{
		"start_time": bson.M{"$lte": toTime},
		"end_time":   bson.M{"$gte": fromTime},
	}

	// Generate all year-month combinations in the range
	var yearMonths []struct {
		Year  int
		Month int32
	}

	current := fromTime
	for current.Before(toTime) || current.Equal(toTime) {
		yearMonths = append(yearMonths, struct {
			Year  int
			Month int32
		}{
			Year:  current.Year(),
			Month: int32(current.Month()),
		})
		// Move to next month
		current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
	}

	// Build patterns for each repeat type
	repeatPatterns := []bson.M{
		{
			"repeat": "daily",
			"$and": bson.A{
				bson.M{"$or": util.BuildYearMonthConditions(yearMonths)},
				bson.M{"start_end_time": bson.M{"$elemMatch": timeConstraints}},
			},
		},
		{
			"repeat": "weekly",
			"$and": bson.A{
				bson.M{"$or": util.BuildYearMonthConditions(yearMonths)},
				bson.M{"start_end_time": bson.M{"$elemMatch": timeConstraints}},
			},
		},
		{
			"repeat": "monthly",
			"$and": bson.A{
				bson.M{"$or": util.BuildYearMonthConditions(yearMonths)},
				bson.M{"start_end_time": bson.M{"$elemMatch": timeConstraints}},
			},
		},
		{
			"repeat": "yearly",
			"$and": bson.A{
				bson.M{"$or": util.BuildYearMonthConditions(yearMonths)},
				bson.M{"start_end_time": bson.M{"$elemMatch": timeConstraints}},
			},
		},
	}

	// Combine conditions using $and and $or
	filter["$and"] = []bson.M{
		{"$or": repeatPatterns},
		{"$or": []bson.M{
			{"repeat_until": nil},
			{"repeat_until": bson.M{"$gte": fromTime}},
		}},
	}

	return filter, nil
}

func (repo implRepository) buildUpdateRepeatUntilRecurringTrackingsQuery(ctx context.Context, sc models.Scope, opt repository.UpdateRepeatUntilRecurringTrackingsOptions) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildUpdateRepeatUntilRecurringTrackingsQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	eID, err := primitive.ObjectIDFromHex(opt.EventID)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildUpdateRepeatUntilRecurringTrackingsQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	filter["event_id"] = eID

	return filter, nil
}

func (repo implRepository) buildDeleteRecurringTrackingQuery(ctx context.Context, sc models.Scope, opt repository.DeleteRecurringTrackingOptions) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "event.mongo.buildDeleteRecurringTrackingQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	// Handle deletion by IDs
	if len(opt.IDs) > 0 {
		eIDs := make([]primitive.ObjectID, 0, len(opt.IDs))
		for _, id := range opt.IDs {
			eID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "event.mongo.buildDeleteRecurringTrackingQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			eIDs = append(eIDs, eID)
		}
		filter["_id"] = bson.M{"$in": eIDs}
	}

	// Handle deletion by event ID
	if opt.EventID != "" {
		eID, err := primitive.ObjectIDFromHex(opt.EventID)
		if err != nil {
			repo.l.Errorf(ctx, "event.mongo.buildDeleteRecurringTrackingQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}

		filter["event_id"] = eID
	}

	// Handle deletion by start time (month and year)
	if opt.Month != nil {
		filter["month"] = bson.M{"$gte": opt.Month}
	}

	if opt.Year != nil {
		filter["year"] = bson.M{"$gte": opt.Year}
	}

	return filter, nil
}

func (repo implRepository) buildGenRTsNotInDateRangeQuery(ctx context.Context, sc models.Scope, fromTime, toTime time.Time) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	// Generate all year-month combinations in the range
	var yearMonths []struct {
		Year  int32
		Month int32
	}

	current := fromTime
	for current.Before(toTime) || current.Equal(toTime) {
		yearMonths = append(yearMonths, struct {
			Year  int32
			Month int32
		}{
			Year:  int32(current.Year()),
			Month: int32(current.Month()),
		})
		// Move to next month
		current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
	}

	filter["$and"] = []bson.M{
		// Must have a repeat type
		{"repeat": bson.M{"$exists": true}},

		// Not ended yet
		{
			"$or": []bson.M{
				{"repeat_until": nil},
				{"repeat_until": bson.M{"$gte": fromTime}},
			},
		},

		// No tracking has been created for this time range
		{
			"$or": []bson.M{
				// Daily events
				{
					"$and": []bson.M{
						{"repeat": "daily"},
						{"start_end_time": bson.M{"$not": bson.M{"$elemMatch": bson.M{
							"start_time": bson.M{"$gte": fromTime, "$lte": toTime},
						}}}},
					},
				},
				// Weekly events
				{
					"$and": []bson.M{
						{"repeat": "weekly"},
						{"start_end_time": bson.M{"$not": bson.M{"$elemMatch": bson.M{
							"start_time": bson.M{"$gte": fromTime, "$lte": toTime},
						}}}},
					},
				},
				// Monthly events - improved to prevent duplication
				{
					"$and": []bson.M{
						{"repeat": "monthly"},
						// At least one month-year not created tracking
						{
							"$or": buildMonthYearNotExistsConditions(yearMonths),
						},
					},
				},
				// Yearly events - improved to prevent duplication
				{
					"$and": []bson.M{
						{"repeat": "yearly"},
						{
							"$or": buildMonthYearNotExistsConditions(yearMonths),
						},
					},
				},
			},
		},
	}

	return filter, nil
}

// Helper function to create conditions for months-years that don't have tracking
func buildMonthYearNotExistsConditions(yearMonths []struct{ Year, Month int32 }) []bson.M {
	// Group all months by year
	monthsByYear := make(map[int32][]int32)
	for _, ym := range yearMonths {
		monthsByYear[ym.Year] = append(monthsByYear[ym.Year], ym.Month)
	}

	// For each year, create a condition
	conditions := make([]bson.M, 0)
	for year, months := range monthsByYear {
		// Create a condition that matches if either:
		// 1. The year doesn't match, OR
		// 2. The month doesn't match any of the months in this year
		conditions = append(conditions, bson.M{
			"$or": []bson.M{
				{
					"year": bson.M{"$ne": year},
				},
				{
					"month": bson.M{"$nin": months},
				},
			},
		})
	}

	return conditions
}
