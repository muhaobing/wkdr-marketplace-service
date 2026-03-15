package cron

import (
	"context"
	"fmt"
	"time"

	"wdkr-marketplace-service/internal/domain/order"
	ordermodel "wdkr-marketplace-service/internal/domain/order/order_model"
	"wdkr-marketplace-service/internal/domain/order/repo"
)

const paidScanLimit = 200

// OrderFulfillTask 已支付订单履约扫描任务
type OrderFulfillTask struct {
	orderRepo    repo.OrderRepo
	orderService order.OrderService
}

func NewOrderFulfillTask(orderRepo repo.OrderRepo, orderService order.OrderService) *OrderFulfillTask {
	return &OrderFulfillTask{
		orderRepo:    orderRepo,
		orderService: orderService,
	}
}

func (t *OrderFulfillTask) Name() string {
	return "order_fulfill_scan"
}

func (t *OrderFulfillTask) Ticker() time.Duration {
	return 10 * time.Second
}

func (t *OrderFulfillTask) Handle(ctx context.Context) error {
	orders, err := t.orderRepo.ListOrdersByStatus(ctx, ordermodel.OrderStatusPaid, paidScanLimit)
	if err != nil {
		return fmt.Errorf("list paid orders: %w", err)
	}

	for _, o := range orders {
		if err := t.orderService.AutoFulfill(ctx, o.OrderNo); err != nil {
			fmt.Printf("[OrderFulfillTask] fulfill order %s failed: %v\n", o.OrderNo, err)
		}
	}
	return nil
}
