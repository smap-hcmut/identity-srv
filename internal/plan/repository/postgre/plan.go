package postgres

import (
	"context"
	"database/sql"

	"smap-api/internal/model"
	"smap-api/internal/plan/repository"
	"smap-api/internal/sqlboiler"
	"smap-api/pkg/paginator"
	postgresPkg "smap-api/pkg/postgre"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

func (r *implRepository) Detail(ctx context.Context, sc model.Scope, id string) (model.Plan, error) {
	mods, err := r.buildDetailQuery(ctx, id)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Detail.buildDetailQuery: %v", err)
		return model.Plan{}, err
	}

	plan, err := sqlboiler.Plans(mods...).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Plan{}, repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Detail.One: %v", err)
		return model.Plan{}, err
	}

	return *model.NewPlanFromDB(plan), nil
}

func (r *implRepository) List(ctx context.Context, sc model.Scope, opts repository.ListOptions) ([]model.Plan, error) {
	mods, err := r.buildListQuery(ctx, opts)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.List.buildListQuery: %v", err)
		return nil, err
	}

	plans, err := sqlboiler.Plans(mods...).All(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.List.All: %v", err)
		return nil, err
	}

	res := make([]model.Plan, len(plans))
	for i, p := range plans {
		res[i] = *model.NewPlanFromDB(p)
	}

	return res, nil
}

func (r *implRepository) Create(ctx context.Context, sc model.Scope, opts repository.CreateOptions) (model.Plan, error) {
	if opts.Plan.ID != "" {
		if err := postgresPkg.IsUUID(opts.Plan.ID); err != nil {
			r.l.Errorf(ctx, "internal.plan.repository.postgres.Create.IsUUID: %v", err)
			return model.Plan{}, err
		}
	}

	dbPlan := opts.Plan.ToDBPlan()
	if err := dbPlan.Insert(ctx, r.db, boil.Infer()); err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Create.Insert: %v", err)
		return model.Plan{}, err
	}

	plan, err := sqlboiler.Plans(
		sqlboiler.PlanWhere.ID.EQ(dbPlan.ID),
	).One(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Create.Reload: %v", err)
		return model.Plan{}, err
	}

	return *model.NewPlanFromDB(plan), nil
}

func (r *implRepository) Update(ctx context.Context, sc model.Scope, opts repository.UpdateOptions) (model.Plan, error) {
	if err := postgresPkg.IsUUID(opts.Plan.ID); err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Update.IsUUID: %v", err)
		return model.Plan{}, err
	}

	_, err := sqlboiler.Plans(
		sqlboiler.PlanWhere.ID.EQ(opts.Plan.ID),
		sqlboiler.PlanWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Plan{}, repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Update.Find: %v", err)
		return model.Plan{}, err
	}

	dbPlan := opts.Plan.ToDBPlan()
	rows, err := dbPlan.Update(ctx, r.db, boil.Infer())
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Update.Update: %v", err)
		return model.Plan{}, err
	}

	if rows == 0 {
		return model.Plan{}, repository.ErrNotFound
	}

	plan, err := sqlboiler.Plans(
		sqlboiler.PlanWhere.ID.EQ(opts.Plan.ID),
	).One(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Update.Reload: %v", err)
		return model.Plan{}, err
	}

	return *model.NewPlanFromDB(plan), nil
}

func (r *implRepository) GetOne(ctx context.Context, sc model.Scope, opts repository.GetOneOptions) (model.Plan, error) {
	mods, err := r.buildGetOneQuery(ctx, opts)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.GetOne.buildGetOneQuery: %v", err)
		return model.Plan{}, err
	}

	plan, err := sqlboiler.Plans(mods...).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Plan{}, repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.plan.repository.postgres.GetOne.One: %v", err)
		return model.Plan{}, err
	}

	return *model.NewPlanFromDB(plan), nil
}

func (r *implRepository) Get(ctx context.Context, sc model.Scope, opts repository.GetOptions) ([]model.Plan, paginator.Paginator, error) {
	mods, err := r.buildGetQuery(ctx, opts, opts.PaginateQuery)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Get.buildGetQuery: %v", err)
		return nil, paginator.Paginator{}, err
	}

	cntMods := []qm.QueryMod{
		sqlboiler.PlanWhere.DeletedAt.IsNull(),
	}
	if len(opts.Filter.IDs) > 0 {
		cntMods = append(cntMods, sqlboiler.PlanWhere.ID.IN(opts.Filter.IDs))
	}
	if len(opts.Filter.Codes) > 0 {
		cntMods = append(cntMods, sqlboiler.PlanWhere.Code.IN(opts.Filter.Codes))
	}

	total, err := sqlboiler.Plans(cntMods...).Count(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Get.Count: %v", err)
		return nil, paginator.Paginator{}, err
	}

	plans, err := sqlboiler.Plans(mods...).All(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Get.All: %v", err)
		return nil, paginator.Paginator{}, err
	}

	res := make([]model.Plan, len(plans))
	for i, p := range plans {
		res[i] = *model.NewPlanFromDB(p)
	}

	opts.PaginateQuery.Adjust()
	pag := paginator.Paginator{
		Total:       total,
		Count:       int64(len(res)),
		PerPage:     opts.PaginateQuery.Limit,
		CurrentPage: opts.PaginateQuery.Page,
	}

	return res, pag, nil
}

func (r *implRepository) Delete(ctx context.Context, sc model.Scope, id string) error {
	if err := postgresPkg.IsUUID(id); err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Delete.IsUUID: %v", err)
		return err
	}

	plan, err := sqlboiler.Plans(
		sqlboiler.PlanWhere.ID.EQ(id),
		sqlboiler.PlanWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Delete.Find: %v", err)
		return err
	}

	plan.DeletedAt = null.TimeFrom(r.clock())
	rows, err := plan.Update(ctx, r.db, boil.Whitelist(sqlboiler.PlanColumns.DeletedAt))
	if err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.Delete.Update: %v", err)
		return err
	}

	if rows == 0 {
		return repository.ErrNotFound
	}

	return nil
}

