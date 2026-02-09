package usecase

import (
	"database/sql"
	"smap-api/internal/audit"
	"smap-api/internal/authentication"
	"smap-api/internal/user"
	"smap-api/pkg/encrypter"
	pkgLog "smap-api/pkg/log"
	"smap-api/pkg/scope"
	"time"
)

type implUsecase struct {
	l              pkgLog.Logger
	scope          scope.Manager
	encrypt        encrypter.Encrypter
	userUC         user.UseCase
	db             *sql.DB
	clock          func() time.Time
	auditPublisher audit.Publisher
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

// SetDB sets the database connection (called after initialization)
func (u *implUsecase) SetDB(db *sql.DB) {
	u.db = db
}

// SetAuditPublisher sets the audit publisher (called after initialization)
func (u *implUsecase) SetAuditPublisher(publisher audit.Publisher) {
	u.auditPublisher = publisher
}
