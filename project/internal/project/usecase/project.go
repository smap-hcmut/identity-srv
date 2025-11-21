package usecase

import (
	"context"

	"smap-project/internal/model"
	"smap-project/internal/project"
	"smap-project/internal/project/repository"
)

func (uc *usecase) Detail(ctx context.Context, sc model.Scope, id string) (project.ProjectOutput, error) {
	p, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return project.ProjectOutput{}, project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.Detail: %v", err)
		return project.ProjectOutput{}, err
	}

	// Check if user owns this project
	if p.CreatedBy != sc.UserID {
		return project.ProjectOutput{}, project.ErrUnauthorized
	}

	return project.ProjectOutput{Project: p}, nil
}

func (uc *usecase) List(ctx context.Context, sc model.Scope, ip project.ListInput) ([]model.Project, error) {
	// Users can only see their own projects
	userID := sc.UserID

	opts := repository.ListOptions{
		IDs:        ip.Filter.IDs,
		Statuses:   ip.Filter.Statuses,
		CreatedBy:  &userID,
		SearchName: ip.Filter.SearchName,
	}

	projects, err := uc.repo.List(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.List: %v", err)
		return nil, err
	}

	return projects, nil
}

func (uc *usecase) Get(ctx context.Context, sc model.Scope, ip project.GetInput) (project.GetProjectOutput, error) {
	// Users can only see their own projects
	userID := sc.UserID

	opts := repository.GetOptions{
		IDs:           ip.Filter.IDs,
		Statuses:      ip.Filter.Statuses,
		CreatedBy:     &userID,
		SearchName:    ip.Filter.SearchName,
		PaginateQuery: ip.PaginateQuery,
	}

	projects, pag, err := uc.repo.Get(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.Get: %v", err)
		return project.GetProjectOutput{}, err
	}

	return project.GetProjectOutput{
		Projects:  projects,
		Paginator: pag,
	}, nil
}

func (uc *usecase) Create(ctx context.Context, sc model.Scope, ip project.CreateInput) (project.ProjectOutput, error) {
	// Validate status
	if !model.IsValidProjectStatus(ip.Status) {
		return project.ProjectOutput{}, project.ErrInvalidStatus
	}

	// Validate date range
	if ip.ToDate.Before(ip.FromDate) || ip.ToDate.Equal(ip.FromDate) {
		return project.ProjectOutput{}, project.ErrInvalidDateRange
	}

	p := model.Project{
		Name:                  ip.Name,
		Description:           ip.Description,
		Status:                ip.Status,
		FromDate:              ip.FromDate,
		ToDate:                ip.ToDate,
		BrandName:             ip.BrandName,
		CompetitorNames:       ip.CompetitorNames,
		BrandKeywords:         ip.BrandKeywords,
		CompetitorKeywordsMap: ip.CompetitorKeywordsMap,
		CreatedBy:             sc.UserID,
		CreatedAt:             uc.clock(),
		UpdatedAt:             uc.clock(),
	}

	created, err := uc.repo.Create(ctx, sc, repository.CreateOptions{Project: p})
	if err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.Create: %v", err)
		return project.ProjectOutput{}, err
	}

	return project.ProjectOutput{Project: created}, nil
}

func (uc *usecase) GetOne(ctx context.Context, sc model.Scope, ip project.GetOneInput) (model.Project, error) {
	p, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{
		ID: ip.ID,
	})
	if err != nil {
		if err == repository.ErrNotFound {
			return model.Project{}, project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.GetOne: %v", err)
		return model.Project{}, err
	}

	// Check if user owns this project
	if p.CreatedBy != sc.UserID {
		return model.Project{}, project.ErrUnauthorized
	}

	return p, nil
}

func (uc *usecase) Update(ctx context.Context, sc model.Scope, ip project.UpdateInput) (project.ProjectOutput, error) {
	p, err := uc.repo.Detail(ctx, sc, ip.ID)
	if err != nil {
		if err == repository.ErrNotFound {
			return project.ProjectOutput{}, project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.Update.Detail: %v", err)
		return project.ProjectOutput{}, err
	}

	// Check if user owns this project
	if p.CreatedBy != sc.UserID {
		return project.ProjectOutput{}, project.ErrUnauthorized
	}

	// Update fields
	if ip.Name != nil {
		p.Name = *ip.Name
	}
	if ip.Description != nil {
		p.Description = ip.Description
	}
	if ip.Status != nil {
		if !model.IsValidProjectStatus(*ip.Status) {
			return project.ProjectOutput{}, project.ErrInvalidStatus
		}
		p.Status = *ip.Status
	}
	if ip.FromDate != nil {
		p.FromDate = *ip.FromDate
	}
	if ip.ToDate != nil {
		p.ToDate = *ip.ToDate
	}

	// Validate date range after updates
	if p.ToDate.Before(p.FromDate) || p.ToDate.Equal(p.FromDate) {
		return project.ProjectOutput{}, project.ErrInvalidDateRange
	}

	if ip.BrandName != nil {
		p.BrandName = *ip.BrandName
	}
	if ip.CompetitorNames != nil {
		p.CompetitorNames = ip.CompetitorNames
	}
	if ip.BrandKeywords != nil {
		p.BrandKeywords = ip.BrandKeywords
	}
	if ip.CompetitorKeywordsMap != nil {
		p.CompetitorKeywordsMap = ip.CompetitorKeywordsMap
	}

	p.UpdatedAt = uc.clock()

	updated, err := uc.repo.Update(ctx, sc, repository.UpdateOptions{Project: p})
	if err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.Update: %v", err)
		return project.ProjectOutput{}, err
	}

	return project.ProjectOutput{Project: updated}, nil
}

func (uc *usecase) Delete(ctx context.Context, sc model.Scope, id string) error {
	// Check if project exists and user owns it
	p, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.Delete.Detail: %v", err)
		return err
	}

	// Check if user owns this project
	if p.CreatedBy != sc.UserID {
		return project.ErrUnauthorized
	}

	if err := uc.repo.Delete(ctx, sc, id); err != nil {
		if err == repository.ErrNotFound {
			return project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.Delete: %v", err)
		return err
	}

	return nil
}
