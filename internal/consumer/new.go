package consumer

import (
	"database/sql"
	"errors"

	pkgCrt "github.com/nguyentantai21042004/smap-api/pkg/encrypter"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	l            pkgLog.Logger
	jwtSecretKey string
	amqpConn     *rabbitmq.Connection
	encrypter    pkgCrt.Encrypter
	telegram     TeleCredentials
	internalKey  string
	postgresDB   *sql.DB
	redisClient  *redis.Client
}

type ConsumerConfig struct {
	JwtSecretKey string
	AMQPConn     *rabbitmq.Connection
	Encrypter    pkgCrt.Encrypter
	Telegram     TeleCredentials
	InternalKey  string
	PostgresDB   *sql.DB
	RedisClient  *redis.Client
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
		postgresDB:   cfg.PostgresDB,
		redisClient:  cfg.RedisClient,
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
		{s.postgresDB, "postgresDB is required"},
		{s.redisClient, "redisClient is required"},
	}

	for _, dep := range requiredDeps {
		if dep.dep == nil {
			return errors.New(dep.msg)
		}
	}

	return nil
}
