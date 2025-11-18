package postgres

import (
	"context"

	"smap-api/internal/plan/repository"
	"smap-api/internal/sqlboiler"
	"smap-api/pkg/paginator"
	postgresPkg "smap-api/pkg/postgre"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

func (r *implRepository) buildListQuery(ctx context.Context, opts repository.ListOptions) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		sqlboiler.PlanWhere.DeletedAt.IsNull(),
	}

	if len(opts.Filter.IDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.IDs); err != nil {
			r.l.Errorf(ctx, "internal.plan.repository.postgres.buildListQuery.ValidateUUIDs: %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.PlanWhere.ID.IN(opts.Filter.IDs))
	}

	if len(opts.Filter.Codes) > 0 {
		mods = append(mods, sqlboiler.PlanWhere.Code.IN(opts.Filter.Codes))
	}

	return mods, nil
}

func (r *implRepository) buildDetailQuery(ctx context.Context, id string) ([]qm.QueryMod, error) {
	if err := postgresPkg.IsUUID(id); err != nil {
		r.l.Errorf(ctx, "internal.plan.repository.postgres.buildDetailQuery.IsUUID: %v", err)
		return nil, err
	}

	return []qm.QueryMod{
		sqlboiler.PlanWhere.ID.EQ(id),
		sqlboiler.PlanWhere.DeletedAt.IsNull(),
	}, nil
}

func (r *implRepository) buildGetOneQuery(ctx context.Context, opts repository.GetOneOptions) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		sqlboiler.PlanWhere.DeletedAt.IsNull(),
	}

	if opts.ID != "" {
		if err := postgresPkg.IsUUID(opts.ID); err != nil {
			r.l.Errorf(ctx, "internal.plan.repository.postgres.buildGetOneQuery.IsUUID: %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.PlanWhere.ID.EQ(opts.ID))
	} else if opts.Code != "" {
		mods = append(mods, sqlboiler.PlanWhere.Code.EQ(opts.Code))
	}

	return mods, nil
}

func (r *implRepository) buildGetQuery(ctx context.Context, opts repository.GetOptions, pq paginator.PaginateQuery) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		sqlboiler.PlanWhere.DeletedAt.IsNull(),
	}

	if len(opts.Filter.IDs) > 0 {
		if err := postgresPkg.ValidateUUIDs(opts.Filter.IDs); err != nil {
			r.l.Errorf(ctx, "internal.plan.repository.postgres.buildGetQuery.ValidateUUIDs: %v", err)
			return nil, err
		}
		mods = append(mods, sqlboiler.PlanWhere.ID.IN(opts.Filter.IDs))
	}

	if len(opts.Filter.Codes) > 0 {
		mods = append(mods, sqlboiler.PlanWhere.Code.IN(opts.Filter.Codes))
	}

	pq.Adjust()
	mods = append(mods,
		qm.Limit(int(pq.Limit)),
		qm.Offset(int(pq.Offset())),
		qm.OrderBy(sqlboiler.PlanColumns.CreatedAt+" DESC"),
	)

	return mods, nil
}

