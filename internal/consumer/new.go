package consumer

import (
	"errors"

	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/internal/appconfig/oauth"
	pkgCrt "github.com/nguyentantai21042004/smap-api/pkg/encrypter"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
	"github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
	pkgRedis "github.com/nguyentantai21042004/smap-api/pkg/redis"
)

type Consumer struct {
	l            pkgLog.Logger
	jwtSecretKey string
	amqpConn     *rabbitmq.Connection
	encrypter    pkgCrt.Encrypter
	telegram     TeleCredentials
	internalKey  string
	mongoDB      *mongo.Database
	smtpConfig   config.SMTPConfig
	redisClient  *pkgRedis.Client
	oauthConfig  oauth.OauthConfig
}

type ConsumerConfig struct {
	JwtSecretKey string
	AMQPConn     *rabbitmq.Connection
	Encrypter    pkgCrt.Encrypter
	Telegram     TeleCredentials
	InternalKey  string
	MongoDB      *mongo.Database
	RedisClient  *pkgRedis.Client
	SMTPConfig   config.SMTPConfig
	OauthConfig  oauth.OauthConfig
}

type TeleCredentials struct {
	BotKey string
	ChatIDs
}

type ChatIDs struct {
	ReportBug int64
}

func New(l pkgLog.Logger, cfg ConsumerConfig) (*Consumer, error) {

	h := &Consumer{
		l:            l,
		amqpConn:     cfg.AMQPConn,
		jwtSecretKey: cfg.JwtSecretKey,
		encrypter:    cfg.Encrypter,
		telegram:     cfg.Telegram,
		internalKey:  cfg.InternalKey,
		mongoDB:      cfg.MongoDB,
		smtpConfig:   cfg.SMTPConfig,
		redisClient:  cfg.RedisClient,
		oauthConfig:  cfg.OauthConfig,
	}

	if err := h.validate(); err != nil {
		return nil, err
	}

	return h, nil
}

func (s Consumer) validate() error {
	requiredDeps := []struct {
		dep interface{}
		msg string
	}{
		{s.l, "logger is required"},
		{s.amqpConn, "amqpConn is required"},
		{s.jwtSecretKey, "jwtSecretKey is required"},
		{s.encrypter, "encrypter is required"},
		{s.telegram, "telegram is required"},
		{s.internalKey, "internalKey is required"},
		{s.mongoDB, "mongoDB is required"},
		{s.redisClient, "redisClient is required"},
		{s.oauthConfig, "oauthConfig is required"},
		{s.smtpConfig, "smtpConfig is required"},
	}

	for _, dep := range requiredDeps {
		if dep.dep == nil {
			return errors.New(dep.msg)
		}
	}

	return nil
}
