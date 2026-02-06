package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing-eng/std-go/go-common/cache"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/utils/auth_utils"
	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
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
	conf := config.GetConf()
	path := ctx.Request.URL.Path

	// 检查是否在白名单中
	for _, whitePath := range conf.Auth.Whitelist {
		if path == whitePath || strings.HasPrefix(path, whitePath) {
			ctx.Next()
			return
		}
	}

	token := ctx.GetHeader(authHeaderKey)
	// 处理 Bearer token 格式
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	sessionId, err := auth_utils.ParseAuthToken(token, conf.Auth.AesKey)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	redis := cache.FromContext(ctx.Request.Context())
	session, err := redis.Get(ctx, sessionId).Result()
	if err != nil {
		_ = ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// 检查是否需要管理员权限
	requireAdmin := false
	for _, adminPath := range conf.Auth.AdminPaths {
		if strings.HasPrefix(path, adminPath) {
			requireAdmin = true
			break
		}
	}

	// 如果需要管理员权限，检查用户角色
	if requireAdmin {
		user, err := usermodel.UserFromJSON(session)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusUnauthorized, errors.New("invalid session"))
			return
		}
		if !user.IsAdmin() {
			_ = ctx.AbortWithError(http.StatusForbidden, errors.New("admin permission required"))
			return
		}
	}

	// session 续期
	_ = redis.Expire(ctx, sessionId, time.Duration(conf.Auth.Expiration)*time.Second)

	auth_utils.WrapGinContext(ctx, session)
	ctx.Next()
}
