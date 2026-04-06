package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	"wdkr-marketplace-service/internal/common/utils/bill_id_gen"
	"wdkr-marketplace-service/internal/domain/ecoinbill/ecoinbill_model"
)

type pointsBillRepoImpl struct{}

// NewEcoinBillRepo 创建积分账单仓储
func NewEcoinBillRepo() EcoinBillRepo {
	return &pointsBillRepoImpl{}
}

func (r *pointsBillRepoImpl) Create(ctx context.Context, bill *ecoinbill_model.EcoinBill) error {
	tbl := ecoinbill_model.TableNameForUser(bill.UserId)
	return database.FromContext(ctx).Table(tbl).Create(bill).Error
}

func (r *pointsBillRepoImpl) GetByBillId(ctx context.Context, billId string) (*ecoinbill_model.EcoinBill, error) {
	shard, err := bill_id_gen.ShardIndexFromBillID(billId)
	if err != nil {
		return nil, nil
	}
	var row ecoinbill_model.EcoinBill
	tbl := ecoinbill_model.TableNameForShard(shard)
	err = database.FromContext(ctx).Table(tbl).Where("bill_id = ?", billId).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pointsBillRepoImpl) GetByBillIdForUpdate(ctx context.Context, billId string) (*ecoinbill_model.EcoinBill, error) {
	shard, err := bill_id_gen.ShardIndexFromBillID(billId)
	if err != nil {
		return nil, nil
	}
	var row ecoinbill_model.EcoinBill
	tbl := ecoinbill_model.TableNameForShard(shard)
	err = database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").Table(tbl).Where("bill_id = ?", billId).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pointsBillRepoImpl) UpdateDeductTxIds(ctx context.Context, billId string, userTxId, companyTxId uint64) error {
	shard, err := bill_id_gen.ShardIndexFromBillID(billId)
	if err != nil {
		return err
	}
	tbl := ecoinbill_model.TableNameForShard(shard)
	return database.FromContext(ctx).Table(tbl).
		Where("bill_id = ?", billId).
		Updates(map[string]interface{}{
			"deduct_user_tx_id":    userTxId,
			"deduct_company_tx_id": companyTxId,
		}).Error
}

func (r *pointsBillRepoImpl) UpdateStatus(ctx context.Context, billId string, status uint8) error {
	shard, err := bill_id_gen.ShardIndexFromBillID(billId)
	if err != nil {
		return err
	}
	tbl := ecoinbill_model.TableNameForShard(shard)
	return database.FromContext(ctx).Table(tbl).
		Where("bill_id = ?", billId).
		Update("status", status).Error
}

func (r *pointsBillRepoImpl) ListIncompleteOlderThan(ctx context.Context, shard int, beforeCtime uint32, limit int) ([]*ecoinbill_model.EcoinBill, error) {
	if shard < 0 || shard >= ecoinbill_model.ShardCount {
		return nil, fmt.Errorf("invalid shard index %d", shard)
	}
	if limit <= 0 {
		limit = 200
	}
	tbl := ecoinbill_model.TableNameForShard(shard)
	var rows []*ecoinbill_model.EcoinBill
	err := database.FromContext(ctx).Table(tbl).
		Where("status = ? AND ctime < ?", ecoinbill_model.StatusIncomplete, beforeCtime).
		Order("ctime ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
