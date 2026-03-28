package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

type userRepoImpl struct{}

// NewUserRepo 创建用户仓储实现
func NewUserRepo() UserRepo {
	return &userRepoImpl{}
}

// GetUserById 根据ID获取用户
func (r *userRepoImpl) GetUserById(ctx context.Context, id uint) (*usermodel.User, error) {
	var user usermodel.User
	err := database.FromContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByTelNo 根据手机号获取用户
func (r *userRepoImpl) GetUserByTelNo(ctx context.Context, telNo string) (*usermodel.User, error) {
	if telNo == "" {
		return nil, nil
	}
	var user usermodel.User
	err := database.FromContext(ctx).Where("tel_no = ?", telNo).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (r *userRepoImpl) GetUserByEmail(ctx context.Context, email string) (*usermodel.User, error) {
	if email == "" {
		return nil, nil
	}
	var user usermodel.User
	err := database.FromContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func (r *userRepoImpl) CreateUser(ctx context.Context, user *usermodel.User) error {
	return database.FromContext(ctx).Create(user).Error
}

// UpdateUserSecretKey 更新用户密钥
func (r *userRepoImpl) UpdateUserSecretKey(ctx context.Context, id uint, secretKey string) error {
	return database.FromContext(ctx).Model(&usermodel.User{}).
		Where("id = ?", id).
		Update("secret_key", secretKey).Error
}

// UpdateUserContact 更新手机号、邮箱
func (r *userRepoImpl) UpdateUserContact(ctx context.Context, id uint, telNo, email string) error {
	return database.FromContext(ctx).Model(&usermodel.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"tel_no": telNo,
			"email":  email,
		}).Error
}

// GetUserForUpdate 获取用户信息（加锁）
func (r *userRepoImpl) GetUserForUpdate(ctx context.Context, id uint) (*usermodel.User, error) {
	var user usermodel.User
	err := database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
