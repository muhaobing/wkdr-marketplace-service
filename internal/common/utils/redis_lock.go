package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/muhaobing-eng/std-go/go-common/cache"
)

var luaRelease = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`)

var luaRenew = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("pexpire", KEYS[1], ARGV[2])
end
return 0
`)

type RedisLock struct {
	key    string
	value  string
	ttl    time.Duration
	client *redis.Client

	mu       sync.Mutex
	stopOnce sync.Once
	stopCh   chan struct{}
}

// AcquireDistributedLock attempts to acquire a distributed lock from Redis (obtained via ctx).
// ttl controls both the key expiration and the watch renewal interval (renewed at ttl/3).
// Returns the lock on success, or nil if the lock is already held by another caller.
func AcquireDistributedLock(ctx context.Context, key string, ttl time.Duration) (*RedisLock, error) {
	rdb := cache.FromContext(ctx)
	if rdb == nil {
		return nil, fmt.Errorf("redis client not found in context")
	}

	value, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("generate lock token: %w", err)
	}

	ok, err := rdb.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("setnx failed: %w", err)
	}
	if !ok {
		return nil, nil
	}

	l := &RedisLock{
		key:    key,
		value:  value,
		ttl:    ttl,
		client: rdb,
		stopCh: make(chan struct{}),
	}
	go l.watch()
	return l, nil
}

// Release releases the lock and stops the watch.
// Safe to call multiple times.
func (l *RedisLock) Release(ctx context.Context) error {
	l.stopOnce.Do(func() { close(l.stopCh) })

	res, err := luaRelease.Run(ctx, l.client, []string{l.key}, l.value).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("unlock failed: %w", err)
	}
	if res == 0 {
		return fmt.Errorf("lock already expired or not owned")
	}
	return nil
}

func (l *RedisLock) watch() {
	interval := l.ttl / 3
	if interval < time.Millisecond*100 {
		interval = time.Millisecond * 100
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), interval)
			ttlMs := l.ttl.Milliseconds()
			luaRenew.Run(ctx, l.client, []string{l.key}, l.value, ttlMs)
			cancel()
		}
	}
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
