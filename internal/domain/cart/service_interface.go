package cart

import (
	"context"

	cartmodel "wdkr-marketplace-service/internal/domain/cart/cart_model"
	"wdkr-marketplace-service/internal/domain/order"
)

// AddToCartRequest 添加购物车请求
type AddToCartRequest struct {
	UserId   uint64 `json:"user_id"`  // 用户ID
	SkuId    uint64 `json:"sku_id"`   // 商品ID
	Quantity int    `json:"quantity"` // 数量
}

// RemoveFromCartRequest 移除购物车请求
type RemoveFromCartRequest struct {
	UserId uint64 `json:"user_id"` // 用户ID
	SkuId  uint64 `json:"sku_id"`  // 商品ID
}

// UpdateCartItemRequest 更新购物车商品数量请求
type UpdateCartItemRequest struct {
	UserId   uint64 `json:"user_id"`  // 用户ID
	SkuId    uint64 `json:"sku_id"`   // 商品ID
	Quantity int    `json:"quantity"` // 数量
}

// CartListResponse 购物车列表响应
type CartListResponse struct {
	Items      []*cartmodel.CartItemWithSku `json:"items"`       // 购物车商品列表
	TotalCount int                          `json:"total_count"` // 商品总数
	TotalCost  float32                      `json:"total_cost"`  // 总价
}

// CheckoutRequest 购物车下单请求
type CheckoutRequest struct {
	UserId  uint64   `json:"user_id"`  // 用户ID
	SkuIds  []uint64 `json:"sku_ids"`  // 要下单的商品ID列表（从购物车中选择）
	PayType string   `json:"pay_type"` // 支付类型：ecoin/money
	Remark  string   `json:"remark"`   // 备注
}

// CartService 购物车服务接口
type CartService interface {
	// AddToCart 添加商品到购物车
	AddToCart(ctx context.Context, req *AddToCartRequest) (*cartmodel.CartItem, error)

	// RemoveFromCart 从购物车移除商品
	RemoveFromCart(ctx context.Context, req *RemoveFromCartRequest) error

	// UpdateCartItem 更新购物车商品数量
	UpdateCartItem(ctx context.Context, req *UpdateCartItemRequest) (*cartmodel.CartItem, error)

	// ClearCart 清空购物车
	ClearCart(ctx context.Context, userId uint64) error

	// GetCartList 获取购物车列表
	GetCartList(ctx context.Context, userId uint64) (*CartListResponse, error)

	// Checkout 购物车下单（下单并移除对应商品）
	Checkout(ctx context.Context, req *CheckoutRequest) (*order.CreateOrderResponse, error)
}
