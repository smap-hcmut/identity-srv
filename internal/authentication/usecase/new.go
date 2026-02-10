package usecase

import (
	"smap-api/internal/audit"
	"smap-api/internal/authentication"
	"smap-api/internal/user"
	"smap-api/pkg/encrypter"
	pkgJWT "smap-api/pkg/jwt"
	pkgLog "smap-api/pkg/log"
	"smap-api/pkg/oauth"
	"smap-api/pkg/scope"
	"time"
)

type implUsecase struct {
	l                 pkgLog.Logger
	scope             scope.Manager
	encrypt           encrypter.Encrypter
	userUC            user.UseCase
	clock             func() time.Time
	auditPublisher    audit.Publisher
	sessionManager    *SessionManager
	blacklistManager  *BlacklistManager
	jwtManager        *pkgJWT.Manager
	groupsManager     *GroupsManager
	roleMapper        *RoleMapper
	oauthProvider     oauth.Provider
	rateLimiter       *RateLimiter
	redirectValidator *RedirectValidator
	allowedDomains    []string
	blockedEmails     []string
}

func New(l pkgLog.Logger, scope scope.Manager, encrypt encrypter.Encrypter, userUC user.UseCase) authentication.UseCase {
	return &implUsecase{
		l:       l,
		scope:   scope,
		encrypt: encrypt,
		userUC:  userUC,
		clock:   time.Now,
	}
}

// --- Setters (called after initialization) ---

func (u *implUsecase) SetAuditPublisher(publisher audit.Publisher) {
	u.auditPublisher = publisher
}

func (u *implUsecase) SetSessionManager(manager *SessionManager) {
	u.sessionManager = manager
}

func (u *implUsecase) SetBlacklistManager(manager *BlacklistManager) {
	u.blacklistManager = manager
}

func (u *implUsecase) SetJWTManager(manager *pkgJWT.Manager) {
	u.jwtManager = manager
}

func (u *implUsecase) SetGroupsManager(manager *GroupsManager) {
	u.groupsManager = manager
}

func (u *implUsecase) SetRoleMapper(mapper *RoleMapper) {
	u.roleMapper = mapper
}

func (u *implUsecase) SetOAuthProvider(provider oauth.Provider) {
	u.oauthProvider = provider
}

func (u *implUsecase) SetRateLimiter(limiter *RateLimiter) {
	u.rateLimiter = limiter
}

func (u *implUsecase) SetRedirectValidator(validator *RedirectValidator) {
	u.redirectValidator = validator
}

func (u *implUsecase) SetAccessControl(allowedDomains, blockedEmails []string) {
	u.allowedDomains = allowedDomains
	u.blockedEmails = blockedEmails
}
