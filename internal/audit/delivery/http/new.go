package http

import (
	"identity-srv/internal/audit/repository"
	"identity-srv/pkg/discord"
	pkgLog "identity-srv/pkg/log"
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
