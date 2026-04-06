package ecoinbill

import (
	"context"

	"wdkr-marketplace-service/internal/domain/ecoinbill/ecoinbill_model"
)

// EcoinBillService 积分账单（预扣 / 确认 / 取消）
type EcoinBillService interface {
	// PreDeduct 预扣积分并创建 incomplete 账单（与扣款同事务）
	PreDeduct(ctx context.Context, userId, companyId uint64, cost float64) (*ecoinbill_model.EcoinBill, error)
	// ConfirmBill 确认账单（completed）
	ConfirmBill(ctx context.Context, billId string, userId, companyId uint64) error
	// CancelBill 取消账单并退款（cancelled）
	CancelBill(ctx context.Context, billId string, userId, companyId uint64) error
	// AutoCancelStaleBills 扫描超时未完成的账单并取消退款（供定时任务调用）
	AutoCancelStaleBills(ctx context.Context) error
}
