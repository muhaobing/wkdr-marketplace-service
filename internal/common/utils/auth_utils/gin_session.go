package auth_utils

import (
	"errors"

	"github.com/gin-gonic/gin"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

// UserFromGinContext 从已登录 session 解析当前用户（需先经过 AuthValidationHandler）
func UserFromGinContext(ctx *gin.Context) (*usermodel.User, error) {
	v, ok := ctx.Get(contextKeySession)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil, errors.New("unauthorized")
	}
	return usermodel.UserFromJSON(s)
}
