package usecase

import (
	"smap-project/internal/keyword"
	"smap-project/pkg/collector"
	"smap-project/pkg/llm"
	pkgLog "smap-project/pkg/log"
	"time"
)

type usecase struct {
	l               pkgLog.Logger
	llmProvider     llm.Provider
	collectorClient collector.Client
	clock           func() time.Time
}

// New creates a new keyword use case
func New(l pkgLog.Logger, llmProvider llm.Provider, collectorClient collector.Client) keyword.UseCase {
	return &usecase{
		l:               l,
		llmProvider:     llmProvider,
		collectorClient: collectorClient,
		clock:           time.Now,
	}
}
