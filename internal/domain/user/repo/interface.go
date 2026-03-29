package repo

import (
	"context"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

// UserRepo 用户仓储接口
type UserRepo interface {
	// GetUserById 根据ID获取用户
	GetUserById(ctx context.Context, id uint) (*usermodel.User, error)

	// GetUserByTelNoAndCompany 根据手机号 + company_id 获取用户（company_id=0 为个人）
	GetUserByTelNoAndCompany(ctx context.Context, telNo string, companyId uint64) (*usermodel.User, error)

	// GetUserByEmailAndCompany 根据邮箱 + company_id 获取用户
	GetUserByEmailAndCompany(ctx context.Context, email string, companyId uint64) (*usermodel.User, error)

	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *usermodel.User) error

	// UpdateUserSecretKey 更新用户密钥
	UpdateUserSecretKey(ctx context.Context, id uint, secretKey string) error

	// UpdateUserContact 更新手机号、邮箱
	UpdateUserContact(ctx context.Context, id uint, telNo, email string) error

	// GetUserForUpdate 获取用户信息（加锁）
	GetUserForUpdate(ctx context.Context, id uint) (*usermodel.User, error)
}

// UserBindingRepo 用户绑定仓储接口
type UserBindingRepo interface {
	// GetBindingByBiz 根据业务信息获取绑定记录
	GetBindingByBiz(ctx context.Context, bizCode string, bizUserId uint64) (*usermodel.UserBinding, error)

	// GetBindingsByUserId 获取用户所有绑定记录
	GetBindingsByUserId(ctx context.Context, userId uint) ([]*usermodel.UserBinding, error)

	// GetBindingByUserAndBiz 获取用户在指定业务的绑定记录
	GetBindingByUserAndBiz(ctx context.Context, userId uint, bizCode string) (*usermodel.UserBinding, error)

	// CreateBinding 创建绑定记录
	CreateBinding(ctx context.Context, binding *usermodel.UserBinding) error

	// DeleteBinding 删除绑定记录
	DeleteBinding(ctx context.Context, userId uint, bizCode string) error
}
