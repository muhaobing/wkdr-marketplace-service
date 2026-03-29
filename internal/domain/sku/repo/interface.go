package repo

import (
	"context"

	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
)

// SkuListFilter 商品列表查询条件
type SkuListFilter struct {
	BizCode  string // 业务编码（可选）
	SkuName  string // 商品名称（模糊查询，可选）
	Status   *uint8 // 上架状态（可选）
	SkuScope *uint8 // 运营精确筛选：sku_scope 0/1/2（nil 不按此项过滤）
	// MarketplaceSkuScopeFilter 为 true 时：按访客过滤 sku_scope（个人访客 0+1，企业访客 0+2），对所有商品生效（商城 C 端）
	MarketplaceSkuScopeFilter bool
	// MarketplaceVisitorIsEnterprise 与 MarketplaceSkuScopeFilter 联用：true=企业访客
	MarketplaceVisitorIsEnterprise bool
	Offset                         int // 偏移量
	Limit                          int // 每页数量
}

// SkuRepo SKU仓储接口
type SkuRepo interface {
	// GetSkuById 根据ID获取商品
	GetSkuById(ctx context.Context, id uint64) (*skumodel.Sku, error)

	// GetSkusByIds 根据ID列表批量获取商品
	GetSkusByIds(ctx context.Context, ids []uint64) ([]*skumodel.Sku, error)

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

	// ListSkus 获取商品列表（支持多条件查询）
	ListSkus(ctx context.Context, filter *SkuListFilter) ([]*skumodel.Sku, error)

	// CountSkus 统计商品数量
	CountSkus(ctx context.Context, filter *SkuListFilter) (int64, error)
}
