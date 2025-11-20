package usecase

import (
	"context"

	"smap-api/internal/model"
	planPkg "smap-api/internal/plan"
	"smap-api/internal/subscription"
	"smap-api/internal/subscription/repository"
	postgresPkg "smap-api/pkg/postgre"
)

func (uc *usecase) Detail(ctx context.Context, sc model.Scope, id string) (subscription.SubscriptionOutput, error) {
	sub, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return subscription.SubscriptionOutput{}, subscription.ErrSubscriptionNotFound
		}
		uc.l.Errorf(ctx, "internal.subscription.usecase.Detail: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	return subscription.SubscriptionOutput{Subscription: sub}, nil
}

func (uc *usecase) List(ctx context.Context, sc model.Scope, ip subscription.ListInput) ([]model.Subscription, error) {
	opts := repository.ListOptions{
		Filter: repository.Filter{
			IDs:      ip.Filter.IDs,
			UserIDs:  ip.Filter.UserIDs,
			PlanIDs:  ip.Filter.PlanIDs,
			Statuses: ip.Filter.Statuses,
		},
	}

	subs, err := uc.repo.List(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "internal.subscription.usecase.List: %v", err)
		return nil, err
	}

	return subs, nil
}

func (uc *usecase) Get(ctx context.Context, sc model.Scope, ip subscription.GetInput) (subscription.GetSubscriptionOutput, error) {
	opts := repository.GetOptions{
		Filter: repository.Filter{
			IDs:      ip.Filter.IDs,
			UserIDs:  ip.Filter.UserIDs,
			PlanIDs:  ip.Filter.PlanIDs,
			Statuses: ip.Filter.Statuses,
		},
		PaginateQuery: ip.PaginateQuery,
	}

	subs, pag, err := uc.repo.Get(ctx, sc, opts)
	if err != nil {
		uc.l.Errorf(ctx, "internal.subscription.usecase.Get: %v", err)
		return subscription.GetSubscriptionOutput{}, err
	}

	return subscription.GetSubscriptionOutput{
		Subscriptions: subs,
		Paginator:     pag,
	}, nil
}

