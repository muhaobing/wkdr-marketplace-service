package repo

import (
	"context"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

// UserRepo 用户仓储接口
type UserRepo interface {
	// GetUserById 根据ID获取用户
	GetUserById(ctx context.Context, id uint) (*usermodel.User, error)

	// GetUserByTelNo 根据手机号获取用户
	GetUserByTelNo(ctx context.Context, telNo string) (*usermodel.User, error)

	// GetUserByEmail 根据邮箱获取用户
	GetUserByEmail(ctx context.Context, email string) (*usermodel.User, error)

	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *usermodel.User) error

	// UpdateUserBinding 更新用户绑定信息
	UpdateUserBinding(ctx context.Context, id uint, binding usermodel.BindingMap) error

	// UpdateUserSecretKey 更新用户密钥
	UpdateUserSecretKey(ctx context.Context, id uint, secretKey string) error

	// GetUserForUpdate 获取用户信息（加锁）
	GetUserForUpdate(ctx context.Context, id uint) (*usermodel.User, error)
}
