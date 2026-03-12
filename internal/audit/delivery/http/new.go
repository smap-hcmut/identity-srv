package http

import (
	"identity-srv/internal/audit/repository"

	"github.com/smap-hcmut/shared-libs/go/discord"
	"github.com/smap-hcmut/shared-libs/go/log"
)

type handler struct {
	l       log.Logger
	repo    repository.Repository
	discord discord.IDiscord
}

func New(l log.Logger, repo repository.Repository, discord discord.IDiscord) handler {
	return handler{
		l:       l,
		repo:    repo,
		discord: discord,
	}
}
