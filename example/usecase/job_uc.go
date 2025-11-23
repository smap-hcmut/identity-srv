package usecase

import (
	"context"
	"slices"
	"sync"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/internal/resources/notification"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"golang.org/x/sync/errgroup"
)

func (uc implUseCase) CheckNotifyEvent() error {
	ctx := context.Background()

	now := util.Now()

	roundedNow := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, util.GetDefaultTimezone())

	startPeriod := now
	endPeriod := now.AddDate(0, 1, 0) // 1 month
	needRepeat := false

	var events []models.Event
	var grts []models.RecurringTracking
	var ris []models.RecurringInstance
	emptyScope := models.Scope{}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		events, err = uc.repo.SystemList(ctx, models.Scope{}, repository.SystemListOptions{
			NotifyTime: &roundedNow,
			StartTime:  startPeriod,
			EndTime:    endPeriod,
			NeedRepeat: &needRepeat,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.CheckNotifyEvent.ListEventsByNotifyTime: %v", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		grts, err = uc.repo.GetGenRTsInDateRange(ctx, models.Scope{}, startPeriod, endPeriod)
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.CheckNotifyEvent.GetGenRTsInDateRange: %v", err)
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		uc.l.Errorf(ctx, "event.usecase.CheckNotifyEvent.eg.Wait: %v", err)
		return err
	}

	if len(grts) > 0 {
		grtsIDs := make([]string, 0, len(grts))
		for _, e := range grts {
			grtsIDs = append(grtsIDs, e.EventID.Hex())
		}

		var err error
		ctx = context.Background()
		ris, err = uc.repo.ListRecurringInstancesByEventIDs(ctx, models.Scope{}, repository.ListEventInstancesByEventIDsOptions{
			EventIDs:   grtsIDs,
			StartTime:  startPeriod,
			EndTime:    endPeriod,
			NotifyTime: &roundedNow,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.CheckNotifyEvent.ListRecurringInstancesByEventIDs: %v", err)
			return err
		}
	}

	var eventInstances []event.EventInstance

	for _, e := range events {
		ei := event.EventToEventInstance(emptyScope, e)
		eventInstances = append(eventInstances, ei)
	}

	for _, ri := range ris {
		ei := event.RecurringInstanceToEventInstance(emptyScope, ri)
		eventInstances = append(eventInstances, ei)
	}

	shopIDs := make([]string, 0, len(eventInstances))
	for _, e := range eventInstances {
		shopIDs = append(shopIDs, e.ShopID.Hex())
	}

	var shops []microservice.Shop
	var err error
	if len(shopIDs) > 0 {
		shops, err = uc.shopUC.GetShopsWithoutAuth(ctx, microservice.GetShopsWithoutAuthFilter{
			IDs: shopIDs,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.CheckNotifyEvent.GetShopsWithoutAuth: %v", err)
			return err
		}
	}

	shopSuffixMap := make(map[string]string)
	for _, s := range shops {
		shopSuffixMap[s.ID] = s.Prefix
	}

	uIDsMap := make(map[string][]string)
	var uIDsMutex sync.Mutex
	var wg sync.WaitGroup

	for _, e := range eventInstances {
		wg.Add(1)
		go func(e event.EventInstance) {
			defer wg.Done()

			userIDs := []string{e.CreatedByID}
			userIDs = append(userIDs, e.AssignIDs...)

			if len(e.DepartmentIDs) > 0 || len(e.AssignIDs) > 0 {
				if len(e.DepartmentIDs) > 0 {
					deptIDs := make([]string, len(e.DepartmentIDs))
					for i, deptID := range e.DepartmentIDs {
						deptIDs[i] = deptID.Hex()
					}

					us, err := uc.shopUC.ListAllUsersUnAuth(ctx, microservice.GetUsersFilter{
						DeptIDs: deptIDs,
					})
					if err != nil {
						uc.l.Errorf(ctx, "event.usecase.job_uc.ListAllUsers: %v", err)
					}

					for _, u := range us {
						userIDs = append(userIDs, u.ID)
					}
				}
			} else if len(e.BranchIDs) > 0 {
				brIDs := make([]string, len(e.BranchIDs))
				for i, brID := range e.BranchIDs {
					brIDs[i] = brID.Hex()
				}

				us, err := uc.shopUC.ListAllUsersUnAuth(ctx, microservice.GetUsersFilter{
					BranchIds: brIDs,
				})
				if err != nil {
					uc.l.Errorf(ctx, "event.usecase.job_uc.ListAllUsers: %v", err)
				}

				for _, u := range us {
					userIDs = append(userIDs, u.ID)
				}
			}

			userIDs = slices.DeleteFunc(userIDs, func(uID string) bool {
				return slices.Contains(e.DeclinedIDs, uID)
			})

			userIDs = util.RemoveDuplicates(userIDs)

			uIDsMutex.Lock()
			uIDsMap[e.ID.Hex()] = userIDs
			uIDsMutex.Unlock()
		}(e)
	}

	wg.Wait()

	wg = sync.WaitGroup{}

	for _, e := range eventInstances {
		wg.Add(1)
		go func(e event.EventInstance) {
			defer wg.Done()
			sc := models.Scope{
				ShopID: e.ShopID.Hex(),
				Suffix: shopSuffixMap[e.ShopID.Hex()],
			}

			startTimeUTC := e.StartTime.UTC()
			roundedNowUTC := roundedNow.UTC()
			duration := startTimeUTC.Sub(roundedNowUTC)

			notiInput := getEventNotiInput{
				EI:         e,
				Type:       notification.SourceEventReminder,
				EventTitle: e.Title,
				UserIDs:    uIDsMap[e.ID.Hex()],
				Duration:   duration,
			}

			noti, err := uc.getEventNoti(ctx, sc, notiInput)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.CheckNotifyEvent.getEventNoti: %v", err)
			}

			err = uc.publishPushNotiMsg(ctx, noti)
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.CheckNotifyEvent.publishPushNotiMsg: %v", err)
			}
		}(e)
	}

	wg.Wait()

	return nil
}
