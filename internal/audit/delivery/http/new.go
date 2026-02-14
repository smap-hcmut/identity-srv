package http

import (
	"smap-api/internal/audit/repository"
	"smap-api/pkg/discord"
	pkgLog "smap-api/pkg/log"
)

type handler struct {
	l       pkgLog.Logger
	repo    repository.Repository
	discord *discord.Discord
}

func New(l pkgLog.Logger, repo repository.Repository, discord *discord.Discord) handler {
	return handler{
		l:       l,
		repo:    repo,
		discord: discord,
	}
}