func (uc *usecase) Create(ctx context.Context, sc model.Scope, ip subscription.CreateInput) (subscription.SubscriptionOutput, error) {
	// Validate plan exists
	_, err := uc.planUC.GetOne(ctx, sc, planPkg.GetOneInput{ID: ip.PlanID})
	if err != nil {
		if err == planPkg.ErrPlanNotFound {
			uc.l.Warnf(ctx, "internal.subscription.usecase.Create.planUC.GetOne: %v", err)
			return subscription.SubscriptionOutput{}, planPkg.ErrPlanNotFound
		}
		uc.l.Errorf(ctx, "internal.subscription.usecase.Create.planUC.GetOne: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	// Check if user already has an active subscription
	activeStatuses := []model.SubscriptionStatus{
		model.SubscriptionStatusActive,
		model.SubscriptionStatusTrialing,
	}
	existingSubs, err := uc.repo.List(ctx, sc, repository.ListOptions{
		Filter: repository.Filter{
			UserIDs:  []string{ip.UserID},
			Statuses: activeStatuses,
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.subscription.usecase.Create.repo.List: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	if len(existingSubs) > 0 {
		uc.l.Warnf(ctx, "internal.subscription.usecase.Create: user already has active subscription")
		return subscription.SubscriptionOutput{}, subscription.ErrActiveSubscriptionExists
	}

	// Validate status
	if !ip.Status.IsValid() {
		return subscription.SubscriptionOutput{}, subscription.ErrInvalidStatus
	}

	sub := model.Subscription{
		ID:          postgresPkg.NewUUID(),
		UserID:      ip.UserID,
		PlanID:      ip.PlanID,
		Status:      ip.Status,
		TrialEndsAt: ip.TrialEndsAt,
		StartsAt:    ip.StartsAt,
		EndsAt:      ip.EndsAt,
	}

	created, err := uc.repo.Create(ctx, sc, repository.CreateOptions{Subscription: sub})
	if err != nil {
		uc.l.Errorf(ctx, "internal.subscription.usecase.Create: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	return subscription.SubscriptionOutput{Subscription: created}, nil
}

func (uc *usecase) GetOne(ctx context.Context, sc model.Scope, ip subscription.GetOneInput) (model.Subscription, error) {
	sub, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{
		ID:     ip.ID,
		UserID: ip.UserID,
		PlanID: ip.PlanID,
	})
	if err != nil {
		if err == repository.ErrNotFound {
			return model.Subscription{}, subscription.ErrSubscriptionNotFound
		}
		uc.l.Errorf(ctx, "internal.subscription.usecase.GetOne: %v", err)
		return model.Subscription{}, err
	}

	return sub, nil
}

func (uc *usecase) Update(ctx context.Context, sc model.Scope, ip subscription.UpdateInput) (subscription.SubscriptionOutput, error) {
	sub, err := uc.repo.Detail(ctx, sc, ip.ID)
	if err != nil {
		if err == repository.ErrNotFound {
			return subscription.SubscriptionOutput{}, subscription.ErrSubscriptionNotFound
		}
		uc.l.Errorf(ctx, "internal.subscription.usecase.Update.Detail: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	if ip.Status != nil {
		if !ip.Status.IsValid() {
			return subscription.SubscriptionOutput{}, subscription.ErrInvalidStatus
		}
		sub.Status = *ip.Status
	}
	if ip.TrialEndsAt != nil {
		sub.TrialEndsAt = ip.TrialEndsAt
	}
	if ip.EndsAt != nil {
		sub.EndsAt = ip.EndsAt
	}
	if ip.CancelledAt != nil {
		sub.CancelledAt = ip.CancelledAt
	}

	updated, err := uc.repo.Update(ctx, sc, repository.UpdateOptions{Subscription: sub})
	if err != nil {
		uc.l.Errorf(ctx, "internal.subscription.usecase.Update: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	return subscription.SubscriptionOutput{Subscription: updated}, nil
}

func (uc *usecase) Delete(ctx context.Context, sc model.Scope, id string) error {
	if err := uc.repo.Delete(ctx, sc, id); err != nil {
		if err == repository.ErrNotFound {
			return subscription.ErrSubscriptionNotFound
		}
		uc.l.Errorf(ctx, "internal.subscription.usecase.Delete: %v", err)
		return err
	}

	return nil
}

func (uc *usecase) GetActiveSubscription(ctx context.Context, sc model.Scope, userID string) (model.Subscription, error) {
	activeStatus := model.SubscriptionStatusActive
	sub, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{
		UserID: userID,
		Status: &activeStatus,
	})
	if err != nil {
		if err == repository.ErrNotFound {
			return model.Subscription{}, subscription.ErrSubscriptionNotFound
		}
		uc.l.Errorf(ctx, "internal.subscription.usecase.GetActiveSubscription: %v", err)
		return model.Subscription{}, err
	}

	return sub, nil
}

func (uc *usecase) Cancel(ctx context.Context, sc model.Scope, id string) (subscription.SubscriptionOutput, error) {
	sub, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return subscription.SubscriptionOutput{}, subscription.ErrSubscriptionNotFound
		}
		uc.l.Errorf(ctx, "internal.subscription.usecase.Cancel.Detail: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	// Only active or trialing subscriptions can be cancelled
	if sub.Status != model.SubscriptionStatusActive && sub.Status != model.SubscriptionStatusTrialing {
		return subscription.SubscriptionOutput{}, subscription.ErrCannotCancel
	}

	now := uc.clock()
	cancelledStatus := model.SubscriptionStatusCancelled
	sub.Status = cancelledStatus
	sub.CancelledAt = &now

	updated, err := uc.repo.Update(ctx, sc, repository.UpdateOptions{Subscription: sub})
	if err != nil {
		uc.l.Errorf(ctx, "internal.subscription.usecase.Cancel: %v", err)
		return subscription.SubscriptionOutput{}, err
	}

	return subscription.SubscriptionOutput{Subscription: updated}, nil
}

