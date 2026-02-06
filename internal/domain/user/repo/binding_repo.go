package repo

import (
	"context"
	"errors"

	"github.com/muhaobing-eng/std-go/go-common/database"
	"gorm.io/gorm"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

type userBindingRepoImpl struct{}

// NewUserBindingRepo 创建用户绑定仓储实现
func NewUserBindingRepo() UserBindingRepo {
	return &userBindingRepoImpl{}
}

// GetBindingByBiz 根据业务信息获取绑定记录
func (r *userBindingRepoImpl) GetBindingByBiz(ctx context.Context, bizCode string, bizUserId uint64) (*usermodel.UserBinding, error) {
	var binding usermodel.UserBinding
	err := database.FromContext(ctx).
		Where("biz_code = ? AND biz_user_id = ?", bizCode, bizUserId).
		First(&binding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &binding, nil
}

// GetBindingsByUserId 获取用户所有绑定记录
func (r *userBindingRepoImpl) GetBindingsByUserId(ctx context.Context, userId uint) ([]*usermodel.UserBinding, error) {
	var bindings []*usermodel.UserBinding
	err := database.FromContext(ctx).
		Where("user_id = ?", userId).
		Find(&bindings).Error
	if err != nil {
		return nil, err
	}
	return bindings, nil
}

// GetBindingByUserAndBiz 获取用户在指定业务的绑定记录
func (r *userBindingRepoImpl) GetBindingByUserAndBiz(ctx context.Context, userId uint, bizCode string) (*usermodel.UserBinding, error) {
	var binding usermodel.UserBinding
	err := database.FromContext(ctx).
		Where("user_id = ? AND biz_code = ?", userId, bizCode).
		First(&binding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &binding, nil
}

// CreateBinding 创建绑定记录
func (r *userBindingRepoImpl) CreateBinding(ctx context.Context, binding *usermodel.UserBinding) error {
	return database.FromContext(ctx).Create(binding).Error
}

// DeleteBinding 删除绑定记录
func (r *userBindingRepoImpl) DeleteBinding(ctx context.Context, userId uint, bizCode string) error {
	return database.FromContext(ctx).
		Where("user_id = ? AND biz_code = ?", userId, bizCode).
		Delete(&usermodel.UserBinding{}).Error
}
