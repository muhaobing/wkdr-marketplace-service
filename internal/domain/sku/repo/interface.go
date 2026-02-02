package repo

import (
	"context"

	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
)

// SkuRepo SKU仓储接口
type SkuRepo interface {
	// GetSkuById 根据ID获取商品
	GetSkuById(ctx context.Context, id uint64) (*skumodel.Sku, error)

	// GetSkuByCode 根据业务编码和商品编码获取商品
	GetSkuByCode(ctx context.Context, bizCode, skuCode string) (*skumodel.Sku, error)

	// GetSkuForUpdate 获取商品信息（加锁）
	GetSkuForUpdate(ctx context.Context, id uint64) (*skumodel.Sku, error)

	// CreateSku 创建商品
	CreateSku(ctx context.Context, sku *skumodel.Sku) error

	// UpdateSku 更新商品信息
	UpdateSku(ctx context.Context, sku *skumodel.Sku) error

	// UpdateSkuStatus 更新商品状态
	UpdateSkuStatus(ctx context.Context, id uint64, status uint8) error

	// DeleteSku 删除商品
	DeleteSku(ctx context.Context, id uint64) error

	// ListSkusByBizCode 根据业务编码获取商品列表
	ListSkusByBizCode(ctx context.Context, bizCode string, status *uint8, offset, limit int) ([]*skumodel.Sku, error)

	// CountSkusByBizCode 统计业务编码下的商品数量
	CountSkusByBizCode(ctx context.Context, bizCode string, status *uint8) (int64, error)
}
