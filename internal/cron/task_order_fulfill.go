package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/muhaobing-eng/std-go/go-common/cache"
	"github.com/muhaobing-eng/std-go/go-common/database"

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

// NewOrderFulfillTask 创建已支付订单履约任务
func NewOrderFulfillTask(orderRepo repo.OrderRepo, orderService order.OrderService) *Task {
	t := &OrderFulfillTask{
		orderRepo:    orderRepo,
		orderService: orderService,
	}
	return &Task{
		Name:    "order_fulfill_scan",
		Ticker:  10 * time.Second,
		Handler: t.Run,
	}
}

func (t *OrderFulfillTask) Run() {
	ctx := t.buildContext()

	orders, err := t.orderRepo.ListOrdersByStatus(ctx, ordermodel.OrderStatusPaid, paidScanLimit)
	if err != nil {
		fmt.Printf("[OrderFulfillTask] list paid orders failed: %v\n", err)
		return
	}

	for _, o := range orders {
		if err := t.orderService.AutoFulfill(ctx, o.OrderNo); err != nil {
			fmt.Printf("[OrderFulfillTask] fulfill order %s failed: %v\n", o.OrderNo, err)
		}
	}
}

func (t *OrderFulfillTask) buildContext() context.Context {
	ctx := context.Background()
	db, err := database.New(database.GetDefaultOption())
	if err != nil {
		fmt.Printf("[OrderFulfillTask] create db failed: %v\n", err)
		return ctx
	}
	ctx = database.Context(ctx, db)

	redis, err := cache.New(cache.GetDefaultOption())
	if err != nil {
		fmt.Printf("[OrderFulfillTask] create cache failed: %v\n", err)
		return ctx
	}
	ctx = cache.Context(ctx, redis)
	return ctx
}
