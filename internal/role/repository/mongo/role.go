package mongo

import (
	"context"
	"sync"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	roleCollection = "roles"
)

func (repo implRepository) getRoleCollection() mongo.Collection {
	return repo.db.Collection(roleCollection)
}

func (repo implRepository) Create(ctx context.Context, sc models.Scope, opt role.CreateOptions) (models.Role, error) {
	col := repo.getRoleCollection()

	role := repo.buildRoleModel(sc, opt)

	_, err := col.InsertOne(ctx, role)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Create.InsertOne: %v", err)
		return models.Role{}, err
	}

	return role, nil
}

func (repo implRepository) GetOne(ctx context.Context, sc models.Scope, opt role.GetOneOptions) (models.Role, error) {
	col := repo.getRoleCollection()

	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.GetOne.buildFilter: %v", err)
		return models.Role{}, err
	}

	var role models.Role
	err = col.FindOne(ctx, filter).Decode(&role)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.GetOne.Decode: %v", err)
		return models.Role{}, err
	}

	return role, nil
}

func (repo implRepository) Detail(ctx context.Context, sc models.Scope, id string) (role models.Role, err error) {
	col := repo.getRoleCollection()

	filter, err := repo.buildDetailQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Detail.buildDetailQuery: %v", err)
		return models.Role{}, err
	}

	err = col.FindOne(ctx, filter).Decode(&role)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Detail.Decode: %v", err)
		return models.Role{}, err
	}

	return role, nil
}

func (repo implRepository) Update(ctx context.Context, sc models.Scope, opt role.UpdateOptions) (models.Role, error) {
	col := repo.getRoleCollection()

	filter, err := repo.buildUpdateQuery(ctx, sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Update.buildUpdateQuery: %v", err)
		return models.Role{}, err
	}

	m, update := repo.buildUpdate(opt)

	_, err = col.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Update.UpdateOne: %v", err)
		return models.Role{}, err
	}

	return m, nil
}

func (repo implRepository) Delete(ctx context.Context, sc models.Scope, ids []string) error {
	col := repo.getRoleCollection()

	filter, err := repo.buildDeleteQuery(ctx, sc, ids)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Delete.buildDeleteQuery: %v", err)
		return err
	}

	_, err = col.DeleteSoftMany(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Delete.DeleteSoftMany: %v", err)
		return err
	}

	return nil
}

func (repo implRepository) Get(ctx context.Context, sc models.Scope, opt role.GetOptions) ([]models.Role, paginator.Paginator, error) {
	col := repo.getRoleCollection()

	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.Get.buildFilter: %v", err)
		return nil, paginator.Paginator{}, err
	}

	var wg sync.WaitGroup
	var wgErrs error
	var total int64
	var roles []models.Role

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
			repo.l.Errorf(ctx, "role.mongo.Get.Find: %v", err)
			wgErrs = err
			return
		}
		err = cur.All(ctx, &roles)
		if err != nil {
			repo.l.Errorf(ctx, "role.mongo.Get.All: %v", err)
			wgErrs = err
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		total, err = col.CountDocuments(ctx, filter)
		if err != nil {
			repo.l.Errorf(ctx, "role.mongo.Get.CountDocuments: %v", err)
			wgErrs = err
		}
	}()
	wg.Wait()

	if wgErrs != nil {
		repo.l.Errorf(ctx, "role.mongo.Get.Wait: %v", wgErrs)
		return nil, paginator.Paginator{}, wgErrs
	}

	return roles, paginator.Paginator{
		Total:       total,
		Count:       int64(len(roles)),
		PerPage:     opt.PagQuery.Limit,
		CurrentPage: opt.PagQuery.Page,
	}, nil
}

func (repo implRepository) List(ctx context.Context, sc models.Scope, opt role.ListOptions) ([]models.Role, error) {
	col := repo.getRoleCollection()

	filter, err := repo.buildFilter(ctx, sc, opt.Filter)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.List.buildFilter: %v", err)
		return nil, err
	}

	cur, err := col.Find(ctx, filter, options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: -1},
			{Key: "_id", Value: -1},
		}),
	)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.List.Find: %v", err)
		return nil, err
	}

	var roles []models.Role
	err = cur.All(ctx, &roles)
	if err != nil {
		repo.l.Errorf(ctx, "role.mongo.List.All: %v", err)
		return nil, err
	}

	return roles, nil
}
