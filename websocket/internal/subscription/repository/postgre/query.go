package postgres

import (
	"context"

	"smap-api/internal/sqlboiler"
	"smap-api/internal/subscription/repository"
	"smap-api/pkg/paginator"
	postgresPkg "smap-api/pkg/postgre"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

func (r *implRepository) buildListQuery(ctx context.Context, opts repository.ListOptions) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		sqlboiler.SubscriptionWhere.DeletedAt.IsNull(),
	}

	if len(opts.Filter.IDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.IDs); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildListQuery.ValidateUUIDs: %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.ID.IN(opts.Filter.IDs))
	}

	if len(opts.Filter.UserIDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.UserIDs); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildListQuery.ValidateUUIDs(UserIDs): %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.UserID.IN(opts.Filter.UserIDs))
	}

	if len(opts.Filter.PlanIDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.PlanIDs); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildListQuery.ValidateUUIDs(PlanIDs): %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.PlanID.IN(opts.Filter.PlanIDs))
	}

	if len(opts.Filter.Statuses) > 0 {
		statuses := make([]string, len(opts.Filter.Statuses))
		for i, s := range opts.Filter.Statuses {
			statuses[i] = s.String()
		}
		mods = append(mods, qm.WhereIn("status IN ?", toInterfaceSlice(statuses)...))
	}

	return mods, nil
}

func (r *implRepository) buildDetailQuery(ctx context.Context, id string) ([]qm.QueryMod, error) {
	if err := postgresPkg.IsUUID(id); err != nil {
		r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildDetailQuery.IsUUID: %v", err)
		return nil, err
	}

	return []qm.QueryMod{
		sqlboiler.SubscriptionWhere.ID.EQ(id),
		sqlboiler.SubscriptionWhere.DeletedAt.IsNull(),
	}, nil
}

func (r *implRepository) buildGetOneQuery(ctx context.Context, opts repository.GetOneOptions) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		sqlboiler.SubscriptionWhere.DeletedAt.IsNull(),
	}

	if opts.ID != "" {
		if err := postgresPkg.IsUUID(opts.ID); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildGetOneQuery.IsUUID: %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.ID.EQ(opts.ID))
	}

	if opts.UserID != "" {
		if err := postgresPkg.IsUUID(opts.UserID); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildGetOneQuery.IsUUID(UserID): %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.UserID.EQ(opts.UserID))
	}

	if opts.PlanID != "" {
		if err := postgresPkg.IsUUID(opts.PlanID); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildGetOneQuery.IsUUID(PlanID): %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.PlanID.EQ(opts.PlanID))
	}

	if opts.Status != nil {
		mods = append(mods, qm.Where("status = ?", opts.Status.String()))
	}

	return mods, nil
}

func (r *implRepository) buildGetQuery(ctx context.Context, opts repository.GetOptions, pq paginator.PaginateQuery) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		sqlboiler.SubscriptionWhere.DeletedAt.IsNull(),
	}

	if len(opts.Filter.IDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.IDs); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildGetQuery.ValidateUUIDs: %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.ID.IN(opts.Filter.IDs))
	}

	if len(opts.Filter.UserIDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.UserIDs); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildGetQuery.ValidateUUIDs(UserIDs): %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.UserID.IN(opts.Filter.UserIDs))
	}

	if len(opts.Filter.PlanIDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.PlanIDs); err != nil {
			r.l.Errorf(ctx, "internal.subscription.repository.postgres.buildGetQuery.ValidateUUIDs(PlanIDs): %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.SubscriptionWhere.PlanID.IN(opts.Filter.PlanIDs))
	}

	if len(opts.Filter.Statuses) > 0 {
		statuses := make([]string, len(opts.Filter.Statuses))
		for i, s := range opts.Filter.Statuses {
			statuses[i] = s.String()
		}
		mods = append(mods, qm.WhereIn("status IN ?", toInterfaceSlice(statuses)...))
	}

	pq.Adjust()
	mods = append(mods,
		qm.Limit(int(pq.Limit)),
		qm.Offset(int(pq.Offset())),
		qm.OrderBy(sqlboiler.SubscriptionColumns.CreatedAt+" DESC"),
	)

	return mods, nil
}

func toInterfaceSlice(s []string) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

