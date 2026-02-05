package repo

import (
	"context"

	"wdkr-marketplace-service/internal/domain/payment/payment_model"
)

// PaymentRepo 支付仓储接口
type PaymentRepo interface {
	// ==================== 支付订单 ====================

	// CreatePaymentOrder 创建支付订单
	CreatePaymentOrder(ctx context.Context, order *payment_model.PaymentOrder) error

	// GetPaymentOrderByOrderNo 根据订单号获取支付订单
	GetPaymentOrderByOrderNo(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error)

	// GetPaymentOrderForUpdate 获取支付订单（加锁）
	GetPaymentOrderForUpdate(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error)

	// GetPaymentOrderById 根据ID获取支付订单
	GetPaymentOrderById(ctx context.Context, id uint64) (*payment_model.PaymentOrder, error)

	// GetPaymentOrderByBizOrderNo 根据业务订单号获取支付订单
	GetPaymentOrderByBizOrderNo(ctx context.Context, bizOrderNo string) (*payment_model.PaymentOrder, error)

	// UpdatePaymentOrderStatus 更新支付订单状态
	UpdatePaymentOrderStatus(ctx context.Context, orderNo string, status uint8, channelOrderNo string, payTime uint32) error

	// UpdatePaymentOrderToClosed 更新支付订单为已关闭
	UpdatePaymentOrderToClosed(ctx context.Context, orderNo string) error

	// UpdatePaymentOrderToRefunded 更新支付订单为已退款
	UpdatePaymentOrderToRefunded(ctx context.Context, orderNo string) error

	// ListPaymentOrdersByUserId 根据用户ID获取支付订单列表
	ListPaymentOrdersByUserId(ctx context.Context, userId uint64, offset, limit int) ([]*payment_model.PaymentOrder, error)

	// CountPaymentOrdersByUserId 统计用户的支付订单数量
	CountPaymentOrdersByUserId(ctx context.Context, userId uint64) (int64, error)

	// ==================== 退款记录 ====================

	// CreatePaymentRefund 创建退款记录
	CreatePaymentRefund(ctx context.Context, refund *payment_model.PaymentRefund) error

	// GetPaymentRefundByRefundNo 根据退款单号获取退款记录
	GetPaymentRefundByRefundNo(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error)

	// GetPaymentRefundForUpdate 获取退款记录（加锁）
	GetPaymentRefundForUpdate(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error)

	// UpdatePaymentRefundStatus 更新退款状态
	UpdatePaymentRefundStatus(ctx context.Context, refundNo string, status uint8, channelRefundNo string) error

	// ListPaymentRefundsByOrderNo 根据支付订单号获取退款记录列表
	ListPaymentRefundsByOrderNo(ctx context.Context, orderNo string) ([]*payment_model.PaymentRefund, error)
}
