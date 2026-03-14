package repo

import (
	"context"

	ordermodel "wdkr-marketplace-service/internal/domain/order/order_model"
)

// OrderRepo 订单仓储接口
type OrderRepo interface {
	// ==================== 订单 ====================

	// CreateOrder 创建订单
	CreateOrder(ctx context.Context, order *ordermodel.Order) error

	// GetOrderById 根据ID获取订单
	GetOrderById(ctx context.Context, id uint64) (*ordermodel.Order, error)

	// GetOrderByOrderNo 根据订单号获取订单
	GetOrderByOrderNo(ctx context.Context, orderNo string) (*ordermodel.Order, error)

	// GetOrderForUpdate 获取订单（加锁）
	GetOrderForUpdate(ctx context.Context, orderNo string) (*ordermodel.Order, error)

	// UpdateOrderStatus 更新订单状态
	UpdateOrderStatus(ctx context.Context, orderNo string, status uint8) error

	// UpdateOrderToPaid 更新订单为已支付
	UpdateOrderToPaid(ctx context.Context, orderNo string, payTime uint32) error

	// UpdateOrderToFulfilled 更新订单为已履约
	UpdateOrderToFulfilled(ctx context.Context, orderNo string, fulfillTime uint32) error

	// UpdateOrderToCancelled 更新订单为已取消
	UpdateOrderToCancelled(ctx context.Context, orderNo string, cancelTime uint32, cancelReason string) error

	// UpdateOrderToRefunded 更新订单为已退款
	UpdateOrderToRefunded(ctx context.Context, orderNo string) error

	// UpdateOrderPaymentOrderNo 更新订单的支付订单号
	UpdateOrderPaymentOrderNo(ctx context.Context, orderNo string, paymentOrderNo string) error

	// ListOrdersByUserId 根据用户ID获取订单列表
	ListOrdersByUserId(ctx context.Context, userId uint64, status *uint8, offset, limit int) ([]*ordermodel.Order, error)

	// CountOrdersByUserId 统计用户的订单数量
	CountOrdersByUserId(ctx context.Context, userId uint64, status *uint8) (int64, error)

	// ListOrdersByStatus 按状态查询订单列表（定时任务使用）
	ListOrdersByStatus(ctx context.Context, status uint8, limit int) ([]*ordermodel.Order, error)

	// ==================== 订单明细 ====================

	// CreateOrderItems 批量创建订单明细
	CreateOrderItems(ctx context.Context, items []*ordermodel.OrderItem) error

	// GetOrderItemsByOrderId 获取订单的所有明细项
	GetOrderItemsByOrderId(ctx context.Context, orderId uint64) ([]*ordermodel.OrderItem, error)

	// GetOrderItemsByOrderNo 根据订单号获取所有明细项
	GetOrderItemsByOrderNo(ctx context.Context, orderNo string) ([]*ordermodel.OrderItem, error)

	// UpdateOrderItemFulfillStatus 更新订单明细的履约状态
	UpdateOrderItemFulfillStatus(ctx context.Context, id uint64, status uint8, fulfillTime uint32, fulfillMsg string) error

	// GetPendingFulfillItems 获取待履约的订单明细
	GetPendingFulfillItems(ctx context.Context, orderNo string) ([]*ordermodel.OrderItem, error)
}
