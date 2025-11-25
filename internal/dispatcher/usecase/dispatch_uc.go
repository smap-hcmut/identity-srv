package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	"github.com/nguyentantai21042004/smap-api/internal/models"
)

func (uc implUseCase) Dispatch(ctx context.Context, req models.CrawlRequest) ([]models.CollectorTask, error) {
	if req.TaskType == "" {
		return nil, dispatcher.ErrInvalidInput
	}

	targetPlatforms, err := uc.selectPlatforms(req.Platform)
	if err != nil {
		return nil, err
	}

	tasks := make([]models.CollectorTask, 0, len(targetPlatforms))
	for _, platform := range targetPlatforms {
		task := models.CollectorTask{
			JobID:         req.JobID,
			Platform:      platform,
			TaskType:      req.TaskType,
			TimeRange:     req.TimeRange,
			Attempt:       req.Attempt,
			MaxAttempts:   req.MaxAttempts,
			SchemaVersion: uc.defaultOptions.SchemaVersion,
			EmittedAt:     req.EmittedAt,
		}

		if task.Attempt <= 0 {
			task.Attempt = 1
		}
		if task.MaxAttempts <= 0 {
			task.MaxAttempts = uc.defaultOptions.DefaultMaxAttempts
		}
		if task.EmittedAt.IsZero() {
			task.EmittedAt = time.Now().UTC()
		}

		payload, err := mapPayload(platform, req.TaskType, req.Payload)
		if err != nil {
			return nil, err
		}
		task.Payload = payload

		queue, err := uc.queueRoutingKey(platform)
		if err != nil {
			return nil, err
		}
		task.RoutingKey = queue
		task.Headers = map[string]any{
			"x-schema-version": uc.defaultOptions.SchemaVersion,
		}

		if err := uc.prod.PublishTask(ctx, task); err != nil {
			return nil, fmt.Errorf("%w: %v", dispatcher.ErrPublish, err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func mapPayload(platform models.Platform, taskType models.TaskType, raw map[string]any) (any, error) {
	if raw == nil {
		return nil, nil
	}

	switch platform {
	case models.PlatformYouTube:
		return mapYouTubePayload(taskType, raw)
	case models.PlatformTikTok:
		return mapTikTokPayload(taskType, raw)
	default:
		return nil, dispatcher.ErrUnknownRoute
	}
}

func mapYouTubePayload(taskType models.TaskType, raw map[string]any) (any, error) {
	switch taskType {
	case models.TaskTypeResearchKeyword:
		var payload models.YouTubeResearchKeywordPayload
		return decodePayload(raw, &payload)
	case models.TaskTypeCrawlLinks:
		var payload models.YouTubeCrawlLinksPayload
		return decodePayload(raw, &payload)
	case models.TaskTypeResearchAndCrawl:
		var payload models.YouTubeResearchAndCrawlPayload
		return decodePayload(raw, &payload)
	default:
		return nil, dispatcher.ErrUnknownRoute
	}
}

func mapTikTokPayload(taskType models.TaskType, raw map[string]any) (any, error) {
	switch taskType {
	case models.TaskTypeResearchKeyword:
		var payload models.TikTokResearchKeywordPayload
		return decodePayload(raw, &payload)
	case models.TaskTypeCrawlLinks:
		var payload models.TikTokCrawlLinksPayload
		return decodePayload(raw, &payload)
	case models.TaskTypeResearchAndCrawl:
		var payload models.TikTokResearchAndCrawlPayload
		return decodePayload(raw, &payload)
	default:
		return nil, dispatcher.ErrUnknownRoute
	}
}

func decodePayload(raw map[string]any, dest any) (any, error) {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, dispatcher.ErrInvalidInput
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return nil, dispatcher.ErrInvalidInput
	}
	return dest, nil
}

func (uc implUseCase) queueRoutingKey(p models.Platform) (string, error) {
	if queue, ok := uc.defaultOptions.PlatformQueues[p]; ok && queue != "" {
		return queue, nil
	}
	return "", dispatcher.ErrUnknownRoute
}

func (uc implUseCase) selectPlatforms(platform models.Platform) ([]models.Platform, error) {
	if platform != "" && platform != "all" {
		if _, ok := uc.defaultOptions.PlatformQueues[platform]; !ok {
			return nil, dispatcher.ErrUnknownRoute
		}
		return []models.Platform{platform}, nil
	}

	platforms := make([]models.Platform, 0, len(uc.defaultOptions.PlatformQueues))
	for p := range uc.defaultOptions.PlatformQueues {
		platforms = append(platforms, p)
	}

	if len(platforms) == 0 {
		return nil, dispatcher.ErrUnknownRoute
	}

	return platforms, nil
}
