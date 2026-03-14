package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/muhaobing-eng/std-go/go-common/cache"
	"github.com/muhaobing-eng/std-go/go-common/database"

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

// NewOrderTimeoutTask 创建待支付订单超时任务
func NewOrderTimeoutTask(orderRepo repo.OrderRepo, paymentSvc payment.PaymentService) *Task {
	t := &OrderTimeoutTask{
		orderRepo:  orderRepo,
		paymentSvc: paymentSvc,
	}
	return &Task{
		Name:    "order_timeout_scan",
		Ticker:  10 * time.Second,
		Handler: t.Run,
	}
}

func (t *OrderTimeoutTask) Run() {
	ctx := t.buildContext()
	now := time.Now()

	orders, err := t.orderRepo.ListOrdersByStatus(ctx, ordermodel.OrderStatusPending, pendingScanLimit)
	if err != nil {
		fmt.Printf("[OrderTimeoutTask] list pending orders failed: %v\n", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		orderAge := now.Sub(time.Unix(int64(order.Ctime), 0))

		if orderAge > paymentTimeoutMinutes*time.Minute {
			t.cancelOrder(ctx, order)
			continue
		}

		// 货币支付且有支付单号 → 主动查询微信支付状态
		if order.IsMoneyPay() && order.PaymentOrderNo != "" {
			t.syncPayment(ctx, order)
		}
	}
}

func (t *OrderTimeoutTask) cancelOrder(ctx context.Context, order *ordermodel.Order) {
	cancelTime := uint32(time.Now().Unix())
	if err := t.orderRepo.UpdateOrderToCancelled(ctx, order.OrderNo, cancelTime, "支付超时自动取消"); err != nil {
		fmt.Printf("[OrderTimeoutTask] cancel order %s failed: %v\n", order.OrderNo, err)
		return
	}

	// 同时关闭支付单
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

func (t *OrderTimeoutTask) buildContext() context.Context {
	ctx := context.Background()
	db, err := database.New(database.GetDefaultOption())
	if err != nil {
		fmt.Printf("[OrderTimeoutTask] create db failed: %v\n", err)
		return ctx
	}
	ctx = database.Context(ctx, db)

	redis, err := cache.New(cache.GetDefaultOption())
	if err != nil {
		fmt.Printf("[OrderTimeoutTask] create cache failed: %v\n", err)
		return ctx
	}
	ctx = cache.Context(ctx, redis)
	return ctx
}
