package usecase

import (
	"smap-collector/internal/results"
	"smap-collector/pkg/log"
	"smap-collector/pkg/project"
)

type implUseCase struct {
	l              log.Logger
	projectClient  project.Client
}

func NewUseCase(l log.Logger, projectClient project.Client) results.UseCase {
	return &implUseCase{
		l:              l,
		projectClient:  projectClient,
	}
}
