package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func Connect(opts ClientOptions) (Client, error) {
	if opts.csclo != nil {
		cscl := redis.NewClusterClient(opts.csclo)
		return &redisClient{cscl: cscl}, nil
	}
	cl := redis.NewClient(opts.clo)
	return &redisClient{cl: cl}, nil
}

type Database interface {
	Client() Client
}

type Client interface {
	Disconnect() error
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration int) error
	Expire(ctx context.Context, key string, expiration int) error
	Del(ctx context.Context, keys ...string) error

	MGet(ctx context.Context, keys ...string) ([]interface{}, error)

	HSet(ctx context.Context, key string, field string, value interface{}) error
	HGet(ctx context.Context, key string, field string) ([]byte, error)
}

type redisClient struct {
	cl   *redis.Client
	cscl *redis.ClusterClient
}

func (rc *redisClient) Disconnect() error {
	if rc.cscl != nil {
		return rc.cscl.Close()
	}
	return rc.cl.Close()
}

func (rc *redisClient) Get(ctx context.Context, key string) ([]byte, error) {
	if rc.cscl != nil {
		key = fmt.Sprintf("%s%s", PREFIX, key)
		return rc.cscl.Get(ctx, key).Bytes()
	}
	return rc.cl.Get(ctx, key).Bytes()
}

func (rc *redisClient) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	if rc.cscl != nil {
		key = fmt.Sprintf("%s%s", PREFIX, key)
		return rc.cscl.Set(ctx, key, value, time.Second*time.Duration(expiration)).Err()
	}
	return rc.cl.Set(ctx, key, value, time.Second*time.Duration(expiration)).Err()
}

func (rc *redisClient) Expire(ctx context.Context, key string, expiration int) error {
	if rc.cscl != nil {
		key = fmt.Sprintf("%s%s", PREFIX, key)
		return rc.cscl.Expire(ctx, key, time.Second*time.Duration(expiration)).Err()
	}
	return rc.cl.Expire(ctx, key, time.Second*time.Duration(expiration)).Err()
}

func (rc *redisClient) Del(ctx context.Context, keys ...string) error {
	if rc.cscl != nil {
		for i, key := range keys {
			keys[i] = fmt.Sprintf("%s%s", PREFIX, key)
		}
		return rc.cscl.Del(ctx, keys...).Err()
	}
	return rc.cl.Del(ctx, keys...).Err()
}

func (rc *redisClient) HSet(ctx context.Context, key string, field string, value interface{}) error {
	if rc.cscl != nil {
		key = fmt.Sprintf("%s%s", PREFIX, key)
		return rc.cscl.HSet(ctx, key, field, value).Err()
	}
	return rc.cl.HSet(ctx, key, field, value).Err()
}

func (rc *redisClient) HGet(ctx context.Context, key string, field string) ([]byte, error) {
	if rc.cscl != nil {
		key = fmt.Sprintf("%s%s", PREFIX, key)
		return rc.cscl.HGet(ctx, key, field).Bytes()
	}
	return rc.cl.HGet(ctx, key, field).Bytes()
}

func (rc *redisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if rc.cscl != nil {
		for i, key := range keys {
			keys[i] = fmt.Sprintf("%s%s", PREFIX, key)
		}
		return rc.cscl.MGet(ctx, keys...).Result()
	}
	return rc.cl.MGet(ctx, keys...).Result()
}
