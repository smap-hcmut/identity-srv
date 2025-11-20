package usecase

import (
	"context"

	"smap-api/internal/model"
	"smap-api/internal/plan"
	"smap-api/internal/plan/repository"
	postgresPkg "smap-api/pkg/postgre"
)

func (uc *usecase) Detail(ctx context.Context, sc model.Scope, id string) (plan.PlanOutput, error) {
	p, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return plan.PlanOutput{}, plan.ErrPlanNotFound
		}
		uc.l.Errorf(ctx, "internal.plan.usecase.Detail: %v", err)
		return plan.PlanOutput{}, err
	}

	return plan.PlanOutput{Plan: p}, nil
}

func (uc *usecase) List(ctx context.Context, sc model.Scope, ip plan.ListInput) ([]model.Plan, error) {
	opts := repository.ListOptions{
		Filter: repository.Filter{
			IDs:   ip.Filter.IDs,
			Codes: ip.Filter.Codes,
		},
	}

	plans, err := uc.repo.List(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "internal.plan.usecase.List: %v", err)
		return nil, err
	}

	return plans, nil
}

func (uc *usecase) Get(ctx context.Context, sc model.Scope, ip plan.GetInput) (plan.GetPlanOutput, error) {
	opts := repository.GetOptions{
		Filter: repository.Filter{
			IDs:   ip.Filter.IDs,
			Codes: ip.Filter.Codes,
		},
		PaginateQuery: ip.PaginateQuery,
	}

	plans, pag, err := uc.repo.Get(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "internal.plan.usecase.Get: %v", err)
		return plan.GetPlanOutput{}, err
	}

	return plan.GetPlanOutput{
		Plans:     plans,
		Paginator: pag,
	}, nil
}

func (uc *usecase) Create(ctx context.Context, sc model.Scope, ip plan.CreateInput) (plan.PlanOutput, error) {
	// Check if plan code already exists
	_, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{Code: ip.Code})
	if err == nil {
		return plan.PlanOutput{}, plan.ErrPlanCodeExists
	}
	if err != repository.ErrNotFound {
		uc.l.Errorf(ctx, "internal.plan.usecase.Create.GetOne: %v", err)
		return plan.PlanOutput{}, err
	}

	p := model.Plan{
		ID:          postgresPkg.NewUUID(),
		Name:        ip.Name,
		Code:        ip.Code,
		Description: ip.Description,
		MaxUsage:    ip.MaxUsage,
	}

	created, err := uc.repo.Create(ctx, sc, repository.CreateOptions{Plan: p})
	if err != nil {
		uc.l.Errorf(ctx, "internal.plan.usecase.Create: %v", err)
		return plan.PlanOutput{}, err
	}

	return plan.PlanOutput{Plan: created}, nil
}

func (uc *usecase) GetOne(ctx context.Context, sc model.Scope, ip plan.GetOneInput) (model.Plan, error) {
	p, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{
		Code: ip.Code,
		ID:   ip.ID,
	})
	if err != nil {
		if err == repository.ErrNotFound {
			return model.Plan{}, plan.ErrPlanNotFound
		}
		uc.l.Errorf(ctx, "internal.plan.usecase.GetOne: %v", err)
		return model.Plan{}, err
	}

	return p, nil
}

func (uc *usecase) Update(ctx context.Context, sc model.Scope, ip plan.UpdateInput) (plan.PlanOutput, error) {
	p, err := uc.repo.Detail(ctx, sc, ip.ID)
	if err != nil {
		if err == repository.ErrNotFound {
			return plan.PlanOutput{}, plan.ErrPlanNotFound
		}
		uc.l.Errorf(ctx, "internal.plan.usecase.Update.Detail: %v", err)
		return plan.PlanOutput{}, err
	}

	if ip.Name != nil {
		p.Name = *ip.Name
	}
	if ip.Description != nil {
		p.Description = ip.Description
	}
	if ip.MaxUsage != nil {
		p.MaxUsage = *ip.MaxUsage
	}

	updated, err := uc.repo.Update(ctx, sc, repository.UpdateOptions{Plan: p})
	if err != nil {
		uc.l.Errorf(ctx, "internal.plan.usecase.Update: %v", err)
		return plan.PlanOutput{}, err
	}

	return plan.PlanOutput{Plan: updated}, nil
}

func (uc *usecase) Delete(ctx context.Context, sc model.Scope, id string) error {
	if err := uc.repo.Delete(ctx, sc, id); err != nil {
		if err == repository.ErrNotFound {
			return plan.ErrPlanNotFound
		}
		uc.l.Errorf(ctx, "internal.plan.usecase.Delete: %v", err)
		return err
	}

	return nil
}

