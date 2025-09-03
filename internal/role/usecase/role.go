package usecase

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
)

func (uc implUsecase) Create(ctx context.Context, sc models.Scope, input role.CreateInput) (role.CreateOutput, error) {
	r, err := uc.repo.Create(ctx, sc, role.CreateOptions{
		Name: input.Name,
	})
	if err != nil {
		uc.l.Errorf(ctx, "role.usecase.Create.repo.Create: %v", err)
		return role.CreateOutput{}, err
	}

	return role.CreateOutput{
		Role: r,
	}, nil
}

func (uc implUsecase) Update(ctx context.Context, sc models.Scope, input role.UpdateInput) (role.UpdateOutput, error) {
	existing, err := uc.repo.Detail(ctx, sc, input.ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "role.usecase.Update.repo.Detail: %v", err)
			return role.UpdateOutput{}, role.ErrRoleNotFound
		}
		uc.l.Errorf(ctx, "role.usecase.Update.repo.Detail: %v", err)
		return role.UpdateOutput{}, err
	}

	r, err := uc.repo.Update(ctx, sc, role.UpdateOptions{
		Model: existing,
		Name:  input.Name,
	})
	if err != nil {
		uc.l.Errorf(ctx, "role.usecase.Update.repo.Update: %v", err)
		return role.UpdateOutput{}, err
	}

	return role.UpdateOutput{
		Role: r,
	}, nil
}

func (uc implUsecase) Delete(ctx context.Context, sc models.Scope, ids []string) error {
	err := uc.repo.Delete(ctx, sc, ids)
	if err != nil {
		uc.l.Errorf(ctx, "role.usecase.Delete.repo.Delete: %v", err)
		return err
	}

	return nil
}

func (uc implUsecase) Detail(ctx context.Context, sc models.Scope, id string) (role.DetailOutput, error) {
	r, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "role.usecase.Detail.repo.Detail: %v", err)
			return role.DetailOutput{}, role.ErrRoleNotFound
		}
		uc.l.Errorf(ctx, "role.usecase.Detail.repo.Detail: %v", err)
		return role.DetailOutput{}, err
	}

	return role.DetailOutput{
		Role: r,
	}, nil
}

func (uc implUsecase) GetOne(ctx context.Context, sc models.Scope, ip role.GetOneInput) (role.GetOneOutput, error) {
	r, err := uc.repo.GetOne(ctx, sc, role.GetOneOptions{Filter: ip.Filter})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "role.usecase.GetOne.repo.GetOne: %v", err)
			return role.GetOneOutput{}, role.ErrRoleNotFound
		}
		uc.l.Errorf(ctx, "role.usecase.GetOne.repo.GetOne: %v", err)
		return role.GetOneOutput{}, err
	}

	return role.GetOneOutput{
		Role: r,
	}, nil
}

func (uc implUsecase) Get(ctx context.Context, sc models.Scope, ip role.GetInput) (role.GetOutput, error) {
	rs, pag, err := uc.repo.Get(ctx, sc, role.GetOptions{Filter: ip.Filter, PagQuery: ip.PagQuery})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "role.usecase.Get.repo.Get: %v", err)
			return role.GetOutput{}, role.ErrRoleNotFound
		}
		uc.l.Errorf(ctx, "role.usecase.Get.repo.Get: %v", err)
		return role.GetOutput{}, err
	}

	return role.GetOutput{
		Roles:     rs,
		Paginator: pag,
	}, nil
}

func (uc implUsecase) List(ctx context.Context, sc models.Scope, ip role.ListInput) (role.ListOutput, error) {
	rs, err := uc.repo.List(ctx, sc, role.ListOptions{Filter: ip.Filter})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "role.usecase.List.repo.List: %v", err)
			return role.ListOutput{}, role.ErrRoleNotFound
		}
		uc.l.Errorf(ctx, "role.usecase.List.repo.List: %v", err)
		return role.ListOutput{}, err
	}

	return role.ListOutput{
		Roles: rs,
	}, nil
}
