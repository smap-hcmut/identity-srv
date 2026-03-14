package usecase

import (
	"identity-srv/internal/user"
	"identity-srv/pkg/oauth"
	"time"

	"github.com/smap-hcmut/shared-libs/go/auth"
	"github.com/smap-hcmut/shared-libs/go/encrypter"
	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/redis"

	"identity-srv/config"
)

type ImplUsecase struct {
	l                 log.Logger
	scope             auth.Manager
	encrypt           encrypter.Encrypter
	userUC            user.UseCase
	clock             func() time.Time
	sessionManager    *SessionManager
	blacklistManager  *BlacklistManager
	jwtManager        auth.Manager
	roleMapper        *RoleMapper
	oauthProvider     oauth.Provider
	redirectValidator *RedirectValidator
	allowedDomains    []string
	blockedEmails     []string
}

// --- Session types ---

// SessionManager handles session storage and retrieval
type SessionManager struct {
	redis redis.IRedis
	ttl   time.Duration
}

// SessionData represents session information stored in Redis
type SessionData struct {
	UserID    string    `json:"user_id"`
	JTI       string    `json:"jti"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// --- Blacklist types ---

// BlacklistManager handles token blacklist operations
type BlacklistManager struct {
	redis redis.IRedis
}

// --- Role mapping types ---

// RoleMapper handles email-to-role mapping logic
type RoleMapper struct {
	userRoles   map[string]string
	defaultRole string
}

// --- Redirect types ---

// RedirectValidator validates redirect URLs against allowed list
type RedirectValidator struct {
	allowedURLs []string
}

func New(l log.Logger, scope auth.Manager, encrypt encrypter.Encrypter, userUC user.UseCase) *ImplUsecase {
	return &ImplUsecase{
		l:       l,
		scope:   scope,
		encrypt: encrypt,
		userUC:  userUC,
		clock:   time.Now,
	}
}

// --- Sub-manager factory functions ---

// NewSessionManager creates a new session manager
func NewSessionManager(redisClient redis.IRedis, ttl time.Duration) *SessionManager {
	return &SessionManager{
		redis: redisClient,
		ttl:   ttl,
	}
}

// NewBlacklistManager creates a new blacklist manager
func NewBlacklistManager(redisClient redis.IRedis) *BlacklistManager {
	return &BlacklistManager{
		redis: redisClient,
	}
}

// NewRoleMapper creates a new role mapper
func NewRoleMapper(cfg *config.Config) *RoleMapper {
	return &RoleMapper{
		userRoles:   cfg.AccessControl.UserRoles,
		defaultRole: cfg.AccessControl.DefaultRole,
	}
}

// NewRedirectValidator creates a new redirect validator
func NewRedirectValidator(allowedURLs []string) *RedirectValidator {
	return &RedirectValidator{
		allowedURLs: allowedURLs,
	}
}

// --- Setters (called after initialization) ---

func (u *ImplUsecase) SetSessionManager(manager *SessionManager) {
	u.sessionManager = manager
}

func (u *ImplUsecase) SetBlacklistManager(manager *BlacklistManager) {
	u.blacklistManager = manager
}

func (u *ImplUsecase) SetJWTManager(manager auth.Manager) {
	u.jwtManager = manager
}

func (u *ImplUsecase) SetRoleMapper(mapper *RoleMapper) {
	u.roleMapper = mapper
}

func (u *ImplUsecase) SetOAuthProvider(provider oauth.Provider) {
	u.oauthProvider = provider
}

func (u *ImplUsecase) SetRedirectValidator(validator *RedirectValidator) {
	u.redirectValidator = validator
}

func (u *ImplUsecase) SetAccessControl(allowedDomains, blockedEmails []string) {
	u.allowedDomains = allowedDomains
	u.blockedEmails = blockedEmails
}
