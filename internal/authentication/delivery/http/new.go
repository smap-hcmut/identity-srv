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
	"smap-api/pkg/oauth"
)

type handler struct {
	l                 pkgLog.Logger
	uc                authentication.UseCase
	discord           *discord.Discord
	cookieConfig      config.CookieConfig
	config            *config.Config
	oauthProvider     oauth.Provider
	jwtManager        *pkgJWT.Manager
	sessionManager    *usecase.SessionManager
	blacklistManager  *usecase.BlacklistManager
	googleClient      *pkgGoogle.Client
	groupsManager     *usecase.GroupsManager
	roleMapper        *usecase.RoleMapper
	userRepo          userrepo.Repository
	redirectValidator *usecase.RedirectValidator
	rateLimiter       *usecase.RateLimiter
}

func New(l pkgLog.Logger, uc authentication.UseCase, discord *discord.Discord, cfg *config.Config, jwtManager *pkgJWT.Manager, sessionManager *usecase.SessionManager, blacklistManager *usecase.BlacklistManager, googleClient *pkgGoogle.Client, groupsManager *usecase.GroupsManager, roleMapper *usecase.RoleMapper, userRepo userrepo.Repository, redirectValidator *usecase.RedirectValidator, rateLimiter *usecase.RateLimiter, oauthProvider oauth.Provider) handler {
	return handler{
		l:                 l,
		uc:                uc,
		discord:           discord,
		cookieConfig:      cfg.Cookie,
		config:            cfg,
		oauthProvider:     oauthProvider,
		jwtManager:        jwtManager,
		sessionManager:    sessionManager,
		blacklistManager:  blacklistManager,
		googleClient:      googleClient,
		groupsManager:     groupsManager,
		roleMapper:        roleMapper,
		userRepo:          userRepo,
		redirectValidator: redirectValidator,
		rateLimiter:       rateLimiter,
	}
}
