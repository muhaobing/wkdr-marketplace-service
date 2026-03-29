package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
)

type skuRepoImpl struct{}

// NewSkuRepo 创建SKU仓储实现
func NewSkuRepo() SkuRepo {
	return &skuRepoImpl{}
}

// GetSkuById 根据ID获取商品
func (r *skuRepoImpl) GetSkuById(ctx context.Context, id uint64) (*skumodel.Sku, error) {
	var sku skumodel.Sku
	err := database.FromContext(ctx).Where("id = ?", id).First(&sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

// GetSkuByCode 根据业务编码和商品编码获取商品
func (r *skuRepoImpl) GetSkuByCode(ctx context.Context, bizCode, skuCode string) (*skumodel.Sku, error) {
	var sku skumodel.Sku
	err := database.FromContext(ctx).Where("biz_code = ? AND sku_code = ?", bizCode, skuCode).First(&sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

// GetSkuForUpdate 获取商品信息（加锁）
func (r *skuRepoImpl) GetSkuForUpdate(ctx context.Context, id uint64) (*skumodel.Sku, error) {
	var sku skumodel.Sku
	err := database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(&sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

// CreateSku 创建商品
func (r *skuRepoImpl) CreateSku(ctx context.Context, sku *skumodel.Sku) error {
	return database.FromContext(ctx).Create(sku).Error
}

// UpdateSku 更新商品信息
func (r *skuRepoImpl) UpdateSku(ctx context.Context, sku *skumodel.Sku) error {
	return database.FromContext(ctx).Model(&skumodel.Sku{}).
		Where("id = ?", sku.Id).
		Updates(map[string]interface{}{
			"sku_name":             sku.SkuName,
			"sku_avatar":           sku.SkuAvatar,
			"sku_desc":             sku.SkuDesc,
			"cost":                 sku.Cost,
			"delivery_method":      sku.DeliveryMethod,
			"fulfill_mode":         sku.FulfillMode,
			"fulfill_ecoin_amount": sku.FulfillEcoinAmount,
			"multi_select":         sku.MultiSelect,
			"sku_scope":            sku.SkuScope,
		}).Error
}

// UpdateSkuStatus 更新商品状态
func (r *skuRepoImpl) UpdateSkuStatus(ctx context.Context, id uint64, status uint8) error {
	return database.FromContext(ctx).Model(&skumodel.Sku{}).
		Where("id = ?", id).
		Update("sku_status", status).Error
}

// DeleteSku 删除商品
func (r *skuRepoImpl) DeleteSku(ctx context.Context, id uint64) error {
	return database.FromContext(ctx).Where("id = ?", id).Delete(&skumodel.Sku{}).Error
}

// buildSkuListQuery 构建商品列表查询条件
func (r *skuRepoImpl) buildSkuListQuery(ctx context.Context, filter *SkuListFilter) *gorm.DB {
	query := database.FromContext(ctx).Model(&skumodel.Sku{})

	if filter.BizCode != "" {
		query = query.Where("biz_code = ?", filter.BizCode)
	}

	if filter.SkuName != "" {
		query = query.Where("sku_name LIKE ?", "%"+filter.SkuName+"%")
	}

	if filter.Status != nil {
		query = query.Where("sku_status = ?", *filter.Status)
	}

	if filter.MarketplaceSkuScopeFilter {
		// 必须用 []uint64/[]int 等，勿用 []uint8：与 []byte 同型，GORM 会按二进制绑定导致 SQL 语法错误
		var inScopes []uint64
		if filter.MarketplaceVisitorIsEnterprise {
			inScopes = []uint64{uint64(skumodel.SkuScopeUniversal), uint64(skumodel.SkuScopeEnterprise)}
		} else {
			inScopes = []uint64{uint64(skumodel.SkuScopeUniversal), uint64(skumodel.SkuScopePersonal)}
		}
		query = query.Where("sku_scope IN ?", inScopes)
	} else if filter.SkuScope != nil {
		query = query.Where("sku_scope = ?", *filter.SkuScope)
	}

	return query
}

// ListSkus 获取商品列表（支持多条件查询）
func (r *skuRepoImpl) ListSkus(ctx context.Context, filter *SkuListFilter) ([]*skumodel.Sku, error) {
	var skus []*skumodel.Sku

	query := r.buildSkuListQuery(ctx, filter)

	err := query.Order("id DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&skus).Error

	return skus, err
}

// CountSkus 统计商品数量
func (r *skuRepoImpl) CountSkus(ctx context.Context, filter *SkuListFilter) (int64, error) {
	var count int64

	query := r.buildSkuListQuery(ctx, filter)

	err := query.Count(&count).Error
	return count, err
}

// GetSkusByIds 根据ID列表批量获取商品
func (r *skuRepoImpl) GetSkusByIds(ctx context.Context, ids []uint64) ([]*skumodel.Sku, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var skus []*skumodel.Sku
	err := database.FromContext(ctx).Where("id IN ?", ids).Find(&skus).Error
	return skus, err
}
