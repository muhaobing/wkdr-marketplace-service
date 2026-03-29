package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/muhaobing/std-go/go-common/cache"
)

// EcoinIdempotencyRedisKey 仅用 source_type 与 source_id 拼接（加固定前缀避免与其它业务键冲突）
func EcoinIdempotencyRedisKey(sourceType, sourceID string) string {
	return fmt.Sprintf("wkdr:ecoin:idem:%s:%s", sourceType, sourceID)
}

// TryAcquireIdempotencyKey SETNX，TTL 内同一业务流水仅允许成功执行一次；成功后勿删除 key，直至 TTL 自然过期。
func TryAcquireIdempotencyKey(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	rdb := cache.FromContext(ctx)
	if rdb == nil {
		return false, fmt.Errorf("redis client not found in context")
	}
	return rdb.SetNX(ctx, key, "1", ttl).Result()
}

// ReleaseIdempotencyKey 业务失败时删除 key，允许调用方重试。
func ReleaseIdempotencyKey(ctx context.Context, key string) {
	rdb := cache.FromContext(ctx)
	if rdb == nil {
		return
	}
	_ = rdb.Del(ctx, key).Err()
}
