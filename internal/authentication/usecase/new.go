package usecase

import (
	"smap-api/internal/audit"
	"smap-api/internal/authentication"
	"smap-api/internal/user"
	"smap-api/pkg/encrypter"
	pkgLog "smap-api/pkg/log"
	"smap-api/pkg/scope"
	"time"
)

type implUsecase struct {
	l                pkgLog.Logger
	scope            scope.Manager
	encrypt          encrypter.Encrypter
	userUC           user.UseCase
	clock            func() time.Time
	auditPublisher   audit.Publisher
	sessionManager   *SessionManager
	blacklistManager *BlacklistManager
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

// SetAuditPublisher sets the audit publisher (called after initialization)
func (u *implUsecase) SetAuditPublisher(publisher audit.Publisher) {
	u.auditPublisher = publisher
}

// SetSessionManager sets the session manager (called after initialization)
func (u *implUsecase) SetSessionManager(manager *SessionManager) {
	u.sessionManager = manager
}

// SetBlacklistManager sets the blacklist manager (called after initialization)
func (u *implUsecase) SetBlacklistManager(manager *BlacklistManager) {
	u.blacklistManager = manager
}
