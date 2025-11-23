package usecase

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/element"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
	"gitlab.com/gma-vietnam/tanca-connect/internal/resources/notification"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/locale"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Default system time zone offset
const UTC_PLUS_7_OFFSET int = 7 * 3600

// Removed eventNotificationTypes array - duration formatting is now handled at the call site

func (uc implUseCase) validateAssign(ctx context.Context, sc models.Scope, assignIDs []string, branchIDs []string, departmentIDs []string) error {
	if len(assignIDs) == 0 && len(branchIDs) == 0 && len(departmentIDs) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	var wgErr error

	// Validate branches
	if len(branchIDs) > 0 {
		wg.Add(1)
		go func(brIDs []string) {
			defer wg.Done()
			branches, err := uc.shopUC.GetBranches(ctx, sc, microservice.GetBranchesFilter{
				IDs: brIDs,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.validateAssign.GetBranches: %v", err)
				wgErr = err
				return
			}

			if len(branches) == 0 {
				uc.l.Errorf(ctx, "event.usecase.validateAssign.GetBranches: %v", microservice.ErrBranchNotFound)
				wgErr = microservice.ErrBranchNotFound
				return
			}

			// Update branchIDs with valid IDs
			branchIDs = make([]string, 0)
			for _, branch := range branches {
				branchIDs = append(branchIDs, branch.ID)
			}
		}(branchIDs)
	}

	// Validate departments
	if len(departmentIDs) > 0 {
		wg.Add(1)
		go func(deptIDs []string) {
			defer wg.Done()
			depts, err := uc.shopUC.GetDepartments(ctx, sc, microservice.GetDepartmentsFilter{
				IDs: deptIDs,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.validateAssign.GetDepartments: %v", err)
				wgErr = err
				return
			}

			if len(depts) == 0 {
				uc.l.Errorf(ctx, "event.usecase.validateAssign.GetDepartments: %v", microservice.ErrDepartmentNotFound)
				wgErr = microservice.ErrDepartmentNotFound
				return
			}

			// Update departmentIDs with valid IDs
			departmentIDs = make([]string, 0)
			for _, dept := range depts {
				departmentIDs = append(departmentIDs, dept.ID)
			}
		}(departmentIDs)
	}

	// Validate users
	if len(assignIDs) > 0 {
		wg.Add(1)
		go func(aIDs []string) {
			defer wg.Done()
			u, err := uc.shopUC.ListAllUsers(ctx, sc, microservice.GetUsersFilter{
				BranchIds: branchIDs,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.validateAssign.GetUsers: %v", err)
				wgErr = err
				return
			}

			uIDs := make([]string, 0)
			for _, u := range u {
				uIDs = append(uIDs, u.ID)
			}

			for _, assignID := range aIDs {
				if !util.Contains(uIDs, assignID) {
					uc.l.Errorf(ctx, "event.usecase.validateAssign.GetUsers: %v", event.ErrAssignNotBelongToBranch)
					wgErr = event.ErrAssignNotBelongToBranch
					return
				}
			}
		}(assignIDs)
	}

	// Validate departments belong to branches
	if len(departmentIDs) > 0 && len(branchIDs) > 0 {
		wg.Add(1)
		go func(deptIDs []string) {
			defer wg.Done()
			deptsByBranch, err := uc.shopUC.GetDepartments(ctx, sc, microservice.GetDepartmentsFilter{
				BranchIds: branchIDs,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.validateAssign.GetDepartments: %v", err)
				wgErr = err
				return
			}

			depts, err := uc.shopUC.GetDepartments(ctx, sc, microservice.GetDepartmentsFilter{
				IDs: deptIDs,
			})
			if err != nil {
				uc.l.Errorf(ctx, "event.usecase.validateAssign.GetDepartments: %v", err)
				wgErr = err
				return
			}

			deptIDsEmptyBranch := make([]string, 0)
			for _, dept := range depts {
				if len(dept.BranchIDs) == 0 {
					deptIDsEmptyBranch = append(deptIDsEmptyBranch, dept.ID)
				}
			}

			deptIDsByBranch := make([]string, 0)
			for _, dept := range deptsByBranch {
				deptIDsByBranch = append(deptIDsByBranch, dept.ID)
			}

			for _, deptID := range departmentIDs {
				if !util.Contains(deptIDsByBranch, deptID) && !util.Contains(deptIDsEmptyBranch, deptID) {
					uc.l.Errorf(ctx, "event.usecase.validateAssign.GetDepartments: %v", event.ErrDepartmentNotBelongToBranch)
					wgErr = event.ErrDepartmentNotBelongToBranch
					return
				}
			}
		}(departmentIDs)
	}

	wg.Wait()

	if wgErr != nil {
		return wgErr
	}

	return nil
}

// calculateNotifyTimeForEvent calculates the notification time for an event
func (uc implUseCase) calculateNotifyTimeForEvent(
	startTime time.Time,
	allDay bool,
	alert *models.DateConfig,
	eventTimezoneOffsetSeconds int,
) *time.Time {
	systemTimezoneOffsetSeconds := UTC_PLUS_7_OFFSET
	var notifyTimeUTC time.Time

	if allDay {
		notifyTimeUTC = uc.calculateAllDayEventNotifyTime(startTime, alert, eventTimezoneOffsetSeconds)
	} else {
		notifyTimeUTC = uc.calculateRegularEventNotifyTime(startTime, alert, eventTimezoneOffsetSeconds)
	}

	if notifyTimeUTC.IsZero() {
		return nil
	}

	systemLoc := time.FixedZone("SystemTimezone", systemTimezoneOffsetSeconds)
	notifyTime := notifyTimeUTC.In(systemLoc)
	return &notifyTime
}

// calculateRegularEventNotifyTime calculates notification time for regular events
func (uc implUseCase) calculateRegularEventNotifyTime(eventTime time.Time, eventAlert *models.DateConfig, timezoneOffsetSeconds int) time.Time {
	// Check if alert is configured
	if eventAlert == nil || (eventAlert.Num == 0 && eventAlert.Unit == "" && !eventAlert.Instant) {
		return time.Time{}
	}

	loc := time.FixedZone("EventTimezone", timezoneOffsetSeconds)
	eventTimeInTZ := eventTime.In(loc)
	notifyTimeInTZ := eventTimeInTZ.Add(-uc.getDuration(*eventAlert))

	return notifyTimeInTZ.UTC()
}

// calculateAllDayEventNotifyTime calculates notification time for all-day events
func (uc implUseCase) calculateAllDayEventNotifyTime(eventDate time.Time, allDayEventAlert *models.DateConfig, timezoneOffsetSeconds int) time.Time {
	// Check if alert is configured
	if allDayEventAlert == nil || (allDayEventAlert.Hour == 0 && allDayEventAlert.Num == 0 && !allDayEventAlert.Instant) {
		return time.Time{}
	}

	loc := time.FixedZone("EventTimezone", timezoneOffsetSeconds)
	eventDateInTZ := eventDate.In(loc)

	// Get the start of the event day
	eventDay := time.Date(eventDateInTZ.Year(), eventDateInTZ.Month(), eventDateInTZ.Day(), 0, 0, 0, 0, loc)

	// Calculate days before based on the DateConfig
	var daysBefore int
	switch allDayEventAlert.Unit {
	case models.DateUnitDay:
		daysBefore = allDayEventAlert.Num
	case models.DateUnitWeek:
		daysBefore = allDayEventAlert.Num * 7
	default:
		daysBefore = 0
	}

	notifyDay := eventDay.AddDate(0, 0, -daysBefore)

	// Use the Hour from the DateConfig and 0 for minutes
	notifyTime := time.Date(
		notifyDay.Year(),
		notifyDay.Month(),
		notifyDay.Day(),
		allDayEventAlert.Hour,
		0, // Always use 0 for minutes
		0, 0, loc)

	return notifyTime.UTC()
}

// getDuration returns the duration based on Num and Unit
func (uc implUseCase) getDuration(config models.DateConfig) time.Duration {
	// If instant notification, return 0 duration
	if config.Instant {
		return 0
	}

	switch config.Unit {
	case models.DateUnitMinute:
		return time.Duration(config.Num) * time.Minute
	case models.DateUnitHour:
		return time.Duration(config.Num) * time.Hour
	case models.DateUnitDay:
		return time.Duration(config.Num) * 24 * time.Hour
	case models.DateUnitWeek:
		return time.Duration(config.Num) * 7 * 24 * time.Hour
	default:
		return 0
	}
}

func (uc implUseCase) getNextOccurrence(start time.Time, end time.Time, repeat models.EventRepeat) (time.Time, time.Time) {
	switch repeat {
	case models.EventRepeatDaily:
		nextStart := start.AddDate(0, 0, 1)
		nextEnd := end.AddDate(0, 0, 1)
		return nextStart, nextEnd
	case models.EventRepeatWeekly:
		nextStart := start.AddDate(0, 0, 7)
		nextEnd := end.AddDate(0, 0, 7)
		return nextStart, nextEnd
	case models.EventRepeatMonthly:
		// Calculate the duration between start and end
		duration := end.Sub(start)
		originalDay := start.Day()

		// Start with next month
		nextStart := start.AddDate(0, 1, 0)

		// Find the next valid month
		attempts := 0
		for attempts < MaxMonthlyOccurrenceAttempts {
			year, month, _ := nextStart.Date()
			lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, start.Location()).Day()

			if lastDay >= originalDay {
				// This month has enough days
				nextStart = time.Date(year, month, originalDay,
					start.Hour(), start.Minute(), start.Second(),
					start.Nanosecond(), start.Location())
				break
			}

			// Try the next month
			nextStart = time.Date(year, month+1, 1,
				start.Hour(), start.Minute(), start.Second(),
				start.Nanosecond(), start.Location())
			attempts++
		}

		if attempts >= MaxMonthlyOccurrenceAttempts {
			uc.l.Errorf(context.Background(), "getNextOccurrence: monthly - exceeded max attempts, using fallback")
			nextStart = start.AddDate(0, 1, 0)
		}

		nextEnd := nextStart.Add(duration)
		return nextStart, nextEnd
	case models.EventRepeatYearly:
		// For yearly events, we need to handle days that don't exist in certain months
		originalDay := start.Day()
		originalMonth := start.Month()

		// Calculate next year
		nextStart := start.AddDate(1, 0, 0)

		// Check if the day exists in the target month
		year := nextStart.Year()
		lastDay := time.Date(year, originalMonth+1, 0, 0, 0, 0, 0, start.Location()).Day()

		if originalDay > lastDay {
			uc.l.Warnf(context.Background(), "getNextOccurrence: yearly - day %d doesn't exist in %s %d, skipping to next year", originalDay, originalMonth, year)
			// If the day doesn't exist in the target month, skip to the next year
			nextStart = time.Date(year+1, originalMonth, 1,
				start.Hour(), start.Minute(), start.Second(),
				start.Nanosecond(), start.Location())

			// Recursively find the next valid occurrence
			nextStart, nextEnd := uc.getNextOccurrence(nextStart, nextStart.Add(end.Sub(start)), repeat)
			return nextStart, nextEnd
		}

		// Set the correct day in the target month
		nextStart = time.Date(year, originalMonth, originalDay,
			start.Hour(), start.Minute(), start.Second(),
			start.Nanosecond(), start.Location())

		// Calculate the duration between original start and end
		duration := end.Sub(start)
		nextEnd := nextStart.Add(duration)

		return nextStart, nextEnd
	default:
		uc.l.Warnf(context.Background(), "getNextOccurrence: unknown repeat type %s, returning original times", repeat)
		return start, end
	}
}

func (uc implUseCase) getPreviousOccurrence(start time.Time, end time.Time, repeat models.EventRepeat) (time.Time, time.Time) {
	switch repeat {
	case models.EventRepeatDaily:
		return start.AddDate(0, 0, -1), end.AddDate(0, 0, -1)
	case models.EventRepeatWeekly:
		return start.AddDate(0, 0, -7), end.AddDate(0, 0, -7)
	case models.EventRepeatMonthly:
		// Calculate the duration between start and end
		duration := end.Sub(start)

		// Get the original day of month
		originalDay := start.Day()

		// Get the previous month's date
		prevStart := start.AddDate(0, -1, 0)

		// Keep subtracting months until we find one with enough days
		for {
			year, month, _ := prevStart.Date()
			lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, start.Location()).Day()

			// If this month has enough days for our original date
			if lastDay >= originalDay {
				prevStart = time.Date(year, month, originalDay,
					start.Hour(), start.Minute(), start.Second(),
					start.Nanosecond(), start.Location())
				break
			}

			// Try previous month
			prevStart = prevStart.AddDate(0, -1, 0)
		}

		// Calculate the new end time by adding the original duration
		prevEnd := prevStart.Add(duration)
		return prevStart, prevEnd
	case models.EventRepeatYearly:
		// For yearly events, we need to handle days that don't exist in certain months
		originalDay := start.Day()
		originalMonth := start.Month()

		// Calculate previous year
		prevStart := start.AddDate(-1, 0, 0)

		// Check if the day exists in the target month
		year := prevStart.Year()
		lastDay := time.Date(year, originalMonth+1, 0, 0, 0, 0, 0, start.Location()).Day()

		if originalDay > lastDay {
			// If the day doesn't exist in the target month, skip to the previous year
			prevStart = time.Date(year-1, originalMonth, 1,
				start.Hour(), start.Minute(), start.Second(),
				start.Nanosecond(), start.Location())

			// Recursively find the previous valid occurrence
			prevStart, prevEnd := uc.getPreviousOccurrence(prevStart, prevStart.Add(end.Sub(start)), repeat)
			return prevStart, prevEnd
		}

		// Set the correct day in the target month
		prevStart = time.Date(year, originalMonth, originalDay,
			start.Hour(), start.Minute(), start.Second(),
			start.Nanosecond(), start.Location())

		// Calculate the duration between original start and end
		duration := end.Sub(start)
		prevEnd := prevStart.Add(duration)

		return prevStart, prevEnd
	default:
		return start, end
	}
}

func (uc implUseCase) generateRecurringInstances(ctx context.Context, sc models.Scope, e models.Event, fromTime time.Time, tz *models.Element) ([]models.RecurringInstance, error) {
	if e.Repeat == models.EventRepeatNone {
		return nil, nil
	}

	if fromTime.Before(e.StartTime) {
		return nil, nil
	}

	currentYear, currentMonth, _ := fromTime.Date()
	monthEnd := time.Date(currentYear, currentMonth+1, 0, 23, 59, 59, 999999999, fromTime.Location())

	toTime := monthEnd
	if e.Repeat == models.EventRepeatYearly {
		toTime = time.Date(currentYear, 12, 31, 23, 59, 59, 0, fromTime.Location())
	}

	finalTime := toTime
	if e.RepeatUntil != nil && e.RepeatUntil.Before(toTime) {
		finalTime = *e.RepeatUntil
	}

	// Find the next occurrence from the original event that falls after fromTime
	currentStartTime := e.StartTime
	currentEndTime := e.EndTime

	// Move forward until we reach or pass fromTime
	for currentStartTime.Before(fromTime) {
		currentStartTime, currentEndTime = uc.getNextOccurrence(currentStartTime, currentEndTime, e.Repeat)
	}

	// Calculate the duration between original start and end times
	duration := e.EndTime.Sub(e.StartTime)
	currentEndTime = currentStartTime.Add(duration)

	instances := make([]models.RecurringInstance, 0)
	opts := make([]repository.CreateRecurringInstanceOptions, 0)
	trackingByMonth := make(map[string][]repository.StartEndTime)

	instanceCount := 0
	for currentStartTime.Before(finalTime) || currentStartTime.Equal(finalTime) {
		startTime, endTime := currentStartTime, currentEndTime

		// Only generate instances that start within the target month
		if startTime.Year() != currentYear || startTime.Month() != currentMonth {
			break
		}

		opt := repository.CreateRecurringInstanceOptions{
			EventID:       e.ID.Hex(),
			Title:         e.Title,
			BranchIDs:     mongo.HexFromObjectIDsOrNil(e.BranchIDs),
			AssignIDs:     e.AssignIDs,
			DepartmentIDs: mongo.HexFromObjectIDsOrNil(e.DepartmentIDs),
			TimezoneID:    e.TimezoneID.Hex(),
			StartTime:     startTime,
			EndTime:       endTime,
			AllDay:        e.AllDay,
			Repeat:        e.Repeat,
			RoomIDs:       mongo.HexFromObjectIDsOrNil(e.RoomIDs),
			Description:   e.Description,
			RepeatUntil:   e.RepeatUntil,
			Notify:        e.Notify,
			System:        e.System,
			Alert:         e.Alert,
			Public:        e.Public,
		}

		if e.CategoryID != nil {
			opt.CategoryID = e.CategoryID.Hex()
		}

		if e.Notify && e.Alert != nil {
			opt.NotifyTime = uc.calculateNotifyTimeForEvent(startTime, e.AllDay, e.Alert, *tz.Offset)
		}

		opts = append(opts, opt)

		// Create tracking record for the month this instance belongs to
		instanceYear := startTime.Year()
		instanceMonth := startTime.Month()
		monthKey := fmt.Sprintf("%d-%02d", instanceYear, int(instanceMonth))
		trackingByMonth[monthKey] = append(trackingByMonth[monthKey], repository.StartEndTime{
			StartTime: startTime,
			EndTime:   endTime,
		})

		currentStartTime, currentEndTime = uc.getNextOccurrence(currentStartTime, currentEndTime, e.Repeat)
		instanceCount++

		if e.RepeatUntil != nil && currentStartTime.After(*e.RepeatUntil) {
			break
		}
	}

	if len(opts) > 0 {
		ris, err := uc.repo.CreateManyRecurringInstances(ctx, sc, repository.CreateManyRecurringInstancesOptions{
			EventID:            e.ID.Hex(),
			RecurringInstances: opts,
		})
		if err != nil {
			uc.l.Errorf(ctx, "generateRecurringInstances: Error creating instances: %v", err)
			return nil, err
		}
		instances = append(instances, ris...)
	}

	// Create tracking records for all months that have instances
	for monthKey, startEndTimes := range trackingByMonth {
		var year, month int
		fmt.Sscanf(monthKey, "%d-%d", &year, &month)

		_, err := uc.repo.CreateRecurringTracking(ctx, sc, repository.CreateRecurringTrackingOptions{
			EventID:      e.ID.Hex(),
			Month:        int32(month),
			Year:         int32(year),
			Repeat:       e.Repeat,
			RepeatUntil:  e.RepeatUntil,
			StartEndTime: startEndTimes,
		})
		if err != nil {
			uc.l.Errorf(ctx, "generateRecurringInstances: Error creating tracking record for month %s: %v", monthKey, err)
			return nil, err
		}
	}

	return instances, nil
}

func (uc implUseCase) getEventNoti(ctx context.Context, sc models.Scope, input getEventNotiInput) (notification.Notification, error) {
	shopScope := notification.ShopScope{
		ID:     sc.ShopID,
		Suffix: sc.Suffix,
	}

	langs := locale.SupportedLanguages
	headings := make(map[string]string)
	contents := make(map[string]string)

	for lang, ok := range langs {
		if !ok {
			continue
		}

		localizedTimeText := input.TimeText
		if input.Type == notification.SourceEventReminder {
			// For reminder notifications, format duration per language (e.g., "in 30 minutes")
			localizedTimeText = notification.FormatDurationI18n(ctx, notification.FormatDurationInput{
				Duration: input.Duration,
				Lang:     lang,
			})
		}

		heading, err := notification.GetNotiHeading(ctx, notification.GetNotiHeadingInput{
			Lang: lang,
			From: notification.SourceEvent,
		})
		if err != nil {
			return notification.Notification{}, err
		}
		headings[lang] = heading

		content, err := notification.GetNotiContent(ctx, notification.GetNotiContentInput{
			Lang:         lang,
			From:         input.Type,
			TimeText:     localizedTimeText,
			CreatedName:  input.CreatedName,
			AssigneeName: input.AssigneeName,
			EventTitle:   input.EI.Title,
			EventDate:    input.DateText,
			OldTimeText:  input.OldTimeText,
			OldDateText:  input.OldDateText,
			DeleteType:   input.DeleteType,
		})
		if err != nil {
			return notification.Notification{}, err
		}
		contents[lang] = content
	}
	n := notification.Notification{
		ShopScope:     shopScope,
		Content:       contents[locale.ViLanguage],
		Heading:       headings[locale.ViLanguage],
		UserIDs:       input.UserIDs,
		CreatedUserID: sc.UserID,
		Data: notification.NotiData{
			Data: event.PublishNotiEventInput{
				ID:          input.EI.ID.Hex(),
				EventID:     input.EI.EventID.Hex(),
				CreatedByID: input.EI.CreatedByID,
				StartTime:   util.DateTimeToStr(input.EI.StartTime, nil),
			},
			Activity: notification.ActivityEventDetail,
		},
		Source: notification.SourceEvent,
	}
	if _, ok := langs[locale.EnLanguage]; ok {
		n.En = notification.MultiLangObj{
			Heading: headings[locale.EnLanguage],
			Content: contents[locale.EnLanguage],
		}
	}

	if _, ok := langs[locale.JaLanguage]; ok {
		n.Ja = notification.MultiLangObj{
			Heading: headings[locale.JaLanguage],
			Content: contents[locale.JaLanguage],
		}
	}

	return n, nil
}

// MonthYear represents a specific month and year combination
type MonthYear struct {
	Month int
	Year  int
}

// getRecurringInstanceInDateRange is the main function that coordinates the generation and retrieval of event instances
func (uc implUseCase) getRecurringInstanceInDateRange(ctx context.Context, sc models.Scope, fromTime, toTime time.Time) ([]models.RecurringInstance, error) {
	// 1. Get all trackings first
	generatedTrackingsInDateRange, err := uc.repo.GetGenRTsInDateRange(ctx, sc, fromTime, toTime)
	if err != nil {
		uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error getting trackings in date range: %v", err)
		return nil, err
	}

	generatedTrackingsNotInDateRange, err := uc.repo.GetGenRTsNotInDateRange(ctx, sc, fromTime, toTime)
	if err != nil {
		uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error getting trackings not in date range: %v", err)
		return nil, err
	}

	// 2. Build a comprehensive map of all generated months by event
	generatedMonthsByEvent := make(map[string]map[string]bool)

	// Add all months from both in-range and out-of-range trackings
	allTrackings := append(generatedTrackingsInDateRange, generatedTrackingsNotInDateRange...)
	for _, rt := range allTrackings {
		eventID := rt.EventID.Hex()
		if generatedMonthsByEvent[eventID] == nil {
			generatedMonthsByEvent[eventID] = make(map[string]bool)
		}
		monthKey := fmt.Sprintf("%d-%02d", rt.Year, rt.Month)
		generatedMonthsByEvent[eventID][monthKey] = true
	}

	// 3. Get unique event IDs that need processing
	eventIDs := make(map[string]bool)

	// Get all unique event IDs from both tracking ranges
	for _, rt := range allTrackings {
		eventIDs[rt.EventID.Hex()] = true
	}

	// 4. For each event, determine which months need to be generated
	monthsToGenerate := make(map[string][]MonthYear)
	totalMonthsToGenerate := 0

	for eventID := range eventIDs {
		current := fromTime
		eventMonths := 0
		for current.Before(toTime) {
			my := MonthYear{
				Month: int(current.Month()),
				Year:  current.Year(),
			}

			// Only add if month hasn't been generated
			monthKey := fmt.Sprintf("%d-%02d", my.Year, my.Month)
			if generated := generatedMonthsByEvent[eventID]; generated == nil || !generated[monthKey] {
				monthsToGenerate[eventID] = append(monthsToGenerate[eventID], my)
				eventMonths++
				totalMonthsToGenerate++
			}

			// Move to first day of next month
			current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
		}
	}

	// If no months need to be generated, return existing instances
	if len(monthsToGenerate) == 0 {
		existingInstances, err := uc.getExistingInstances(ctx, sc, generatedTrackingsInDateRange, fromTime, toTime)
		if err != nil {
			uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error getting existing instances: %v", err)
			return nil, err
		}
		return existingInstances, nil
	}

	// 5. Get all events that need processing
	allEventIDs := make([]string, 0, len(monthsToGenerate))
	for eventID := range monthsToGenerate {
		allEventIDs = append(allEventIDs, eventID)
	}

	events, err := uc.repo.ListByIDs(ctx, sc, allEventIDs)
	if err != nil {
		uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error listing events by IDs: %v", err)
		return nil, err
	}

	eventMap := make(map[string]models.Event)
	for _, e := range events {
		eventMap[e.ID.Hex()] = e
	}

	// 6. Get all timezones
	tzIDs := make([]string, 0, len(events))
	for _, e := range events {
		tzIDs = append(tzIDs, e.TimezoneID.Hex())
	}

	tzs, err := uc.elementUC.List(ctx, sc, element.ListInput{
		Filter: element.Filter{
			IDs: tzIDs,
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error listing timezones: %v", err)
		return nil, err
	}

	tzMap := make(map[string]models.Element)
	for _, tz := range tzs {
		tzMap[tz.ID.Hex()] = tz
	}

	// 7. Generate new instances only for months that haven't been generated
	var allInstances []models.RecurringInstance
	processedEvents := 0

	// Process each event sequentially to avoid race conditions
	for eventID, months := range monthsToGenerate {
		ev, exists := eventMap[eventID]
		if !exists {
			uc.l.Warnf(ctx, "getRecurringInstanceInDateRange: Event %s not found in event map, skipping", eventID)
			continue
		}

		// Filter out months that are before the event start date
		var validMonths []MonthYear
		for _, my := range months {
			monthStart := time.Date(my.Year, time.Month(my.Month), 1, 0, 0, 0, 0, time.UTC)
			if !monthStart.Before(ev.StartTime) {
				validMonths = append(validMonths, my)
			}
		}

		if len(validMonths) == 0 {
			uc.l.Warnf(ctx, "getRecurringInstanceInDateRange: Event %s has no valid months (all before start date), skipping", eventID)
			continue
		}

		months = validMonths
		tz := tzMap[ev.TimezoneID.Hex()]

		// Batch all instances for this event
		allOpts := make([]repository.CreateRecurringInstanceOptions, 0)
		allTrackings := make(map[string][]repository.StartEndTime)

		// For yearly events, only process the month that matches the original event's month
		if ev.Repeat == models.EventRepeatYearly {
			originalMonth := ev.StartTime.Month()

			// Filter months to only include those matching the original month
			var relevantMonths []MonthYear
			for _, monthYear := range months {
				if time.Month(monthYear.Month) == originalMonth {
					relevantMonths = append(relevantMonths, monthYear)
				}
			}

			if len(relevantMonths) == 0 {
				uc.l.Warnf(ctx, "getRecurringInstanceInDateRange: Event %s yearly event has no relevant months, skipping", eventID)
				continue
			}

			months = relevantMonths
		}

		// For monthly and yearly events, we need to be more careful about which months to process
		if ev.Repeat == models.EventRepeatMonthly || ev.Repeat == models.EventRepeatYearly {
			// Get the original day of the month from the event
			originalDay := ev.StartTime.Day()

			monthsProcessed := 0
			// Process only months that have the required day
			for _, monthYear := range months {
				// Check if this month has the required day
				daysInMonth := time.Date(monthYear.Year, time.Month(monthYear.Month+1), 0, 0, 0, 0, 0, time.UTC).Day()
				monthKey := fmt.Sprintf("%d-%02d", monthYear.Year, monthYear.Month)

				// Double check if month is already generated
				if generated := generatedMonthsByEvent[eventID]; generated != nil && generated[monthKey] {
					continue
				}

				// Skip months that don't have the required day
				if originalDay > daysInMonth {
					continue
				}

				startTime := uc.monthYearToTime(monthYear, time.UTC)
				monthEnd := time.Date(startTime.Year(), startTime.Month()+1, 0, 23, 59, 59, 0, startTime.Location())

				instances, tracking, err := uc.generateInstancesForMonth(ctx, sc, ev, startTime, monthEnd, &tz)
				if err != nil {
					uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error generating instances for month %s: %v", monthKey, err)
					return nil, err
				}

				if len(instances) > 0 {
					allOpts = append(allOpts, instances...)
					allTrackings[monthKey] = tracking
					monthsProcessed++
				}
			}
		} else {
			// For daily and weekly events, process all months
			monthsProcessed := 0

			for _, monthYear := range months {
				monthKey := fmt.Sprintf("%d-%02d", monthYear.Year, monthYear.Month)

				// Double check if month is already generated
				if generated := generatedMonthsByEvent[eventID]; generated != nil && generated[monthKey] {
					continue
				}

				startTime := uc.monthYearToTime(monthYear, time.UTC)
				monthEnd := time.Date(startTime.Year(), startTime.Month()+1, 0, 23, 59, 59, 0, startTime.Location())

				instances, tracking, err := uc.generateInstancesForMonth(ctx, sc, ev, startTime, monthEnd, &tz)
				if err != nil {
					uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error generating instances for month %s: %v", monthKey, err)
					return nil, err
				}

				if len(instances) > 0 {
					allOpts = append(allOpts, instances...)
					allTrackings[monthKey] = tracking
					monthsProcessed++
				}
			}
		}

		// Create all instances for this event in a single transaction
		if len(allOpts) > 0 {
			// Create tracking records first
			trackingCreated := 0
			for monthKey, startEndTimes := range allTrackings {
				var year, month int
				fmt.Sscanf(monthKey, "%d-%d", &year, &month)

				// Double check if month is already generated before creating tracking record
				if generated := generatedMonthsByEvent[eventID]; generated != nil && generated[monthKey] {
					continue
				}

				// Create tracking record with upsert semantics
				_, err := uc.repo.CreateRecurringTracking(ctx, sc, repository.CreateRecurringTrackingOptions{
					EventID:      eventID,
					Month:        int32(month),
					Year:         int32(year),
					Repeat:       ev.Repeat,
					RepeatUntil:  ev.RepeatUntil,
					StartEndTime: startEndTimes,
				})
				if err != nil {
					uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error creating tracking record for month %s: %v", monthKey, err)
					return nil, err
				}
				trackingCreated++
			}

			// Create all instances after tracking records are created
			ris, err := uc.repo.CreateManyRecurringInstances(ctx, sc, repository.CreateManyRecurringInstancesOptions{
				EventID:            eventID,
				RecurringInstances: allOpts,
			})
			if err != nil {
				uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error creating recurring instances: %v", err)
				return nil, err
			}

			// Convert to EventInstance
			allInstances = append(allInstances, ris...)
		}

		processedEvents++
	}

	// 8. Get existing instances
	existingInstances, err := uc.getExistingInstances(ctx, sc, generatedTrackingsInDateRange, fromTime, toTime)
	if err != nil {
		uc.l.Errorf(ctx, "getRecurringInstanceInDateRange: Error getting existing instances: %v", err)
		return nil, err
	}

	// 9. Combine and return all instances
	result := append(existingInstances, allInstances...)
	return uc.deduplicateRecurringInstances(result), nil
}

// deduplicateRecurringInstances removes duplicate RecurringInstance entries based on their ID
func (uc implUseCase) deduplicateRecurringInstances(instances []models.RecurringInstance) []models.RecurringInstance {
	seen := make(map[string]bool)
	result := make([]models.RecurringInstance, 0)

	for _, instance := range instances {
		id := instance.ID.Hex()
		if !seen[id] {
			seen[id] = true
			result = append(result, instance)
		}
	}

	return result
}

// Helper function to generate instances for a single month
func (uc implUseCase) generateInstancesForMonth(ctx context.Context, sc models.Scope, e models.Event, startTime, endTime time.Time, tz *models.Element) ([]repository.CreateRecurringInstanceOptions, []repository.StartEndTime, error) {
	var opts []repository.CreateRecurringInstanceOptions
	var tracking []repository.StartEndTime

	// Get existing tracking for this month
	existingTracking, err := uc.repo.GetGenRTsInDateRange(ctx, sc, startTime, endTime)
	if err != nil {
		uc.l.Errorf(ctx, "generateInstancesForMonth: Error getting existing tracking: %v", err)
		return nil, nil, err
	}

	// If we have tracking for this month, skip it
	for _, rt := range existingTracking {
		if rt.EventID.Hex() == e.ID.Hex() && int32(startTime.Month()) == rt.Month && int32(startTime.Year()) == rt.Year {
			return nil, nil, nil
		}
	}

	currentTime, err := uc.getInitialStartTimeForMonth(e, startTime)
	if err != nil {
		uc.l.Errorf(ctx, "generateInstancesForMonth: Error getting initial start time: %v", err)
		return nil, nil, err
	}

	if currentTime.IsZero() {
		return nil, nil, nil
	}

	// Calculate the duration between original start and end times
	duration := e.EndTime.Sub(e.StartTime)

	instanceCount := 0

	for currentTime.Before(endTime) && instanceCount < MaxInstancesPerMonth {
		// For daily events, we only need to check if we're within the generation period
		// For other types, we need to check against the original event's start time
		if e.Repeat != models.EventRepeatDaily && currentTime.Before(e.StartTime) {
			currentEndTime := currentTime.Add(duration)
			currentTime, _ = uc.getNextOccurrence(currentTime, currentEndTime, e.Repeat)
			continue
		}

		// Skip if after repeat until
		if e.RepeatUntil != nil && currentTime.After(*e.RepeatUntil) {
			break
		}

		// Check if current time is still in the month we're generating for
		if currentTime.Month() != startTime.Month() || currentTime.Year() != startTime.Year() {
			break
		}

		instanceEndTime := currentTime.Add(duration)
		if instanceEndTime.After(endTime) {
			instanceEndTime = endTime
		}

		opt := repository.CreateRecurringInstanceOptions{
			EventID:       e.ID.Hex(),
			Title:         e.Title,
			AssignIDs:     e.AssignIDs,
			DepartmentIDs: mongo.HexFromObjectIDsOrNil(e.DepartmentIDs),
			BranchIDs:     mongo.HexFromObjectIDsOrNil(e.BranchIDs),
			TimezoneID:    e.TimezoneID.Hex(),
			StartTime:     currentTime,
			EndTime:       instanceEndTime,
			AllDay:        e.AllDay,
			Repeat:        e.Repeat,
			RoomIDs:       mongo.HexFromObjectIDsOrNil(e.RoomIDs),
			Description:   e.Description,
			RepeatUntil:   e.RepeatUntil,
			Notify:        e.Notify,
			System:        e.System,
			Alert:         e.Alert,
			Public:        e.Public,
		}

		if e.CategoryID != nil {
			opt.CategoryID = e.CategoryID.Hex()
		}

		if e.Notify && e.Alert != nil {
			opt.NotifyTime = uc.calculateNotifyTimeForEvent(currentTime, e.AllDay, e.Alert, *tz.Offset)
		}

		opts = append(opts, opt)
		tracking = append(tracking, repository.StartEndTime{
			StartTime: currentTime,
			EndTime:   instanceEndTime,
		})

		instanceCount++

		// Get next occurrence
		prevTime := currentTime
		currentTime, _ = uc.getNextOccurrence(currentTime, instanceEndTime, e.Repeat)

		// Safety check for infinite loops
		if currentTime.Equal(prevTime) {
			uc.l.Errorf(ctx, "generateInstancesForMonth: Event %s detected infinite loop - currentTime not advancing from %v", e.ID.Hex(), currentTime)
			break
		}
	}

	if instanceCount >= MaxInstancesPerMonth {
		uc.l.Warnf(ctx, "generateInstancesForMonth: Event %s hit maximum instance limit (%d), stopping generation", e.ID.Hex(), MaxInstancesPerMonth)
	}

	return opts, tracking, nil
}

func (uc implUseCase) getInitialStartTimeForMonth(e models.Event, startTime time.Time) (time.Time, error) {
	if e.Repeat == models.EventRepeatYearly {
		// For yearly events, use the original month and day
		result := time.Date(startTime.Year(), e.StartTime.Month(), e.StartTime.Day(),
			e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(),
			e.StartTime.Nanosecond(), startTime.Location())
		return result, nil
	}

	if e.Repeat == models.EventRepeatMonthly {
		// For monthly events, use the original day but current month
		originalDay := e.StartTime.Day()
		daysInMonth := time.Date(startTime.Year(), startTime.Month()+1, 0, 0, 0, 0, 0, startTime.Location()).Day()

		// If the original day doesn't exist in this month, skip the month
		if originalDay > daysInMonth {
			return time.Time{}, nil
		}

		result := time.Date(startTime.Year(), startTime.Month(), originalDay,
			e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(),
			e.StartTime.Nanosecond(), startTime.Location())
		return result, nil
	}

	if e.Repeat == models.EventRepeatWeekly {
		// For weekly events, start from the first day of the month
		// and adjust to the next occurrence of the same weekday
		firstDayOfMonth := time.Date(startTime.Year(), startTime.Month(), 1,
			e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(),
			e.StartTime.Nanosecond(), startTime.Location())

		// Calculate days to add to reach the same weekday as the original event
		originalWeekday := e.StartTime.Weekday()
		firstDayWeekday := firstDayOfMonth.Weekday()
		daysToAdd := int(originalWeekday - firstDayWeekday)
		if daysToAdd < 0 {
			daysToAdd += 7
		}

		result := firstDayOfMonth.AddDate(0, 0, daysToAdd)
		return result, nil
	}

	// For daily events, we need to handle the transition between months carefully
	// to avoid duplicate events on the first day of the month

	lastDayOfPrevMonth := time.Date(startTime.Year(), startTime.Month(), 0, 0, 0, 0, 0, startTime.Location())
	lastEventTime := time.Date(lastDayOfPrevMonth.Year(), lastDayOfPrevMonth.Month(), lastDayOfPrevMonth.Day(),
		e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(),
		e.StartTime.Nanosecond(), startTime.Location())

	// Calculate the time difference between the last event and the start of the current month
	timeDiff := startTime.Sub(lastEventTime)

	// Additional validation for edge cases
	if timeDiff < 0 {
		uc.l.Warnf(context.Background(), "getInitialStartTimeForMonth: Event %s daily - negative time difference detected (%v), this may indicate a timezone or calculation issue", e.ID.Hex(), timeDiff)
	}

	// Check if event started before the current month being generated
	monthStart := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, startTime.Location())
	if e.StartTime.Before(monthStart) {
		// Event started before this month, safe to start from day 1
		result := time.Date(startTime.Year(), startTime.Month(), 1,
			e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(),
			e.StartTime.Nanosecond(), startTime.Location())
		return result, nil
	}

	// If the last event of the previous month would occur after the start time of this month,
	// or if the time difference is less than 24 hours, we should start from day 2
	if lastEventTime.After(startTime) || timeDiff < 24*time.Hour {
		result := time.Date(startTime.Year(), startTime.Month(), 2,
			e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(),
			e.StartTime.Nanosecond(), startTime.Location())
		return result, nil
	}

	// Otherwise, start from the first day of the month
	result := time.Date(startTime.Year(), startTime.Month(), 1,
		e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(),
		e.StartTime.Nanosecond(), startTime.Location())
	return result, nil
}

// monthYearToTime converts MonthYear to time.Time
func (uc implUseCase) monthYearToTime(my MonthYear, loc *time.Location) time.Time {
	return time.Date(my.Year, time.Month(my.Month), 1, 0, 0, 0, 0, loc)
}

func (uc implUseCase) containsAny(slice []primitive.ObjectID, ids []string) bool {
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		for _, s := range slice {
			if s == oid {
				return true
			}
		}
	}
	return false
}

func (uc implUseCase) filterInstances(instances []models.RecurringInstance, sc models.Scope, departmentIDs []string, branchIDs []string, excludeCategoryIDs []string) []models.RecurringInstance {
	filtered := make([]models.RecurringInstance, 0)

	for _, instance := range instances {
		isVisible := false

		// Check branch conditions
		if len(branchIDs) > 0 {
			// Check branch condition with no department/assign
			if uc.containsAny(instance.BranchIDs, branchIDs) &&
				len(instance.DepartmentIDs) == 0 &&
				len(instance.AssignIDs) == 0 {
				isVisible = true
			}

			// Check department condition if exists
			if !isVisible && len(departmentIDs) > 0 &&
				uc.containsAny(instance.DepartmentIDs, departmentIDs) {
				isVisible = true
			}

			// Check assign condition
			if !isVisible && slices.Contains(instance.AssignIDs, sc.UserID) {
				isVisible = true
			}
		}

		// Check created by condition
		if !isVisible && instance.CreatedByID == sc.UserID {
			isVisible = true
		}

		// Check system condition
		if !isVisible && instance.System {
			isVisible = true
		}

		// Check public condition
		if !isVisible && instance.Public {
			isVisible = true
		}

		// Check category condition
		if instance.CategoryID != nil {
			if !isVisible && len(excludeCategoryIDs) > 0 && slices.Contains(excludeCategoryIDs, instance.CategoryID.Hex()) {
				isVisible = true
			}
		}

		if isVisible {
			filtered = append(filtered, instance)
		}
	}

	return filtered
}

func (uc implUseCase) getAssignUserIDs(ctx context.Context, sc models.Scope, deptIDs []string, branchIDs []string, assignIDs []string) ([]string, error) {
	uIDs := []string{}

	if len(deptIDs) > 0 && len(branchIDs) > 0 {
		us, err := uc.shopUC.ListAllUsers(ctx, sc, microservice.GetUsersFilter{
			DeptIDs: deptIDs,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Create.ListAllUsers: %v", err)
			return nil, err
		}

		for _, u := range us {
			uIDs = append(uIDs, u.ID)
		}

		if len(assignIDs) > 0 {
			uIDs = util.Intersection(uIDs, assignIDs)
		}

	} else if len(branchIDs) > 0 {
		us, err := uc.shopUC.ListAllUsers(ctx, sc, microservice.GetUsersFilter{
			BranchIds: branchIDs,
		})
		if err != nil {
			uc.l.Errorf(ctx, "event.usecase.Create.ListAllUsers: %v", err)
			return nil, err
		}

		for _, u := range us {
			uIDs = append(uIDs, u.ID)
		}

		if len(assignIDs) > 0 {
			uIDs = util.Intersection(uIDs, assignIDs)
		}
	}

	return util.RemoveDuplicates(uIDs), nil
}

func (uc implUseCase) handleUpdateEventNotifications(ctx context.Context, sc models.Scope, e event.EventInstance, updateInput event.UpdateInput, shop microservice.Shop) error {
	// Get existing user assignments
	oldAssignIDs, err := uc.getEventAssignIDs(ctx, sc, e)
	if err != nil {
		uc.l.Errorf(ctx, "handleUpdateEventNotifications: %v", err)
		return err
	}

	newAssignIDs, err := uc.getAssignUserIDs(ctx, sc, updateInput.DepartmentIDs, updateInput.BranchIDs, updateInput.AssignIDs)
	if err != nil {
		uc.l.Errorf(ctx, "handleUpdateEventNotifications: %v", err)
		return err
	}

	// Detect changes
	// Convert both times to UTC to avoid timezone comparison issues
	hasTimeChange := !e.StartTime.UTC().Truncate(time.Second).Equal(updateInput.StartTime.UTC().Truncate(time.Second))
	hasLocationChange := uc.hasRoomChanges(e.RoomIDs, updateInput.RoomIDs)
	hasParticipantChange := uc.hasParticipantChanges(oldAssignIDs, newAssignIDs)

	// Handle participant removals and additions first (these are always separate notifications)
	if hasParticipantChange {
		// Get users that were removed from the event
		removedUsers := util.Difference(oldAssignIDs, newAssignIDs)
		if len(removedUsers) > 0 {
			err := uc.handleEventRemoveNotification(ctx, sc, e, removedUsers, shop)
			if err != nil {
				return err
			}
		}

		// Get users that are newly added to the event
		addedUsers := util.Difference(newAssignIDs, oldAssignIDs)
		if len(addedUsers) > 0 {
			err := uc.handleEventAssignNotification(ctx, sc, e, addedUsers, shop)
			if err != nil {
				uc.l.Errorf(ctx, "handleUpdateEventNotifications.handleEventRemoveNotification: %v", err)
				return err
			}
		}
	}

	// For existing users, send appropriate change notifications
	// Priority: Time change > Location change > Participant list change
	existingUsers := util.Intersection(oldAssignIDs, newAssignIDs)
	existingUsers = append(existingUsers, e.CreatedByID)
	existingUsers = util.RemoveDuplicates(existingUsers)

	if len(existingUsers) > 0 {
		if hasTimeChange {
			// Time change has highest priority
			err := uc.handleEventTimeChangeNotification(ctx, sc, e, updateInput, existingUsers, shop)
			if err != nil {
				uc.l.Errorf(ctx, "handleUpdateEventNotifications.handleEventTimeChangeNotification: %v", err)
				return err
			}
		} else if hasLocationChange {
			// Location change if no time change
			err := uc.handleEventLocationChangeNotification(ctx, sc, e, existingUsers, shop)
			if err != nil {
				uc.l.Errorf(ctx, "handleUpdateEventNotifications.handleEventLocationChangeNotification: %v", err)
				return err
			}
		} else if hasParticipantChange {
			// Participant change if only participant list changed (no time/location changes)
			err := uc.handleEventParticipantChangeNotification(ctx, sc, e, existingUsers, shop)
			if err != nil {
				uc.l.Errorf(ctx, "handleUpdateEventNotifications.handleEventParticipantChangeNotification: %v", err)
				return err
			}
		} else {
			err := uc.handleEventChangeNotification(ctx, sc, e, existingUsers, shop)
			if err != nil {
				uc.l.Errorf(ctx, "handleUpdateEventNotifications.handleEventChangeNotification: %v", err)
				return err
			}
		}
	}

	return nil
}

// hasRoomChanges checks if there are changes in room assignments
func (uc implUseCase) hasRoomChanges(oldRoomIDs []primitive.ObjectID, newRoomIDs []string) bool {
	if len(oldRoomIDs) != len(newRoomIDs) {
		return true
	}

	oldRoomIDStrings := make([]string, len(oldRoomIDs))
	for i, id := range oldRoomIDs {
		oldRoomIDStrings[i] = id.Hex()
	}

	for _, newID := range newRoomIDs {
		found := false
		for _, oldID := range oldRoomIDStrings {
			if oldID == newID {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

// hasParticipantChanges checks if there are changes in participant assignments
func (uc implUseCase) hasParticipantChanges(oldAssignIDs, newAssignIDs []string) bool {
	if len(oldAssignIDs) != len(newAssignIDs) {
		return true
	}

	for _, newID := range newAssignIDs {
		found := false
		for _, oldID := range oldAssignIDs {
			if oldID == newID {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

// handleEventTimeChangeNotification handles notifications for time changes
func (uc implUseCase) handleEventTimeChangeNotification(ctx context.Context, sc models.Scope, e event.EventInstance, updateInput event.UpdateInput, userIDs []string, shop microservice.Shop) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Format old time and date
	oldTimeText, oldDateText := uc.formatEventDateTime(ctx, e, shop)

	// Format new time and date using updateInput
	newEvent := e
	newEvent.StartTime = updateInput.StartTime
	newEvent.EndTime = updateInput.EndTime
	newTimeText, newDateText := uc.formatEventDateTime(ctx, newEvent, shop)

	notiInput := getEventNotiInput{
		EI:          e,
		Type:        notification.SourceEventTimeChange,
		UserIDs:     userIDs,
		TimeText:    newTimeText,
		DateText:    newDateText,
		OldTimeText: oldTimeText,
		OldDateText: oldDateText,
	}

	noti, err := uc.getEventNoti(ctx, sc, notiInput)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventTimeChangeNotification.getEventNoti: %v", err)
		return err
	}

	err = uc.publishPushNotiMsg(ctx, noti)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventTimeChangeNotification.publishPushNotiMsg: %v", err)
		return err
	}

	return nil
}

// handleEventLocationChangeNotification handles notifications for location changes
func (uc implUseCase) handleEventLocationChangeNotification(ctx context.Context, sc models.Scope, e event.EventInstance, userIDs []string, shop microservice.Shop) error {
	if len(userIDs) == 0 {
		return nil
	}

	timeText, dateText := uc.formatEventDateTime(ctx, e, shop)

	notiInput := getEventNotiInput{
		EI:       e,
		Type:     notification.SourceEventLocationChange,
		UserIDs:  userIDs,
		TimeText: timeText,
		DateText: dateText,
	}

	noti, err := uc.getEventNoti(ctx, sc, notiInput)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventLocationChangeNotification.getEventNoti: %v", err)
		return err
	}

	err = uc.publishPushNotiMsg(ctx, noti)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventLocationChangeNotification.publishPushNotiMsg: %v", err)
		return err
	}

	return nil
}

// handleEventParticipantChangeNotification handles notifications for participant changes
func (uc implUseCase) handleEventParticipantChangeNotification(ctx context.Context, sc models.Scope, e event.EventInstance, userIDs []string, shop microservice.Shop) error {
	if len(userIDs) == 0 {
		return nil
	}

	timeText, dateText := uc.formatEventDateTime(ctx, e, shop)

	notiInput := getEventNotiInput{
		EI:       e,
		Type:     notification.SourceEventParticipantChange,
		UserIDs:  userIDs,
		TimeText: timeText,
		DateText: dateText,
	}

	noti, err := uc.getEventNoti(ctx, sc, notiInput)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventParticipantChangeNotification.getEventNoti: %v", err)
		return err
	}

	err = uc.publishPushNotiMsg(ctx, noti)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventParticipantChangeNotification.publishPushNotiMsg: %v", err)
		return err
	}

	return nil
}

func (uc implUseCase) handleEventChangeNotification(ctx context.Context, sc models.Scope, e event.EventInstance, userIDs []string, shop microservice.Shop) error {
	if len(userIDs) == 0 {
		return nil
	}

	timeText, dateText := uc.formatEventDateTime(ctx, e, shop)

	notiInput := getEventNotiInput{
		EI:       e,
		Type:     notification.SourceEventChange,
		UserIDs:  userIDs,
		TimeText: timeText,
		DateText: dateText,
	}

	noti, err := uc.getEventNoti(ctx, sc, notiInput)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventChangeNotification.getEventNoti: %v", err)
		return err
	}

	err = uc.publishPushNotiMsg(ctx, noti)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventChangeNotification.publishPushNotiMsg: %v", err)
		return err
	}

	return nil
}

func (uc implUseCase) formatEventDateTime(ctx context.Context, e event.EventInstance, shop microservice.Shop) (timeText, dateText string) {
	dateFormat := util.DateFormat
	if shop.DateFormat != "" {
		dateFormat = shop.DateFormat
	}
	timeFormat := util.TimeFormat
	if shop.TimeFormat != "" {
		timeFormat = shop.TimeFormat
	}

	location := time.FixedZone("SystemTimezone", UTC_PLUS_7_OFFSET)
	startTime := e.StartTime.In(location)

	if !e.AllDay {
		timeText = notification.FormatClock(ctx, timeFormat, startTime)
	}
	dateText = startTime.Format(dateFormat)

	return timeText, dateText
}

func (uc implUseCase) handleEventRemoveNotification(ctx context.Context, sc models.Scope, e event.EventInstance, removedUsers []string, shop microservice.Shop) error {
	if len(removedUsers) == 0 {
		return nil
	}

	timeText, dateText := uc.formatEventDateTime(ctx, e, shop)

	noti, err := uc.getEventNoti(ctx, sc, getEventNotiInput{
		EI:       e,
		Type:     notification.SourceEventRemove,
		UserIDs:  removedUsers,
		TimeText: timeText,
		DateText: dateText,
	})
	if err != nil {
		uc.l.Errorf(ctx, "handleEventRemoveNotification.getEventNoti: %v", err)
		return err
	}
	err = uc.publishPushNotiMsg(ctx, noti)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventRemoveNotification.publishPushNotiMsg: %v", err)
		return err
	}
	return nil
}

func (uc implUseCase) handleEventAssignNotification(ctx context.Context, sc models.Scope, e event.EventInstance, addedUsers []string, shop microservice.Shop) error {
	if len(addedUsers) == 0 {
		return nil
	}

	timeText, dateText := uc.formatEventDateTime(ctx, e, shop)

	noti, err := uc.getEventNoti(ctx, sc, getEventNotiInput{
		EI:       e,
		Type:     notification.SourceEventAssign,
		UserIDs:  addedUsers,
		TimeText: timeText,
		DateText: dateText,
	})
	if err != nil {
		uc.l.Errorf(ctx, "handleEventAssignNotification.getEventNoti: %v", err)
		return err
	}
	err = uc.publishPushNotiMsg(ctx, noti)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventAssignNotification.publishPushNotiMsg: %v", err)
		return err
	}
	return nil
}

