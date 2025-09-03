package mongo

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildDetailQuery(ctx context.Context, sc models.Scope, id string) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.buildDetailQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"], err = primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.buildDetailQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	return filter, nil
}

func (repo implRepository) buildUpdateQuery(ctx context.Context, sc models.Scope, opt role.UpdateOptions) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.buildUpdateQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"] = opt.Model.ID

	return filter, nil
}

func (repo implRepository) buildDeleteQuery(ctx context.Context, sc models.Scope, ids []string) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.buildDeleteQuery.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	mIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		mID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			repo.l.Errorf(ctx, "role.mongo.buildDeleteQuery.ObjectIDFromHex: %v", mongo.ErrNoDocuments)
			return bson.M{}, mongo.ErrNoDocuments
		}
		mIDs = append(mIDs, mID)
	}

	filter["_id"] = bson.M{
		"$in": mIDs,
	}

	return filter, nil
}

func (repo implRepository) buildFilter(ctx context.Context, sc models.Scope, f role.Filter) (bson.M, error) {
	filter, err := mongo.BuildScopeQuery(ctx, repo.l, sc)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.buildFilter.BuildScopeQuery: %v", err)
		return bson.M{}, err
	}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if len(f.IDs) > 0 {
		mIDs := make([]primitive.ObjectID, 0, len(f.IDs))
		for _, id := range f.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "role.mongo.buildFilter.ObjectIDFromHex: %v", mongo.ErrInvalidObjectID)
				return bson.M{}, mongo.ErrInvalidObjectID
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
			repo.l.Errorf(ctx, "role.mongo.buildFilter.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	if len(f.Alias) > 0 {
		filter["alias"] = bson.M{
			"$in": f.Alias,
		}
	}

	if len(f.Code) > 0 {
		filter["code"] = bson.M{
			"$in": f.Code,
		}
	}

	return filter, nil
}
