package postgre

import (
	"context"
	"smap-api/internal/model"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// buildGetActiveKeyQuery builds query to get active key
func (r *implRepository) buildGetActiveKeyQuery(ctx context.Context) []qm.QueryMod {
	return []qm.QueryMod{
		qm.Where("status = ?", model.KeyStatusActive),
		qm.OrderBy("created_at DESC"),
	}
}

// buildGetActiveAndRotatingKeysQuery builds query to get active and rotating keys
func (r *implRepository) buildGetActiveAndRotatingKeysQuery(ctx context.Context) []qm.QueryMod {
	return []qm.QueryMod{
		qm.WhereIn("status IN ?", model.KeyStatusActive, model.KeyStatusRotating),
		qm.OrderBy("created_at DESC"),
	}
}

// buildUpdateKeyStatusQuery builds query to find key by KID for update
func (r *implRepository) buildUpdateKeyStatusQuery(ctx context.Context, kid string) []qm.QueryMod {
	return []qm.QueryMod{
		qm.Where("kid = ?", kid),
	}
}

// buildGetRotatingKeysQuery builds query to get rotating keys
func (r *implRepository) buildGetRotatingKeysQuery(ctx context.Context) []qm.QueryMod {
	return []qm.QueryMod{
		qm.Where("status = ?", model.KeyStatusRotating),
	}
}
