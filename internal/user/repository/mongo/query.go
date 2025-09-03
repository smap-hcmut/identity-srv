package mongo

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildFilter(ctx context.Context, sc models.Scope, filter user.UserFilter) (bson.M, error) {
	f := bson.M{
		"deleted_at": bson.M{"$exists": false},
	}

	if len(filter.IDs) > 0 {
		objectIDs := make([]primitive.ObjectID, len(filter.IDs))
		for i, id := range filter.IDs {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "user.mongo.buildFilter.ObjectIDFromHex: %v", err)
				return nil, err
			}
			objectIDs[i] = objectID
		}
		f["_id"] = bson.M{"$in": objectIDs}
	}

	if len(filter.RoleIDs) > 0 {
		f["role_id"] = bson.M{"$in": filter.RoleIDs}
	}

	if len(filter.Providers) > 0 {
		f["provider"] = bson.M{"$in": filter.Providers}
	}

	if len(filter.Emails) > 0 {
		f["email"] = bson.M{"$in": filter.Emails}
	}

	if len(filter.NameCodes) > 0 {
		f["name_code"] = bson.M{"$in": filter.NameCodes}
	}

	return f, nil
}

func (repo implRepository) buildDetailQuery(ctx context.Context, sc models.Scope, id string) (bson.M, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.buildDetailQuery.ObjectIDFromHex: %v", err)
		return nil, err
	}

	return bson.M{
		"_id":        objectID,
		"deleted_at": bson.M{"$exists": false},
	}, nil
}

func (repo implRepository) buildUpdateVerifiedQuery(ctx context.Context, sc models.Scope, opt user.UpdateVerifiedOptions) (bson.M, error) {
	objectID, err := primitive.ObjectIDFromHex(opt.ID)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.buildUpdateVerifiedQuery.ObjectIDFromHex: %v", err)
		return nil, err
	}

	return bson.M{
		"_id":        objectID,
		"deleted_at": bson.M{"$exists": false},
	}, nil
}

func (repo implRepository) buildUpdateAvatarQuery(ctx context.Context, sc models.Scope, opt user.UpdateAvatarOptions) (bson.M, error) {
	objectID, err := primitive.ObjectIDFromHex(opt.ID)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.buildUpdateAvatarQuery.ObjectIDFromHex: %v", err)
		return nil, err
	}

	return bson.M{
		"_id":        objectID,
		"deleted_at": bson.M{"$exists": false},
	}, nil
}

func (repo implRepository) buildDeleteQuery(ctx context.Context, sc models.Scope, ids []string) (bson.M, error) {
	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			repo.l.Errorf(ctx, "user.mongo.buildDeleteQuery.ObjectIDFromHex: %v", err)
			return nil, err
		}
		objectIDs[i] = objectID
	}

	return bson.M{
		"_id":        bson.M{"$in": objectIDs},
		"deleted_at": bson.M{"$exists": false},
	}, nil
}