func (uc implUseCase) handleEventDeleteNotification(ctx context.Context, sc models.Scope, e event.EventInstance, shop microservice.Shop, deleteType models.EventAction) error {
	assignIDs, err := uc.getEventAssignIDs(ctx, sc, e)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventDeleteNotification.getEventAssignIDs: %v", err)
		return err
	}
	assignIDs = append(assignIDs, e.CreatedByID)
	assignIDs = util.RemoveDuplicates(assignIDs)

	if len(assignIDs) == 0 {
		return nil
	}

	timeText, dateText := uc.formatEventDateTime(ctx, e, shop)

	noti, err := uc.getEventNoti(ctx, sc, getEventNotiInput{
		EI:         e,
		Type:       notification.SourceEventDelete,
		UserIDs:    assignIDs,
		TimeText:   timeText,
		DateText:   dateText,
		DeleteType: deleteType,
	})
	if err != nil {
		uc.l.Errorf(ctx, "handleEventDeleteNotification.getEventNoti: %v", err)
		return err
	}

	err = uc.publishPushNotiMsg(ctx, noti)
	if err != nil {
		uc.l.Errorf(ctx, "handleEventDeleteNotification.publishPushNotiMsg: %v", err)
		return err
	}
	return nil
}

func (uc implUseCase) getEventAssignIDs(ctx context.Context, sc models.Scope, e event.EventInstance) ([]string, error) {
	departIds := make([]string, len(e.DepartmentIDs))
	for i, id := range e.DepartmentIDs {
		departIds[i] = id.Hex()
	}

	branchIds := make([]string, len(e.BranchIDs))
	for i, id := range e.BranchIDs {
		branchIds[i] = id.Hex()
	}

	uIDs, err := uc.getAssignUserIDs(ctx, sc, departIds, branchIds, e.AssignIDs)
	if err != nil {
		uc.l.Errorf(ctx, "getEventAssignIDs.getAssignUserIDs: %v", err)
		return nil, err
	}

	return util.RemoveDuplicates(uIDs), nil
}

// getExistingInstances gets already generated instances for the date range
func (uc implUseCase) getExistingInstances(ctx context.Context, sc models.Scope, grts []models.RecurringTracking, fromTime, toTime time.Time) ([]models.RecurringInstance, error) {
	if len(grts) == 0 {
		return nil, nil
	}

	eIDs := make([]string, 0, len(grts))
	for _, e := range grts {
		eIDs = append(eIDs, e.EventID.Hex())
	}

	ris, err := uc.repo.ListRecurringInstancesByEventIDs(ctx, sc, repository.ListEventInstancesByEventIDsOptions{
		EventIDs:  eIDs,
		StartTime: fromTime,
		EndTime:   toTime,
	})
	if err != nil {
		return nil, err
	}

	return ris, nil
}
