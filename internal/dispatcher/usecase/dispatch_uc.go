package usecase

import (
	"context"
	"fmt"

	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	"github.com/nguyentantai21042004/smap-api/internal/models"
)

func (uc implUseCase) Dispatch(ctx context.Context, req models.CrawlRequest) ([]models.CollectorTask, error) {
	if req.TaskType == "" {
		return nil, dispatcher.ErrInvalidInput
	}

	targetPlatforms := uc.selectPlatforms()
	if len(targetPlatforms) == 0 {
		return nil, dispatcher.ErrUnknownRoute
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
		validateTask(&task, uc.defaultOptions)

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

		if err := uc.PublishTask(ctx, task); err != nil {
			return nil, fmt.Errorf("%w: %v", dispatcher.ErrPublish, err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}


