package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
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

// CreateBinding 创建绑定记录
func (r *userBindingRepoImpl) CreateBinding(ctx context.Context, binding *usermodel.UserBinding) error {
	return database.FromContext(ctx).Create(binding).Error
}

// DeleteBindingByBiz 删除指定业务身份绑定记录
func (r *userBindingRepoImpl) DeleteBindingByBiz(ctx context.Context, bizCode string, bizUserId uint64) error {
	return database.FromContext(ctx).
		Where("biz_code = ? AND biz_user_id = ?", bizCode, bizUserId).
		Delete(&usermodel.UserBinding{}).Error
}
