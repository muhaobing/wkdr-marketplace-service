package repo

import (
	"context"

	bizcodemodel "wdkr-marketplace-service/internal/domain/bizcode/bizcode_model"
)

// BizCodeRepo 业务平台编码枚举
type BizCodeRepo interface {
	// ListAll 按 id 排序
	ListAll(ctx context.Context) ([]*bizcodemodel.BizCode, error)
	// GetByCode 按 code 查询
	GetByCode(ctx context.Context, code string) (*bizcodemodel.BizCode, error)
}
