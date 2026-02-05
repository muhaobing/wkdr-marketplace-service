package repo

import (
	"context"
	"errors"

	"github.com/muhaobing-eng/std-go/go-common/database"
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
			"sku_name":        sku.SkuName,
			"sku_avatar":      sku.SkuAvatar,
			"sku_desc":        sku.SkuDesc,
			"cost":            sku.Cost,
			"delivery_method": sku.DeliveryMethod,
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

// ListSkusByBizCode 根据业务编码获取商品列表
func (r *skuRepoImpl) ListSkusByBizCode(ctx context.Context, bizCode string, status *uint8, offset, limit int) ([]*skumodel.Sku, error) {
	var skus []*skumodel.Sku
	query := database.FromContext(ctx).Where("biz_code = ?", bizCode)

	if status != nil {
		query = query.Where("sku_status = ?", *status)
	}

	err := query.Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&skus).Error
	return skus, err
}

// CountSkusByBizCode 统计业务编码下的商品数量
func (r *skuRepoImpl) CountSkusByBizCode(ctx context.Context, bizCode string, status *uint8) (int64, error) {
	var count int64
	query := database.FromContext(ctx).Model(&skumodel.Sku{}).Where("biz_code = ?", bizCode)

	if status != nil {
		query = query.Where("sku_status = ?", *status)
	}

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
