package usecase

import (
	"context"

	"smap-project/internal/keyword"
	"smap-project/internal/model"
	"smap-project/internal/project"
	"smap-project/internal/project/repository"
)

func (uc *usecase) Detail(ctx context.Context, sc model.Scope, id string) (project.ProjectOutput, error) {
	p, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		if err == repository.ErrNotFound {
			uc.l.Warnf(ctx, "internal.project.usecase.Detail: %v", err)
			return project.ProjectOutput{}, project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.Detail: %v", err)
		return project.ProjectOutput{}, err
	}

	// Check if user owns this project
	if p.CreatedBy != sc.UserID {
		uc.l.Warnf(ctx, "internal.project.usecase.Detail: user %s does not own project %s", sc.UserID, id)
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
	// Validate date range
	if ip.ToDate.Before(ip.FromDate) || ip.ToDate.Equal(ip.FromDate) {
		uc.l.Warnf(ctx, "internal.project.usecase.Create: invalid date range %s - %s", ip.FromDate, ip.ToDate)
		return project.ProjectOutput{}, project.ErrInvalidDateRange
	}

	// Validate and normalize brand keywords
	brandKeywords := ip.BrandKeywords
	if len(brandKeywords) > 0 {
		validateOut, err := uc.keywordUC.Validate(ctx, keyword.ValidateInput{Keywords: brandKeywords})
		if err != nil {
			uc.l.Errorf(ctx, "internal.project.usecase.Create: %v", err)
			return project.ProjectOutput{}, err
		}
		brandKeywords = validateOut.ValidKeywords
	}

	// Validate and normalize competitor keywords
	competitorKeywords := make([]model.CompetitorKeyword, 0, len(ip.CompetitorKeywords))
	for _, ck := range ip.CompetitorKeywords {
		if len(ck.Keywords) > 0 {
			validateOut, err := uc.keywordUC.Validate(ctx, keyword.ValidateInput{Keywords: ck.Keywords})
			if err != nil {
				uc.l.Errorf(ctx, "internal.project.usecase.Create: %v", err)
				return project.ProjectOutput{}, err
			}
			competitorKeywords = append(competitorKeywords, model.CompetitorKeyword{
				CompetitorName: ck.Name,
				Keywords:       validateOut.ValidKeywords,
			})
		}
	}

	// Extract competitor names from competitor keywords
	competitorNames := make([]string, 0, len(competitorKeywords))
	for _, ck := range competitorKeywords {
		competitorNames = append(competitorNames, ck.CompetitorName)
	}

	p, err := uc.repo.Create(ctx, sc, repository.CreateOptions{
		Name:               ip.Name,
		Description:        ip.Description,
		FromDate:           ip.FromDate,
		ToDate:             ip.ToDate,
		BrandName:          ip.BrandName,
		CompetitorNames:    competitorNames,
		BrandKeywords:      brandKeywords,
		CompetitorKeywords: competitorKeywords,
		CreatedBy:          sc.UserID,
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.Create: %v", err)
		return project.ProjectOutput{}, err
	}

	return project.ProjectOutput{
		Project: p,
	}, nil
}

func (uc *usecase) GetOne(ctx context.Context, sc model.Scope, ip project.GetOneInput) (model.Project, error) {
	p, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{
		ID: ip.ID,
	})
	if err != nil {
		if err == repository.ErrNotFound {
			uc.l.Warnf(ctx, "internal.project.usecase.GetOne: project %s not found", ip.ID)
			return model.Project{}, project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.GetOne: %v", err)
		return model.Project{}, err
	}

	// Check if user owns this project
	if p.CreatedBy != sc.UserID {
		uc.l.Warnf(ctx, "internal.project.usecase.GetOne: user %s does not own project %s", sc.UserID, ip.ID)
		return model.Project{}, project.ErrUnauthorized
	}

	return p, nil
}

func (uc *usecase) Patch(ctx context.Context, sc model.Scope, ip project.PatchInput) (project.ProjectOutput, error) {
	p, err := uc.repo.Detail(ctx, sc, ip.ID)
	if err != nil {
		if err == repository.ErrNotFound {
			uc.l.Warnf(ctx, "internal.project.usecase.Patch.Detail: project %s not found", ip.ID)
			return project.ProjectOutput{}, project.ErrProjectNotFound
		}
		uc.l.Errorf(ctx, "internal.project.usecase.Patch.Detail: %v", err)
		return project.ProjectOutput{}, err
	}

	opts := repository.UpdateOptions{
		ID:          ip.ID,
		Description: ip.Description,
		Status:      ip.Status,
		FromDate:    ip.FromDate,
		ToDate:      ip.ToDate,
	}

	// Check if user owns this project
	if p.CreatedBy != sc.UserID {
		uc.l.Warnf(ctx, "internal.project.usecase.Patch: user %s does not own project %s", sc.UserID, ip.ID)
		return project.ProjectOutput{}, project.ErrUnauthorized
	}

	// Validate and normalize brand keywords
	brandKeywords := ip.BrandKeywords
	if len(brandKeywords) > 0 {
		validateOut, err := uc.keywordUC.Validate(ctx, keyword.ValidateInput{Keywords: brandKeywords})
		if err != nil {
			uc.l.Errorf(ctx, "internal.project.usecase.Create: %v", err)
			return project.ProjectOutput{}, err
		}
		brandKeywords = validateOut.ValidKeywords
	}
	opts.BrandKeywords = brandKeywords

	// Validate and normalize competitor keywords
	competitorKeywords := make([]model.CompetitorKeyword, 0, len(ip.CompetitorKeywords))
	for _, ck := range ip.CompetitorKeywords {
		if len(ck.Keywords) > 0 {
			validateOut, err := uc.keywordUC.Validate(ctx, keyword.ValidateInput{Keywords: ck.Keywords})
			if err != nil {
				uc.l.Errorf(ctx, "internal.project.usecase.Create: %v", err)
				return project.ProjectOutput{}, err
			}
			competitorKeywords = append(competitorKeywords, model.CompetitorKeyword{
				CompetitorName: ck.Name,
				Keywords:       validateOut.ValidKeywords,
			})
		}
	}
	opts.CompetitorKeywords = competitorKeywords

	up, err := uc.repo.Update(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.Update: %v", err)
		return project.ProjectOutput{}, err
	}

	return project.ProjectOutput{Project: up}, nil
}

func (uc *usecase) Delete(ctx context.Context, sc model.Scope, ip project.DeleteInput) error {
	// Check if project exists and user owns it
	p, err := uc.repo.List(ctx, sc, repository.ListOptions{
		IDs: ip.IDs,
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.Delete.repo.List: %v", err)
		return err
	}

	if len(p) != len(ip.IDs) {
		uc.l.Warnf(ctx, "internal.project.usecase.Delete.someProjectsNotFound: %v", ip.IDs)
		return project.ErrProjectNotFound
	}

	for _, proj := range p {
		if proj.CreatedBy != sc.UserID {
			uc.l.Warnf(ctx, "internal.project.usecase.Delete: user %s does not own project %s", sc.UserID, proj.ID)
			return project.ErrUnauthorized
		}
	}

	if err := uc.repo.Delete(ctx, sc, ip.IDs); err != nil {
		uc.l.Errorf(ctx, "internal.project.usecase.Delete.repo.Delete: %v", err)
		return err
	}

	return nil
}
