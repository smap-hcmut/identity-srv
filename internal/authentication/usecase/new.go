package usecase

import (
	"smap-api/internal/audit"
	"smap-api/internal/user"
	"smap-api/pkg/encrypter"
	pkgGoogle "smap-api/pkg/google"
	pkgJWT "smap-api/pkg/jwt"
	pkgLog "smap-api/pkg/log"
	"smap-api/pkg/oauth"
	pkgRedis "smap-api/pkg/redis"
	"smap-api/pkg/scope"
	"time"

	"smap-api/config"
)

type ImplUsecase struct {
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

// --- Session types ---

// SessionManager handles session storage and retrieval
type SessionManager struct {
	redis *pkgRedis.Client
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
	redis *pkgRedis.Client
}

// --- Groups types ---

// GroupsManager handles Google Groups fetching and caching
type GroupsManager struct {
	googleClient *pkgGoogle.Client
	redis        *pkgRedis.Client
	cacheTTL     time.Duration
}

// --- Role mapping types ---

// RoleMapper handles group-to-role mapping logic
type RoleMapper struct {
	roleMapping map[string][]string
	defaultRole string
}

// rolePriority maps roles to their priority level for selecting highest privilege
var rolePriority = map[string]int{
	"ADMIN":   3,
	"ANALYST": 2,
	"VIEWER":  1,
}

// --- Redirect types ---

// RedirectValidator validates redirect URLs against allowed list
type RedirectValidator struct {
	allowedURLs []string
}

// --- Rate limiting types ---

// RateLimiter implements login rate limiting to prevent brute force attacks
type RateLimiter struct {
	redis          *pkgRedis.Client
	maxAttempts    int
	windowDuration time.Duration
	blockDuration  time.Duration
}

func New(l pkgLog.Logger, scope scope.Manager, encrypt encrypter.Encrypter, userUC user.UseCase) *ImplUsecase {
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
func NewSessionManager(redisClient *pkgRedis.Client, ttl time.Duration) *SessionManager {
	return &SessionManager{
		redis: redisClient,
		ttl:   ttl,
	}
}

// NewBlacklistManager creates a new blacklist manager
func NewBlacklistManager(redisClient *pkgRedis.Client) *BlacklistManager {
	return &BlacklistManager{
		redis: redisClient,
	}
}

// NewGroupsManager creates a new groups manager
func NewGroupsManager(googleClient *pkgGoogle.Client, redis *pkgRedis.Client) *GroupsManager {
	return &GroupsManager{
		googleClient: googleClient,
		redis:        redis,
		cacheTTL:     5 * time.Minute,
	}
}

// NewRoleMapper creates a new role mapper
func NewRoleMapper(cfg *config.Config) *RoleMapper {
	return &RoleMapper{
		roleMapping: cfg.AccessControl.RoleMapping,
		defaultRole: cfg.AccessControl.DefaultRole,
	}
}

// NewRedirectValidator creates a new redirect validator
func NewRedirectValidator(allowedURLs []string) *RedirectValidator {
	return &RedirectValidator{
		allowedURLs: allowedURLs,
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *pkgRedis.Client, maxAttempts int, windowDuration, blockDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:          redisClient,
		maxAttempts:    maxAttempts,
		windowDuration: windowDuration,
		blockDuration:  blockDuration,
	}
}

// --- Setters (called after initialization) ---

func (u *ImplUsecase) SetAuditPublisher(publisher audit.Publisher) {
	u.auditPublisher = publisher
}

func (u *ImplUsecase) SetSessionManager(manager *SessionManager) {
	u.sessionManager = manager
}

func (u *ImplUsecase) SetBlacklistManager(manager *BlacklistManager) {
	u.blacklistManager = manager
}

func (u *ImplUsecase) SetJWTManager(manager *pkgJWT.Manager) {
	u.jwtManager = manager
}

func (u *ImplUsecase) SetGroupsManager(manager *GroupsManager) {
	u.groupsManager = manager
}

func (u *ImplUsecase) SetRoleMapper(mapper *RoleMapper) {
	u.roleMapper = mapper
}

func (u *ImplUsecase) SetOAuthProvider(provider oauth.Provider) {
	u.oauthProvider = provider
}

func (u *ImplUsecase) SetRateLimiter(limiter *RateLimiter) {
	u.rateLimiter = limiter
}

func (u *ImplUsecase) SetRedirectValidator(validator *RedirectValidator) {
	u.redirectValidator = validator
}

func (u *ImplUsecase) SetAccessControl(allowedDomains, blockedEmails []string) {
	u.allowedDomains = allowedDomains
	u.blockedEmails = blockedEmails
}
