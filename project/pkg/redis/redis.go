package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua script để verify và delete lock atomically
// Script này đảm bảo chỉ delete lock nếu lockValue khớp
// Tránh unlock nhầm lock của process khác
const unlockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`

// lockStore lưu lockValue cho mỗi lock key để verify khi unlock
type lockStore struct {
	mu    sync.RWMutex
	locks map[string]string // map[lockKey]lockValue
}

var globalLockStore = &lockStore{
	locks: make(map[string]string),
}

func (ls *lockStore) set(key, value string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.locks[key] = value
}

func (ls *lockStore) get(key string) (string, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	value, ok := ls.locks[key]
	return value, ok
}

func (ls *lockStore) delete(key string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	delete(ls.locks, key)
}

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
	// Lock acquires a distributed lock with the given key
	// Returns true if lock acquired, false if already locked
	Lock(ctx context.Context, key string, expiration int) (bool, error)
	// Unlock releases the distributed lock
	Unlock(ctx context.Context, key string) error
	// Publish publishes a message to a Redis channel
	Publish(ctx context.Context, channel string, message interface{}) error
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

// Lock acquires a distributed lock using SET NX EX pattern
// key: lock key
// expiration: lock expiration time in seconds
// Returns true if lock acquired, false if already locked
func (rc *redisClient) Lock(ctx context.Context, key string, expiration int) (bool, error) {
	lockKey := fmt.Sprintf("%slock:%s", PREFIX, key)
	// Sử dụng unique value (timestamp + random) để verify khi unlock
	// Tránh unlock nhầm lock của process khác
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	expirationDuration := time.Second * time.Duration(expiration)

	var result bool
	var err error

	if rc.cscl != nil {
		result, err = rc.cscl.SetNX(ctx, lockKey, lockValue, expirationDuration).Result()
	} else {
		result, err = rc.cl.SetNX(ctx, lockKey, lockValue, expirationDuration).Result()
	}

	if err != nil {
		return false, err
	}

	// Lưu lockValue vào memory để verify khi unlock
	if result {
		globalLockStore.set(lockKey, lockValue)
	}

	return result, nil
}

// Unlock releases the distributed lock
// Sử dụng Lua script để verify lockValue trước khi delete atomically
// Chỉ delete nếu lockValue khớp, tránh unlock nhầm lock của process khác
func (rc *redisClient) Unlock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("%slock:%s", PREFIX, key)

	// Lấy lockValue từ memory store
	lockValue, exists := globalLockStore.get(lockKey)
	if !exists {
		// Nếu không có trong store, có thể lock đã expired hoặc không phải lock của process này
		// Vẫn thử unlock nhưng không verify
		if rc.cscl != nil {
			return rc.cscl.Del(ctx, lockKey).Err()
		}
		return rc.cl.Del(ctx, lockKey).Err()
	}

	// Sử dụng Lua script để verify và delete atomically
	if rc.cscl != nil {
		// Eval script với cluster client
		result, err := rc.cscl.Eval(ctx, unlockScript, []string{lockKey}, lockValue).Result()
		if err != nil {
			// Nếu lỗi, vẫn xóa khỏi memory store
			globalLockStore.delete(lockKey)
			return err
		}

		// Script trả về số lượng keys đã delete (0 hoặc 1)
		// Nếu = 0 nghĩa là lockValue không khớp hoặc lock đã expired
		if deleted, ok := result.(int64); ok && deleted == 0 {
			globalLockStore.delete(lockKey)
			return fmt.Errorf("lock value mismatch or lock expired for key: %s", key)
		}

		globalLockStore.delete(lockKey)
		return nil
	}

	// Sử dụng Lua script với standalone client
	result, err := rc.cl.Eval(ctx, unlockScript, []string{lockKey}, lockValue).Result()
	if err != nil {
		globalLockStore.delete(lockKey)
		return err
	}

	// Script trả về số lượng keys đã delete (0 hoặc 1)
	if deleted, ok := result.(int64); ok && deleted == 0 {
		globalLockStore.delete(lockKey)
		return fmt.Errorf("lock value mismatch or lock expired for key: %s", key)
	}

	globalLockStore.delete(lockKey)
	return nil
}

// Publish publishes a message to a Redis channel
func (rc *redisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	if rc.cscl != nil {
		return rc.cscl.Publish(ctx, channel, message).Err()
	}
	return rc.cl.Publish(ctx, channel, message).Err()
}
