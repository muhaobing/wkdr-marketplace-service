package auth_utils

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

// ErrSessionNotFound 未在上下文中找到 session（与未登录/未带 token 区分，供业务返回明确文案）
var ErrSessionNotFound = errors.New("session not found")

// UserFromGinContext 从已登录 session 解析当前用户（需先经过 AuthValidationHandler）
func UserFromGinContext(ctx *gin.Context) (*usermodel.User, error) {
	v, ok := ctx.Get(contextKeySession)
	if !ok {
		return nil, ErrSessionNotFound
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil, ErrSessionNotFound
	}
	u, err := usermodel.UserFromJSON(s)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}
	return u, nil
}
