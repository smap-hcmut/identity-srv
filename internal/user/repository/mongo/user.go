package mongo

import (
	"context"
	"sync"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	userCollection = "users"
)

func (repo implRepository) getUserCollection() mongo.Collection {
	return repo.db.Collection(userCollection)
}

func (repo implRepository) Create(ctx context.Context, sc models.Scope, opt user.CreateOptions) (models.User, error) {
	col := repo.getUserCollection()

	user, err := repo.buildUserModel(sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.Create.buildUserModel: %v", err)
		return models.User{}, err
	}

	_, err = col.InsertOne(ctx, user)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.Create.InsertOne: %v", err)
		return models.User{}, err
	}

	return user, nil
}

func (repo implRepository) GetOne(ctx context.Context, sc models.Scope, opt user.GetOneOptions) (models.User, error) {
	col := repo.getUserCollection()

	filter, err := repo.buildFilter(ctx, sc, user.UserFilter{Emails: []string{opt.Email}})
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.GetOne.buildFilter: %v", err)
		return models.User{}, err
	}

	var user models.User
	err = col.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.GetOne.Decode: %v", err)
		return models.User{}, err
	}

	return user, nil
}

func (repo implRepository) Detail(ctx context.Context, sc models.Scope, id string) (user models.User, err error) {
	col := repo.getUserCollection()

	filter, err := repo.buildDetailQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.Detail.buildDetailQuery: %v", err)
		return models.User{}, err
	}

	err = col.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.Detail.Decode: %v", err)
		return models.User{}, err
	}

	return user, nil
}

func (repo implRepository) UpdateVerified(ctx context.Context, sc models.Scope, opt user.UpdateVerifiedOptions) (models.User, error) {
	col := repo.getUserCollection()

	filter, err := repo.buildUpdateVerifiedQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.UpdateVerified.buildUpdateVerifiedQuery: %v", err)
		return models.User{}, err
	}

	m, update := repo.buildUpdateVerified(opt)

	_, err = col.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.UpdateVerified.UpdateOne: %v", err)
		return models.User{}, err
	}

	return m, nil
}

func (repo implRepository) UpdateAvatar(ctx context.Context, sc models.Scope, opt user.UpdateAvatarOptions) (models.User, error) {
	col := repo.getUserCollection()

	filter, err := repo.buildUpdateAvatarQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.UpdateAvatar.buildUpdateAvatarQuery: %v", err)
		return models.User{}, err
	}

	m, update := repo.buildUpdateAvatar(opt)

	_, err = col.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.UpdateAvatar.UpdateOne: %v", err)
		return models.User{}, err
	}

	return m, nil
}

func (repo implRepository) Delete(ctx context.Context, sc models.Scope, ids []string) error {
	col := repo.getUserCollection()

	filter, err := repo.buildDeleteQuery(ctx, sc, ids)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.Delete.buildDeleteQuery: %v", err)
		return err
	}

	_, err = col.DeleteSoftMany(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.Delete.DeleteSoftMany: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) Get(ctx context.Context, sc models.Scope, opt user.GetOptions) ([]models.User, paginator.Paginator, error) {
	col := repo.getUserCollection()

	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.Get.buildFilter: %v", err)
		return nil, paginator.Paginator{}, err
	}

	var wg sync.WaitGroup
	var wgErrs error
	var total int64
	var users []models.User

	wg.Add(1)
	go func() {
		defer wg.Done()
		cur, err := col.Find(ctx, filter, options.Find().
			SetSkip(opt.PagQuery.Offset()).
			SetLimit(opt.PagQuery.Limit).
			SetSort(bson.D{
				{Key: "created_at", Value: -1},
				{Key: "_id", Value: -1},
			}),
		)
		if err != nil {
			repo.l.Errorf(ctx, "user.mongo.Get.Find: %v", err)
			wgErrs = err
			return
		}
		err = cur.All(ctx, &users)
		if err != nil {
			repo.l.Errorf(ctx, "user.mongo.Get.All: %v", err)
			wgErrs = err
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		total, err = col.CountDocuments(ctx, filter)
		if err != nil {
			repo.l.Errorf(ctx, "user.mongo.Get.CountDocuments: %v", err)
			wgErrs = err
		}
	}()
	wg.Wait()

	if wgErrs != nil {
		repo.l.Errorf(ctx, "user.mongo.Get.Wait: %v", wgErrs)
		return nil, paginator.Paginator{}, wgErrs
	}

	return users, paginator.Paginator{
		Total:       total,
		Count:       int64(len(users)),
		PerPage:     opt.PagQuery.Limit,
		CurrentPage: opt.PagQuery.Page,
	}, nil
}

func (repo implRepository) List(ctx context.Context, sc models.Scope, opt user.ListOptions) ([]models.User, error) {
	col := repo.getUserCollection()

	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.List.buildFilter: %v", err)
		return nil, err
	}

	cur, err := col.Find(ctx, filter, options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: -1},
			{Key: "_id", Value: -1},
		}),
	)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.List.Find: %v", err)
		return nil, err
	}

	var users []models.User
	err = cur.All(ctx, &users)
	if err != nil {
		repo.l.Errorf(ctx, "user.mongo.List.All: %v", err)
		return nil, err
	}

	return users, nil
}