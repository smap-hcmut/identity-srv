package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	"github.com/nguyentantai21042004/smap-api/internal/models"
)

func (uc implUseCase) Dispatch(ctx context.Context, req models.CrawlRequest) (models.CollectorTask, error) {
	if req.TaskType == "" || req.Platform == "" {
		return models.CollectorTask{}, dispatcher.ErrInvalidInput
	}

	task := models.CollectorTask{
		JobID:         req.JobID,
		Platform:      req.Platform,
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

	payload, err := mapPayload(req.Platform, req.TaskType, req.Payload)
	if err != nil {
		return models.CollectorTask{}, err
	}
	task.Payload = payload
	task.RoutingKey = fmt.Sprintf("crawler.%s.queue", req.Platform)

	if err := uc.prod.PublishTask(ctx, task); err != nil {
		return models.CollectorTask{}, fmt.Errorf("%w: %v", dispatcher.ErrPublish, err)
	}

	return task, nil
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
