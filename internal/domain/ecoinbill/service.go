package ecoinbill

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/muhaobing/std-go/go-common/database"

	"wdkr-marketplace-service/internal/common/constant/sys_err"
	"wdkr-marketplace-service/internal/common/utils/bill_id_gen"
	"wdkr-marketplace-service/internal/domain/companyecoin"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
	"wdkr-marketplace-service/internal/domain/ecoinbill/ecoinbill_model"
	"wdkr-marketplace-service/internal/domain/ecoinbill/repo"
)

const (
	autoCancelScanLimit = 200
	billTimeout         = 15 * time.Minute
)

type pointsBillServiceImpl struct {
	repo         repo.EcoinBillRepo
	ecoin        ecoin.EcoinService
	companyEcoin companyecoin.CompanyEcoinService
}

// NewEcoinBillService 创建金币账单服务
func NewEcoinBillService(
	billRepo repo.EcoinBillRepo,
	ecoinSvc ecoin.EcoinService,
	companyEcoinSvc companyecoin.CompanyEcoinService,
) EcoinBillService {
	return &pointsBillServiceImpl{
		repo:         billRepo,
		ecoin:        ecoinSvc,
		companyEcoin: companyEcoinSvc,
	}
}

func (s *pointsBillServiceImpl) PreDeduct(ctx context.Context, userId, companyId uint64, cost float64) (*ecoinbill_model.EcoinBill, error) {
	if userId == 0 || cost <= 0 {
		return nil, errors.New("invalid pre_deduct request")
	}
	billID, err := bill_id_gen.GenerateBillID(userId)
	if err != nil {
		return nil, err
	}
	var outBillId string
	err = database.Transaction(ctx, func(ctx context.Context) error {
		bill := &ecoinbill_model.EcoinBill{
			BillId:    billID,
			UserId:    userId,
			CompanyId: companyId,
			Amount:    cost,
			Status:    ecoinbill_model.StatusIncomplete,
		}
		if err := s.repo.Create(ctx, bill); err != nil {
			return err
		}
		outBillId = bill.BillId
		sid := bill.BillId
		if companyId > 0 {
			tx, err := s.companyEcoin.DeductCompanyEcoinInTx(ctx, &companyecoin.DeductCompanyEcoinRequest{
				CompanyId:      companyId,
				OperatorUserId: userId,
				Amount:         cost,
				SourceType:     ecoin_model.SourceTypeEcoinBillPreDeduct,
				SourceId:       sid,
				Description:    "金币账单预扣",
			})
			if err != nil {
				return err
			}
			return s.repo.UpdateDeductTxIds(ctx, bill.BillId, 0, tx.Id)
		}
		tx, err := s.ecoin.DeductEcoinInTx(ctx, &ecoin.DeductEcoinRequest{
			UserId:      userId,
			Amount:      cost,
			SourceType:  ecoin_model.SourceTypeEcoinBillPreDeduct,
			SourceId:    sid,
			Description: "金币账单预扣",
		})
		if err != nil {
			return err
		}
		return s.repo.UpdateDeductTxIds(ctx, bill.BillId, tx.Id, 0)
	})
	if err != nil {
		if errors.Is(err, sys_err.ErrInsufficientEcoin) {
			return nil, err
		}
		return nil, err
	}
	return s.repo.GetByBillId(ctx, outBillId)
}

func (s *pointsBillServiceImpl) ConfirmBill(ctx context.Context, billId string, userId, companyId uint64) error {
	return database.Transaction(ctx, func(ctx context.Context) error {
		bill, err := s.repo.GetByBillIdForUpdate(ctx, billId)
		if err != nil {
			return err
		}
		if bill == nil || bill.UserId != userId || bill.CompanyId != companyId {
			return sys_err.ErrEcoinBillNotFound
		}
		if !bill.IsIncomplete() {
			return sys_err.ErrEcoinBillInvalidState
		}
		return s.repo.UpdateStatus(ctx, billId, ecoinbill_model.StatusCompleted)
	})
}

func (s *pointsBillServiceImpl) CancelBill(ctx context.Context, billId string, userId, companyId uint64) error {
	return database.Transaction(ctx, func(ctx context.Context) error {
		bill, err := s.repo.GetByBillIdForUpdate(ctx, billId)
		if err != nil {
			return err
		}
		if bill == nil || bill.UserId != userId || bill.CompanyId != companyId {
			return sys_err.ErrEcoinBillNotFound
		}
		return s.refundAndCancel(ctx, bill)
	})
}

func (s *pointsBillServiceImpl) refundAndCancel(ctx context.Context, bill *ecoinbill_model.EcoinBill) error {
	if !bill.IsIncomplete() {
		return sys_err.ErrEcoinBillInvalidState
	}
	rid := fmt.Sprintf("ecoin_bill_refund:%s", bill.BillId)
	if bill.CompanyId > 0 {
		_, err := s.companyEcoin.AddCompanyEcoinInTx(ctx, &companyecoin.AddCompanyEcoinRequest{
			CompanyId:      bill.CompanyId,
			OperatorUserId: bill.UserId,
			Amount:         bill.Amount,
			SourceType:     ecoin_model.SourceTypeEcoinBillRefund,
			SourceId:       rid,
			Description:    "金币账单取消退回",
		})
		if err != nil {
			return err
		}
	} else {
		_, err := s.ecoin.AddEcoinInTx(ctx, &ecoin.AddEcoinRequest{
			UserId:      bill.UserId,
			Amount:      bill.Amount,
			SourceType:  ecoin_model.SourceTypeEcoinBillRefund,
			SourceId:    rid,
			Description: "金币账单取消退回",
		})
		if err != nil {
			return err
		}
	}
	return s.repo.UpdateStatus(ctx, bill.BillId, ecoinbill_model.StatusCancelled)
}

func (s *pointsBillServiceImpl) AutoCancelStaleBills(ctx context.Context) error {
	cutoff := uint32(time.Now().Add(-billTimeout).Unix())
	for shard := 0; shard < ecoinbill_model.ShardCount; shard++ {
		bills, err := s.repo.ListIncompleteOlderThan(ctx, shard, cutoff, autoCancelScanLimit)
		if err != nil {
			return err
		}
		for _, b := range bills {
			err := database.Transaction(ctx, func(ctx context.Context) error {
				bill, err := s.repo.GetByBillIdForUpdate(ctx, b.BillId)
				if err != nil {
					return err
				}
				if bill == nil || !bill.IsIncomplete() {
					return nil
				}
				if bill.Ctime >= cutoff {
					return nil
				}
				return s.refundAndCancel(ctx, bill)
			})
			if err != nil {
				fmt.Printf("[EcoinBillAutoCancel] bill_id=%s err=%v\n", b.BillId, err)
			}
		}
	}
	return nil
}
