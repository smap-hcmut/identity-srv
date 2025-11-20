package smtp

import (
	"smap-api/config"
	"smap-api/internal/smtp"
	pkgLog "smap-api/pkg/log"
)

type implService struct {
	l   pkgLog.Logger
	cfg config.SMTPConfig
}

func New(l pkgLog.Logger, cfg config.SMTPConfig) smtp.UseCase {
	return implService{
		l:   l,
		cfg: cfg,
	}
}
