package cron

import (
	"context"
	"time"

	"wdkr-marketplace-service/internal/domain/ecoinbill"
)

// EcoinBillAutoCancelTask 预扣积分账单超时自动取消并退款（>15 分钟 incomplete）
type EcoinBillAutoCancelTask struct {
	svc ecoinbill.EcoinBillService
}

func NewEcoinBillAutoCancelTask(svc ecoinbill.EcoinBillService) *EcoinBillAutoCancelTask {
	return &EcoinBillAutoCancelTask{svc: svc}
}

func (t *EcoinBillAutoCancelTask) Name() string {
	return "ecoin_bill_auto_cancel"
}

func (t *EcoinBillAutoCancelTask) Ticker() time.Duration {
	return 30 * time.Second
}

func (t *EcoinBillAutoCancelTask) Handle(ctx context.Context) error {
	return t.svc.AutoCancelStaleBills(ctx)
}
