package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	cartmodel "wdkr-marketplace-service/internal/domain/cart/cart_model"
)

type cartRepoImpl struct{}

// NewCartRepo 创建购物车仓储实现
func NewCartRepo() CartRepo {
	return &cartRepoImpl{}
}

// GetCartItem 获取购物车中指定商品
func (r *cartRepoImpl) GetCartItem(ctx context.Context, userId, skuId uint64) (*cartmodel.CartItem, error) {
	var item cartmodel.CartItem
	err := database.FromContext(ctx).
		Where("user_id = ? AND sku_id = ?", userId, skuId).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetCartItemsByUserId 获取用户购物车列表
func (r *cartRepoImpl) GetCartItemsByUserId(ctx context.Context, userId uint64) ([]*cartmodel.CartItem, error) {
	var items []*cartmodel.CartItem
	err := database.FromContext(ctx).
		Where("user_id = ?", userId).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

// AddCartItem 添加购物车商品
func (r *cartRepoImpl) AddCartItem(ctx context.Context, item *cartmodel.CartItem) error {
	return database.FromContext(ctx).Create(item).Error
}

// UpdateCartItemQuantity 更新购物车商品数量
func (r *cartRepoImpl) UpdateCartItemQuantity(ctx context.Context, id uint64, quantity int) error {
	return database.FromContext(ctx).
		Model(&cartmodel.CartItem{}).
		Where("id = ?", id).
		Update("quantity", quantity).Error
}

// DeleteCartItem 删除购物车商品
func (r *cartRepoImpl) DeleteCartItem(ctx context.Context, id uint64) error {
	return database.FromContext(ctx).
		Where("id = ?", id).
		Delete(&cartmodel.CartItem{}).Error
}

// DeleteCartItemByUserAndSku 根据用户和商品删除购物车项
func (r *cartRepoImpl) DeleteCartItemByUserAndSku(ctx context.Context, userId, skuId uint64) error {
	return database.FromContext(ctx).
		Where("user_id = ? AND sku_id = ?", userId, skuId).
		Delete(&cartmodel.CartItem{}).Error
}

// DeleteCartItemsByUserAndSkus 批量删除购物车商品
func (r *cartRepoImpl) DeleteCartItemsByUserAndSkus(ctx context.Context, userId uint64, skuIds []uint64) error {
	if len(skuIds) == 0 {
		return nil
	}
	return database.FromContext(ctx).
		Where("user_id = ? AND sku_id IN ?", userId, skuIds).
		Delete(&cartmodel.CartItem{}).Error
}

// ClearCart 清空用户购物车
func (r *cartRepoImpl) ClearCart(ctx context.Context, userId uint64) error {
	return database.FromContext(ctx).
		Where("user_id = ?", userId).
		Delete(&cartmodel.CartItem{}).Error
}

// CountCartItems 统计用户购物车商品数量
func (r *cartRepoImpl) CountCartItems(ctx context.Context, userId uint64) (int64, error) {
	var count int64
	err := database.FromContext(ctx).
		Model(&cartmodel.CartItem{}).
		Where("user_id = ?", userId).
		Count(&count).Error
	return count, err
}
