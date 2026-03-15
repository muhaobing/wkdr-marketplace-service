package bootstrap

import (
	"context"

	"github.com/muhaobing/std-go/go-common/cache"
	"github.com/muhaobing/std-go/go-common/database"
	"github.com/muhaobing/std-go/restserver/lib"
)

// BackgroundContext 构建包含全局 DB 和 Redis 的后台 context，供异步 goroutine 使用
func BackgroundContext() context.Context {
	ctx := context.Background()
	if db := lib.GetDB(); db != nil {
		ctx = database.Context(ctx, db)
	}
	if rdb := lib.GetRedis(); rdb != nil {
		ctx = cache.Context(ctx, rdb)
	}
	return ctx
}
