package cron

import (
	"context"
	"fmt"
	"time"

	ordermodel "wdkr-marketplace-service/internal/domain/order/order_model"
	"wdkr-marketplace-service/internal/domain/order/repo"
	"wdkr-marketplace-service/internal/domain/payment"
)

const (
	paymentTimeoutMinutes = 15
	pendingScanLimit      = 200
)

// OrderTimeoutTask 待支付订单超时扫描任务
type OrderTimeoutTask struct {
	orderRepo  repo.OrderRepo
	paymentSvc payment.PaymentService
}

func NewOrderTimeoutTask(orderRepo repo.OrderRepo, paymentSvc payment.PaymentService) *OrderTimeoutTask {
	return &OrderTimeoutTask{
		orderRepo:  orderRepo,
		paymentSvc: paymentSvc,
	}
}

func (t *OrderTimeoutTask) Name() string {
	return "order_timeout_scan"
}

func (t *OrderTimeoutTask) Ticker() time.Duration {
	return 10 * time.Second
}

func (t *OrderTimeoutTask) Handle(ctx context.Context) error {
	now := time.Now()

	orders, err := t.orderRepo.ListOrdersByStatus(ctx, ordermodel.OrderStatusPending, pendingScanLimit)
	if err != nil {
		return fmt.Errorf("list pending orders: %w", err)
	}

	for _, order := range orders {
		orderAge := now.Sub(time.Unix(int64(order.Ctime), 0))

		if orderAge > paymentTimeoutMinutes*time.Minute {
			t.cancelOrder(ctx, order)
			continue
		}

		if order.IsMoneyPay() && order.PaymentOrderNo != "" {
			t.syncPayment(ctx, order)
		}
	}
	return nil
}

func (t *OrderTimeoutTask) cancelOrder(ctx context.Context, order *ordermodel.Order) {
	cancelTime := uint32(time.Now().Unix())
	if err := t.orderRepo.UpdateOrderToCancelled(ctx, order.OrderNo, cancelTime, "支付超时自动取消"); err != nil {
		fmt.Printf("[OrderTimeoutTask] cancel order %s failed: %v\n", order.OrderNo, err)
		return
	}

	if order.PaymentOrderNo != "" {
		_ = t.paymentSvc.ClosePayment(ctx, order.PaymentOrderNo)
	}

	fmt.Printf("[OrderTimeoutTask] order %s cancelled (timeout)\n", order.OrderNo)
}

func (t *OrderTimeoutTask) syncPayment(ctx context.Context, order *ordermodel.Order) {
	paymentOrder, err := t.paymentSvc.SyncPaymentStatus(ctx, order.PaymentOrderNo)
	if err != nil {
		fmt.Printf("[OrderTimeoutTask] sync payment for order %s failed: %v\n", order.OrderNo, err)
		return
	}

	if paymentOrder.IsPaid() {
		if err := t.orderRepo.UpdateOrderToPaid(ctx, order.OrderNo, paymentOrder.PayTime); err != nil {
			fmt.Printf("[OrderTimeoutTask] update order %s to paid failed: %v\n", order.OrderNo, err)
		} else {
			fmt.Printf("[OrderTimeoutTask] order %s payment confirmed\n", order.OrderNo)
		}
	}
}
