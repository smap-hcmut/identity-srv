package postgres

import (
	"context"
	"database/sql"

	"smap-api/internal/model"
	"smap-api/internal/sqlboiler"
	"smap-api/internal/subscription/repository"
	"smap-api/pkg/paginator"
	postgresPkg "smap-api/pkg/postgre"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

func (r *implRepository) Detail(ctx context.Context, sc model.Scope, id string) (model.Subscription, error) {
	mods, err := r.buildDetailQuery(ctx, id)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Detail.buildDetailQuery: %v", err)
		return model.Subscription{}, err
	}

	sub, err := sqlboiler.Subscriptions(mods...).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Subscription{}, repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Detail.One: %v", err)
		return model.Subscription{}, err
	}

	return *model.NewSubscriptionFromDB(sub), nil
}

func (r *implRepository) List(ctx context.Context, sc model.Scope, opts repository.ListOptions) ([]model.Subscription, error) {
	mods, err := r.buildListQuery(ctx, opts)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.List.buildListQuery: %v", err)
		return nil, err
	}

	subs, err := sqlboiler.Subscriptions(mods...).All(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.List.All: %v", err)
		return nil, err
	}

	res := make([]model.Subscription, len(subs))
	for i, s := range subs {
		res[i] = *model.NewSubscriptionFromDB(s)
	}

	return res, nil
}

func (r *implRepository) Create(ctx context.Context, sc model.Scope, opts repository.CreateOptions) (model.Subscription, error) {
	if opts.Subscription.ID != "" {
		if err := postgresPkg.IsUUID(opts.Subscription.ID); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.Create.IsUUID: %v", err)
			return model.Subscription{}, err
		}
	}

	dbSub := opts.Subscription.ToDBSubscription()
	if err := dbSub.Insert(ctx, r.db, boil.Infer()); err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Create.Insert: %v", err)
		return model.Subscription{}, err
	}

	sub, err := sqlboiler.Subscriptions(
		sqlboiler.SubscriptionWhere.ID.EQ(dbSub.ID),
	).One(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Create.Reload: %v", err)
		return model.Subscription{}, err
	}

	return *model.NewSubscriptionFromDB(sub), nil
}

func (r *implRepository) Update(ctx context.Context, sc model.Scope, opts repository.UpdateOptions) (model.Subscription, error) {
	if err := postgresPkg.IsUUID(opts.Subscription.ID); err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Update.IsUUID: %v", err)
		return model.Subscription{}, err
	}

	_, err := sqlboiler.Subscriptions(
		sqlboiler.SubscriptionWhere.ID.EQ(opts.Subscription.ID),
		sqlboiler.SubscriptionWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Subscription{}, repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Update.Find: %v", err)
		return model.Subscription{}, err
	}

	dbSub := opts.Subscription.ToDBSubscription()
	rows, err := dbSub.Update(ctx, r.db, boil.Infer())
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Update.Update: %v", err)
		return model.Subscription{}, err
	}

	if rows == 0 {
		return model.Subscription{}, repository.ErrNotFound
	}

	sub, err := sqlboiler.Subscriptions(
		sqlboiler.SubscriptionWhere.ID.EQ(opts.Subscription.ID),
	).One(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Update.Reload: %v", err)
		return model.Subscription{}, err
	}

	return *model.NewSubscriptionFromDB(sub), nil
}

func (r *implRepository) GetOne(ctx context.Context, sc model.Scope, opts repository.GetOneOptions) (model.Subscription, error) {
	mods, err := r.buildGetOneQuery(ctx, opts)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.GetOne.buildGetOneQuery: %v", err)
		return model.Subscription{}, err
	}

	sub, err := sqlboiler.Subscriptions(mods...).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Subscription{}, repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.GetOne.One: %v", err)
		return model.Subscription{}, err
	}

	return *model.NewSubscriptionFromDB(sub), nil
}

func (r *implRepository) Get(ctx context.Context, sc model.Scope, opts repository.GetOptions) ([]model.Subscription, paginator.Paginator, error) {
	mods, err := r.buildGetQuery(ctx, opts, opts.PaginateQuery)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Get.buildGetQuery: %v", err)
		return nil, paginator.Paginator{}, err
	}

	cntMods := []qm.QueryMod{
		sqlboiler.SubscriptionWhere.DeletedAt.IsNull(),
	}
	if len(opts.Filter.IDs) > 0 {
		cntMods = append(cntMods, sqlboiler.SubscriptionWhere.ID.IN(opts.Filter.IDs))
	}
	if len(opts.Filter.UserIDs) > 0 {
		cntMods = append(cntMods, sqlboiler.SubscriptionWhere.UserID.IN(opts.Filter.UserIDs))
	}
	if len(opts.Filter.PlanIDs) > 0 {
		cntMods = append(cntMods, sqlboiler.SubscriptionWhere.PlanID.IN(opts.Filter.PlanIDs))
	}
	if len(opts.Filter.Statuses) > 0 {
		statuses := make([]string, len(opts.Filter.Statuses))
		for i, s := range opts.Filter.Statuses {
			statuses[i] = s.String()
		}
		cntMods = append(cntMods, qm.WhereIn("status IN ?", toInterfaceSlice(statuses)...))
	}

	total, err := sqlboiler.Subscriptions(cntMods...).Count(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Get.Count: %v", err)
		return nil, paginator.Paginator{}, err
	}

	subs, err := sqlboiler.Subscriptions(mods...).All(ctx, r.db)
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Get.All: %v", err)
		return nil, paginator.Paginator{}, err
	}

	res := make([]model.Subscription, len(subs))
	for i, s := range subs {
		res[i] = *model.NewSubscriptionFromDB(s)
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
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Delete.IsUUID: %v", err)
		return err
	}

	sub, err := sqlboiler.Subscriptions(
		sqlboiler.SubscriptionWhere.ID.EQ(id),
		sqlboiler.SubscriptionWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return repository.ErrNotFound
		}
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Delete.Find: %v", err)
		return err
	}

	sub.DeletedAt = null.TimeFrom(r.clock())
	rows, err := sub.Update(ctx, r.db, boil.Whitelist(sqlboiler.SubscriptionColumns.DeletedAt))
	if err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.Delete.Update: %v", err)
		return err
	}

	if rows == 0 {
		return repository.ErrNotFound
	}

	return nil
}

