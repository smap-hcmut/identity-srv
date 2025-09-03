package smtp

import (
	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/internal/core/smtp"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
)

type implService struct {
	l   log.Logger
	cfg config.SMTPConfig
}

func New(l log.Logger, cfg config.SMTPConfig) smtp.UseCase {
	return implService{
		l:   l,
		cfg: cfg,
	}
}
