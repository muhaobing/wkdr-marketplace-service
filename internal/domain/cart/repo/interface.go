package repo

import (
	"context"

	cartmodel "wdkr-marketplace-service/internal/domain/cart/cart_model"
)

// CartRepo 购物车仓储接口
type CartRepo interface {
	// GetCartItem 获取购物车中指定商品
	GetCartItem(ctx context.Context, userId, skuId uint64) (*cartmodel.CartItem, error)

	// GetCartItemsByUserId 获取用户购物车列表
	GetCartItemsByUserId(ctx context.Context, userId uint64) ([]*cartmodel.CartItem, error)

	// AddCartItem 添加购物车商品
	AddCartItem(ctx context.Context, item *cartmodel.CartItem) error

	// UpdateCartItemQuantity 更新购物车商品数量
	UpdateCartItemQuantity(ctx context.Context, id uint64, quantity int) error

	// DeleteCartItem 删除购物车商品
	DeleteCartItem(ctx context.Context, id uint64) error

	// DeleteCartItemByUserAndSku 根据用户和商品删除购物车项
	DeleteCartItemByUserAndSku(ctx context.Context, userId, skuId uint64) error

	// DeleteCartItemsByUserAndSkus 批量删除购物车商品
	DeleteCartItemsByUserAndSkus(ctx context.Context, userId uint64, skuIds []uint64) error

	// ClearCart 清空用户购物车
	ClearCart(ctx context.Context, userId uint64) error

	// CountCartItems 统计用户购物车商品数量
	CountCartItems(ctx context.Context, userId uint64) (int64, error)
}
