package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	"wdkr-marketplace-service/internal/domain/payment/payment_model"
)

type paymentRepoImpl struct{}

// NewPaymentRepo 创建支付仓储实现
func NewPaymentRepo() PaymentRepo {
	return &paymentRepoImpl{}
}

// ==================== 支付订单 ====================

// CreatePaymentOrder 创建支付订单
func (r *paymentRepoImpl) CreatePaymentOrder(ctx context.Context, order *payment_model.PaymentOrder) error {
	return database.FromContext(ctx).Create(order).Error
}

// GetPaymentOrderByOrderNo 根据订单号获取支付订单
func (r *paymentRepoImpl) GetPaymentOrderByOrderNo(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error) {
	var order payment_model.PaymentOrder
	err := database.FromContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetPaymentOrderForUpdate 获取支付订单（加锁）
func (r *paymentRepoImpl) GetPaymentOrderForUpdate(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error) {
	var order payment_model.PaymentOrder
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

// GetPaymentOrderById 根据ID获取支付订单
func (r *paymentRepoImpl) GetPaymentOrderById(ctx context.Context, id uint64) (*payment_model.PaymentOrder, error) {
	var order payment_model.PaymentOrder
	err := database.FromContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetPaymentOrderByBizOrderNo 根据业务订单号获取支付订单
func (r *paymentRepoImpl) GetPaymentOrderByBizOrderNo(ctx context.Context, bizOrderNo string) (*payment_model.PaymentOrder, error) {
	var order payment_model.PaymentOrder
	err := database.FromContext(ctx).Where("biz_order_no = ?", bizOrderNo).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// UpdatePaymentOrderStatus 更新支付订单状态
func (r *paymentRepoImpl) UpdatePaymentOrderStatus(ctx context.Context, orderNo string, status uint8, channelOrderNo string, payTime uint32) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if channelOrderNo != "" {
		updates["channel_order_no"] = channelOrderNo
	}
	if payTime > 0 {
		updates["pay_time"] = payTime
	}

	return database.FromContext(ctx).Model(&payment_model.PaymentOrder{}).
		Where("order_no = ?", orderNo).
		Updates(updates).Error
}

// UpdatePaymentOrderToClosed 更新支付订单为已关闭
func (r *paymentRepoImpl) UpdatePaymentOrderToClosed(ctx context.Context, orderNo string) error {
	return database.FromContext(ctx).Model(&payment_model.PaymentOrder{}).
		Where("order_no = ?", orderNo).
		Update("status", payment_model.PaymentStatusClosed).Error
}

// UpdatePaymentOrderToRefunded 更新支付订单为已退款
func (r *paymentRepoImpl) UpdatePaymentOrderToRefunded(ctx context.Context, orderNo string) error {
	return database.FromContext(ctx).Model(&payment_model.PaymentOrder{}).
		Where("order_no = ?", orderNo).
		Update("status", payment_model.PaymentStatusRefunded).Error
}

// ListPaymentOrdersByUserId 根据用户ID获取支付订单列表
func (r *paymentRepoImpl) ListPaymentOrdersByUserId(ctx context.Context, userId uint64, offset, limit int) ([]*payment_model.PaymentOrder, error) {
	var orders []*payment_model.PaymentOrder
	err := database.FromContext(ctx).
		Where("user_id = ?", userId).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// CountPaymentOrdersByUserId 统计用户的支付订单数量
func (r *paymentRepoImpl) CountPaymentOrdersByUserId(ctx context.Context, userId uint64) (int64, error) {
	var count int64
	err := database.FromContext(ctx).Model(&payment_model.PaymentOrder{}).
		Where("user_id = ?", userId).
		Count(&count).Error
	return count, err
}

// ==================== 退款记录 ====================

// CreatePaymentRefund 创建退款记录
func (r *paymentRepoImpl) CreatePaymentRefund(ctx context.Context, refund *payment_model.PaymentRefund) error {
	return database.FromContext(ctx).Create(refund).Error
}

// GetPaymentRefundByRefundNo 根据退款单号获取退款记录
func (r *paymentRepoImpl) GetPaymentRefundByRefundNo(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error) {
	var refund payment_model.PaymentRefund
	err := database.FromContext(ctx).Where("refund_no = ?", refundNo).First(&refund).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

// GetPaymentRefundForUpdate 获取退款记录（加锁）
func (r *paymentRepoImpl) GetPaymentRefundForUpdate(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error) {
	var refund payment_model.PaymentRefund
	err := database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").
		Where("refund_no = ?", refundNo).First(&refund).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

// UpdatePaymentRefundStatus 更新退款状态
func (r *paymentRepoImpl) UpdatePaymentRefundStatus(ctx context.Context, refundNo string, status uint8, channelRefundNo string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if channelRefundNo != "" {
		updates["channel_refund_no"] = channelRefundNo
	}

	return database.FromContext(ctx).Model(&payment_model.PaymentRefund{}).
		Where("refund_no = ?", refundNo).
		Updates(updates).Error
}

// ListPaymentRefundsByOrderNo 根据支付订单号获取退款记录列表
func (r *paymentRepoImpl) ListPaymentRefundsByOrderNo(ctx context.Context, orderNo string) ([]*payment_model.PaymentRefund, error) {
	var refunds []*payment_model.PaymentRefund
	err := database.FromContext(ctx).
		Where("order_no = ?", orderNo).
		Order("id DESC").
		Find(&refunds).Error
	return refunds, err
}
