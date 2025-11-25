package usecase

import (
	"encoding/json"
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	"github.com/nguyentantai21042004/smap-api/internal/models"
)

func (uc implUseCase) queueRoutingKey(p models.Platform) (string, error) {
	if queue, ok := uc.defaultOptions.PlatformQueues[p]; ok && queue != "" {
		return queue, nil
	}
	return "", dispatcher.ErrUnknownRoute
}

func (uc implUseCase) selectPlatforms() []models.Platform {
	platforms := make([]models.Platform, 0, len(uc.defaultOptions.PlatformQueues))
	for p := range uc.defaultOptions.PlatformQueues {
		platforms = append(platforms, p)
	}
	return platforms
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

// validateTask sets defaults for attempt/max/timestamps.
func validateTask(task *models.CollectorTask, opts dispatcher.Options) {
	if task.Attempt <= 0 {
		task.Attempt = 1
	}
	if task.MaxAttempts <= 0 {
		task.MaxAttempts = opts.DefaultMaxAttempts
	}
	if task.EmittedAt.IsZero() {
		task.EmittedAt = time.Now().UTC()
	}
}
