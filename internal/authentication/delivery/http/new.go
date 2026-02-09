package http

import (
	"smap-api/config"
	"smap-api/internal/authentication"
	"smap-api/internal/authentication/usecase"
	userrepo "smap-api/internal/user/repository"
	"smap-api/pkg/discord"
	pkgGoogle "smap-api/pkg/google"
	pkgJWT "smap-api/pkg/jwt"
	pkgLog "smap-api/pkg/log"

	"golang.org/x/oauth2"
)

type handler struct {
	l                pkgLog.Logger
	uc               authentication.UseCase
	discord          *discord.Discord
	cookieConfig     config.CookieConfig
	config           *config.Config
	oauth2Config     *oauth2.Config
	jwtManager       *pkgJWT.Manager
	sessionManager   *usecase.SessionManager
	blacklistManager *usecase.BlacklistManager
	googleClient     *pkgGoogle.Client
	groupsManager    *usecase.GroupsManager
	roleMapper       *usecase.RoleMapper
	userRepo         userrepo.Repository
}

func New(l pkgLog.Logger, uc authentication.UseCase, discord *discord.Discord, cfg *config.Config, jwtManager *pkgJWT.Manager, sessionManager *usecase.SessionManager, blacklistManager *usecase.BlacklistManager, googleClient *pkgGoogle.Client, groupsManager *usecase.GroupsManager, roleMapper *usecase.RoleMapper, userRepo userrepo.Repository) handler {
	return handler{
		l:                l,
		uc:               uc,
		discord:          discord,
		cookieConfig:     cfg.Cookie,
		config:           cfg,
		oauth2Config:     nil, // Will be initialized via InitOAuth2Config
		jwtManager:       jwtManager,
		sessionManager:   sessionManager,
		blacklistManager: blacklistManager,
		googleClient:     googleClient,
		groupsManager:    groupsManager,
		roleMapper:       roleMapper,
		userRepo:         userRepo,
	}
}
