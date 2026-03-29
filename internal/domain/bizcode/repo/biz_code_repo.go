package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	bizcodemodel "wdkr-marketplace-service/internal/domain/bizcode/bizcode_model"
)

type bizCodeRepoImpl struct{}

func NewBizCodeRepo() BizCodeRepo {
	return &bizCodeRepoImpl{}
}

func (r *bizCodeRepoImpl) ListAll(ctx context.Context) ([]*bizcodemodel.BizCode, error) {
	var list []*bizcodemodel.BizCode
	err := database.FromContext(ctx).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *bizCodeRepoImpl) GetByCode(ctx context.Context, code string) (*bizcodemodel.BizCode, error) {
	if code == "" {
		return nil, nil
	}
	var row bizcodemodel.BizCode
	err := database.FromContext(ctx).Where("code = ?", code).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}
