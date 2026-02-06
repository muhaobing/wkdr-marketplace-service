package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing-eng/std-go/go-common/cache"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/utils/auth_utils"
)

const (
	AuthValidationHandlerKey = "auth-validation-handler"

	authHeaderKey = "Authorization"
)

type AuthValidationHandler struct{}

func (c *AuthValidationHandler) Name() string {
	return AuthValidationHandlerKey
}

func (c *AuthValidationHandler) Handle(ctx *gin.Context) {
	token := ctx.GetHeader(authHeaderKey)
	sessionId, err := auth_utils.ParseAuthToken(token, config.GetConf().Auth.AesKey)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	redis := cache.FromContext(ctx)
	session, err := redis.Get(ctx, sessionId).Result()
	if err != nil {
		_ = ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	_ = redis.Expire(ctx, sessionId, time.Duration(config.GetConf().Auth.Expiration)*time.Second) // session续期

	auth_utils.WrapGinContext(ctx, session)
	ctx.Next()
}
