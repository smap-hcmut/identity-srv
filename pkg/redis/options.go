package redis

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/redis/go-redis/v9"
)

type ClientOptions struct {
	clo   *redis.Options
	csclo *redis.ClusterOptions
}

// NewClientOptions creates a new ClientOptions instance.
func NewClientOptions() ClientOptions {
	return ClientOptions{
		clo: &redis.Options{},
	}
}

func (co ClientOptions) SetOptions(opts config.RedisConfig) ClientOptions {
	if opts.RedisStandAlone {
		co.clo.Addr = opts.RedisAddr[0]
		co.clo.MinIdleConns = opts.RedisMinIdleConns
		co.clo.PoolSize = opts.RedisPoolSize
		co.clo.PoolTimeout = time.Duration(opts.RedisPoolTimeout) * time.Second
		co.clo.Password = opts.RedisPassword
		co.clo.DB = opts.RedisDB
		return co
	}
	co.csclo = &redis.ClusterOptions{
		Addrs:        opts.RedisAddr,
		MinIdleConns: opts.RedisMinIdleConns,
		PoolSize:     opts.RedisPoolSize,
		PoolTimeout:  time.Duration(opts.RedisPoolTimeout) * time.Second,
		Password:     opts.RedisPassword,
	}
	return co
}
