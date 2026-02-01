package healthy

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing-eng/std-go/go-common/cache"
	"github.com/muhaobing-eng/std-go/go-common/database"
	"github.com/muhaobing-eng/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/utils/http_utils"
)

type HealthyResource struct{}

func NewHealthyResource() *HealthyResource {
	return &HealthyResource{}
}

func (r *HealthyResource) Ping(ctx *gin.Context) {
	if db := database.FromContext(ctx.Request.Context()); db == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("database is nil"))
		return
	}
	if db := cache.FromContext(ctx.Request.Context()); db == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("cache is nil"))
		return
	}
	http_utils.WriteResponse(ctx, nil, nil)
}

func (r *HealthyResource) Router() registry.Registry {
	return func(router *gin.Engine) {
		router.GET("/ping", r.Ping)
	}
}
