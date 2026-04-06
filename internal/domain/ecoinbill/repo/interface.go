package repo

import (
	"context"

	"wdkr-marketplace-service/internal/domain/ecoinbill/ecoinbill_model"
)

// EcoinBillRepo 积分账单仓储
type EcoinBillRepo interface {
	Create(ctx context.Context, bill *ecoinbill_model.EcoinBill) error
	GetByBillId(ctx context.Context, billId string) (*ecoinbill_model.EcoinBill, error)
	GetByBillIdForUpdate(ctx context.Context, billId string) (*ecoinbill_model.EcoinBill, error)
	UpdateDeductTxIds(ctx context.Context, billId string, userTxId, companyTxId uint64) error
	UpdateStatus(ctx context.Context, billId string, status uint8) error
	ListIncompleteOlderThan(ctx context.Context, shard int, beforeCtime uint32, limit int) ([]*ecoinbill_model.EcoinBill, error)
}
