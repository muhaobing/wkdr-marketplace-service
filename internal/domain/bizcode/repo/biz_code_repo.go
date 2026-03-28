package repo

import (
	"context"

	"github.com/muhaobing/std-go/go-common/database"

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
