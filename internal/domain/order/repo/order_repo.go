package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	ordermodel "wdkr-marketplace-service/internal/domain/order/order_model"
)

type orderRepoImpl struct{}

// NewOrderRepo 创建订单仓储实现
func NewOrderRepo() OrderRepo {
	return &orderRepoImpl{}
}

// ==================== 订单 ====================

// CreateOrder 创建订单
func (r *orderRepoImpl) CreateOrder(ctx context.Context, order *ordermodel.Order) error {
	return database.FromContext(ctx).Create(order).Error
}

// GetOrderById 根据ID获取订单
func (r *orderRepoImpl) GetOrderById(ctx context.Context, id uint64) (*ordermodel.Order, error) {
	var order ordermodel.Order
	err := database.FromContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (r *orderRepoImpl) GetOrderByOrderNo(ctx context.Context, orderNo string) (*ordermodel.Order, error) {
	var order ordermodel.Order
	err := database.FromContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetOrderForUpdate 获取订单（加锁）
func (r *orderRepoImpl) GetOrderForUpdate(ctx context.Context, orderNo string) (*ordermodel.Order, error) {
	var order ordermodel.Order
	err := database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").
		Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// UpdateOrderStatus 更新订单状态
func (r *orderRepoImpl) UpdateOrderStatus(ctx context.Context, orderNo string, status uint8) error {
	return database.FromContext(ctx).Model(&ordermodel.Order{}).
		Where("order_no = ?", orderNo).
		Update("status", status).Error
}

// UpdateOrderToPaid 更新订单为已支付
func (r *orderRepoImpl) UpdateOrderToPaid(ctx context.Context, orderNo string, payTime uint32) error {
	return database.FromContext(ctx).Model(&ordermodel.Order{}).
		Where("order_no = ?", orderNo).
		Updates(map[string]interface{}{
			"status":   ordermodel.OrderStatusPaid,
			"pay_time": payTime,
		}).Error
}

// UpdateOrderToFulfilled 更新订单为已履约
func (r *orderRepoImpl) UpdateOrderToFulfilled(ctx context.Context, orderNo string, fulfillTime uint32) error {
	return database.FromContext(ctx).Model(&ordermodel.Order{}).
		Where("order_no = ?", orderNo).
		Updates(map[string]interface{}{
			"status":       ordermodel.OrderStatusFulfilled,
			"fulfill_time": fulfillTime,
		}).Error
}

// UpdateOrderToCancelled 更新订单为已取消
func (r *orderRepoImpl) UpdateOrderToCancelled(ctx context.Context, orderNo string, cancelTime uint32, cancelReason string) error {
	return database.FromContext(ctx).Model(&ordermodel.Order{}).
		Where("order_no = ?", orderNo).
		Updates(map[string]interface{}{
			"status":        ordermodel.OrderStatusCancelled,
			"cancel_time":   cancelTime,
			"cancel_reason": cancelReason,
		}).Error
}

// UpdateOrderToRefunded 更新订单为已退款
func (r *orderRepoImpl) UpdateOrderToRefunded(ctx context.Context, orderNo string) error {
	return database.FromContext(ctx).Model(&ordermodel.Order{}).
		Where("order_no = ?", orderNo).
		Update("status", ordermodel.OrderStatusRefunded).Error
}

// UpdateOrderPayType 更新订单的支付类型
func (r *orderRepoImpl) UpdateOrderPayType(ctx context.Context, orderNo string, payType string) error {
	return database.FromContext(ctx).Model(&ordermodel.Order{}).
		Where("order_no = ?", orderNo).
		Update("pay_type", payType).Error
}

// UpdateOrderPaymentOrderNo 更新订单的支付订单号
func (r *orderRepoImpl) UpdateOrderPaymentOrderNo(ctx context.Context, orderNo string, paymentOrderNo string) error {
	return database.FromContext(ctx).Model(&ordermodel.Order{}).
		Where("order_no = ?", orderNo).
		Update("payment_order_no", paymentOrderNo).Error
}

// ListOrdersByUserId 根据用户ID获取订单列表
func (r *orderRepoImpl) ListOrdersByUserId(ctx context.Context, userId uint64, status *uint8, offset, limit int) ([]*ordermodel.Order, error) {
	var orders []*ordermodel.Order
	query := database.FromContext(ctx).Where("user_id = ?", userId)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// CountOrdersByUserId 统计用户的订单数量
func (r *orderRepoImpl) CountOrdersByUserId(ctx context.Context, userId uint64, status *uint8) (int64, error) {
	var count int64
	query := database.FromContext(ctx).Model(&ordermodel.Order{}).Where("user_id = ?", userId)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&count).Error
	return count, err
}

// ListOrdersByStatus 按状态查询订单列表
func (r *orderRepoImpl) ListOrdersByStatus(ctx context.Context, status uint8, limit int) ([]*ordermodel.Order, error) {
	var orders []*ordermodel.Order
	err := database.FromContext(ctx).
		Where("status = ?", status).
		Order("id ASC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// ==================== 订单明细 ====================

// CreateOrderItems 批量创建订单明细
func (r *orderRepoImpl) CreateOrderItems(ctx context.Context, items []*ordermodel.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return database.FromContext(ctx).Create(&items).Error
}

// GetOrderItemsByOrderId 获取订单的所有明细项
func (r *orderRepoImpl) GetOrderItemsByOrderId(ctx context.Context, orderId uint64) ([]*ordermodel.OrderItem, error) {
	var items []*ordermodel.OrderItem
	err := database.FromContext(ctx).
		Where("order_id = ?", orderId).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

// GetOrderItemsByOrderNo 根据订单号获取所有明细项
func (r *orderRepoImpl) GetOrderItemsByOrderNo(ctx context.Context, orderNo string) ([]*ordermodel.OrderItem, error) {
	var items []*ordermodel.OrderItem
	err := database.FromContext(ctx).
		Where("order_no = ?", orderNo).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

// UpdateOrderItemFulfillStatus 更新订单明细的履约状态
func (r *orderRepoImpl) UpdateOrderItemFulfillStatus(ctx context.Context, id uint64, status uint8, fulfillTime uint32, fulfillMsg string) error {
	return database.FromContext(ctx).Model(&ordermodel.OrderItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"fulfill_status": status,
			"fulfill_time":   fulfillTime,
			"fulfill_msg":    fulfillMsg,
		}).Error
}

// GetPendingFulfillItems 获取待履约的订单明细
func (r *orderRepoImpl) GetPendingFulfillItems(ctx context.Context, orderNo string) ([]*ordermodel.OrderItem, error) {
	var items []*ordermodel.OrderItem
	err := database.FromContext(ctx).
		Where("order_no = ? AND fulfill_status = ?", orderNo, ordermodel.FulfillStatusPending).
		Order("id ASC").
		Find(&items).Error
	return items, err
}
