package companyecoin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/muhaobing/std-go/go-common/database"

	"wdkr-marketplace-service/internal/common/constant/sys_err"
	"wdkr-marketplace-service/internal/domain/companyecoin/companyecoin_model"
	"wdkr-marketplace-service/internal/domain/companyecoin/repo"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
)

type companyEcoinServiceImpl struct {
	repo repo.CompanyEcoinRepo
}

// NewCompanyEcoinService 企业积分服务
func NewCompanyEcoinService(r repo.CompanyEcoinRepo) CompanyEcoinService {
	return &companyEcoinServiceImpl{repo: r}
}

func (s *companyEcoinServiceImpl) InitCompanyEcoin(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error) {
	if companyId == 0 {
		return nil, errors.New("company_id is required")
	}
	existing, err := s.repo.GetCompanyEcoin(ctx, companyId)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	row := &companyecoin_model.CompanyEcoin{CompanyId: companyId, AvailableStock: 0}
	if err := s.repo.CreateCompanyEcoin(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *companyEcoinServiceImpl) GetCompanyEcoin(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error) {
	if companyId == 0 {
		return nil, errors.New("company_id is required")
	}
	var out *companyecoin_model.CompanyEcoin
	now := uint32(time.Now().Unix())
	err := database.Transaction(ctx, func(txCtx context.Context) error {
		var err error
		out, err = s.getOrInitForUpdate(txCtx, companyId)
		if err != nil {
			return err
		}
		groups, err := s.repo.GetStockGroupsByCompanyIdForUpdate(txCtx, companyId)
		if err != nil {
			return err
		}
		groups, err = s.ensureLegacyStockGroup(txCtx, out, groups, now)
		if err != nil {
			return err
		}
		_, err = s.expireGroups(txCtx, out, groups, now, "auto_expire_read")
		if err != nil {
			return err
		}
		display, err := s.repo.GetStockGroupsByCompanyId(txCtx, companyId)
		if err != nil {
			return err
		}
		out.StockGroups = display
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *companyEcoinServiceImpl) AddCompanyEcoin(ctx context.Context, req *AddCompanyEcoinRequest) (*companyecoin_model.CompanyEcoinTransaction, error) {
	if req.CompanyId == 0 || req.OperatorUserId == 0 {
		return nil, errors.New("company_id and operator_user_id are required")
	}
	if req.Amount <= 0 || req.SourceType == "" {
		return nil, errors.New("invalid add request")
	}
	var tx *companyecoin_model.CompanyEcoinTransaction
	now := uint32(time.Now().Unix())
	err := database.Transaction(ctx, func(ctx context.Context) error {
		ce, err := s.getOrInitForUpdate(ctx, req.CompanyId)
		if err != nil {
			return err
		}
		groups, err := s.repo.GetStockGroupsByCompanyIdForUpdate(ctx, req.CompanyId)
		if err != nil {
			return err
		}
		groups, err = s.ensureLegacyStockGroup(ctx, ce, groups, now)
		if err != nil {
			return err
		}
		_, err = s.expireGroups(ctx, ce, groups, now, "auto_expire_add")
		if err != nil {
			return err
		}
		newStock := ce.AvailableStock + req.Amount
		g := &companyecoin_model.CompanyEcoinStockGroup{
			CompanyId:      req.CompanyId,
			TotalStock:     req.Amount,
			RemainingStock: req.Amount,
			ExpireTime:     ecoin.CalcEcoinExpireTimeFromUnix(now),
			SourceType:     req.SourceType,
			SourceId:       req.SourceId,
		}
		if err := s.repo.CreateStockGroup(ctx, g); err != nil {
			return err
		}
		tx = &companyecoin_model.CompanyEcoinTransaction{
			CompanyId:      req.CompanyId,
			OperatorUserId: req.OperatorUserId,
			Amount:         req.Amount,
			BeforeStock:    ce.AvailableStock,
			AfterStock:     newStock,
			TxType:         ecoin_model.TransactionTypeAdd,
			SourceType:     req.SourceType,
			SourceId:       req.SourceId,
			Description:    req.Description,
			Status:         ecoin_model.TransactionStatusCompleted,
		}
		if err := s.repo.AddTransaction(ctx, tx); err != nil {
			return err
		}
		return s.repo.UpdateCompanyEcoinStock(ctx, req.CompanyId, newStock)
	})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *companyEcoinServiceImpl) DeductCompanyEcoin(ctx context.Context, req *DeductCompanyEcoinRequest) (*companyecoin_model.CompanyEcoinTransaction, error) {
	if req.CompanyId == 0 || req.OperatorUserId == 0 {
		return nil, errors.New("company_id and operator_user_id are required")
	}
	if req.Amount <= 0 || req.SourceType == "" {
		return nil, errors.New("invalid deduct request")
	}
	var tx *companyecoin_model.CompanyEcoinTransaction
	now := uint32(time.Now().Unix())
	err := database.Transaction(ctx, func(ctx context.Context) error {
		ce, err := s.getOrInitForUpdate(ctx, req.CompanyId)
		if err != nil {
			return err
		}
		groups, err := s.repo.GetStockGroupsByCompanyIdForUpdate(ctx, req.CompanyId)
		if err != nil {
			return err
		}
		groups, err = s.ensureLegacyStockGroup(ctx, ce, groups, now)
		if err != nil {
			return err
		}
		_, err = s.expireGroups(ctx, ce, groups, now, "auto_expire_deduct")
		if err != nil {
			return err
		}
		availGroups, err := s.repo.GetAvailableStockGroupsForUpdate(ctx, req.CompanyId, now)
		if err != nil {
			return err
		}
		var sum float64
		for _, g := range availGroups {
			sum += g.RemainingStock
		}
		if sum < req.Amount {
			return sys_err.ErrInsufficientEcoin
		}
		need := req.Amount
		for _, g := range availGroups {
			if need <= 0 {
				break
			}
			use := g.RemainingStock
			if use > need {
				use = need
			}
			nr := g.RemainingStock - use
			if nr <= 0 {
				if err := s.repo.DeleteStockGroup(ctx, g.Id); err != nil {
					return err
				}
			} else {
				if err := s.repo.UpdateStockGroupRemaining(ctx, g.Id, nr); err != nil {
					return err
				}
			}
			need -= use
		}
		newStock := ce.AvailableStock - req.Amount
		tx = &companyecoin_model.CompanyEcoinTransaction{
			CompanyId:      req.CompanyId,
			OperatorUserId: req.OperatorUserId,
			Amount:         -req.Amount,
			BeforeStock:    ce.AvailableStock,
			AfterStock:     newStock,
			TxType:         ecoin_model.TransactionTypeDeduct,
			SourceType:     req.SourceType,
			SourceId:       req.SourceId,
			Description:    req.Description,
			Status:         ecoin_model.TransactionStatusCompleted,
		}
		if err := s.repo.AddTransaction(ctx, tx); err != nil {
			return err
		}
		return s.repo.UpdateCompanyEcoinStock(ctx, req.CompanyId, newStock)
	})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *companyEcoinServiceImpl) GetTransactionList(ctx context.Context, req *TransactionListRequest) (*TransactionListResponse, error) {
	if req.CompanyId == 0 {
		return nil, errors.New("company_id is required")
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	total, err := s.repo.CountTransactionsByCompanyId(ctx, req.CompanyId)
	if err != nil {
		return nil, err
	}
	list, err := s.repo.GetTransactionsByCompanyId(ctx, req.CompanyId, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}
	return &TransactionListResponse{Total: total, List: list}, nil
}

func (s *companyEcoinServiceImpl) ExpireCompanyEcoinStock(ctx context.Context, now uint32, limit int) (int, error) {
	if limit <= 0 {
		limit = 500
	}
	processed := 0
	err := database.Transaction(ctx, func(ctx context.Context) error {
		expired, err := s.repo.GetExpiredStockGroupsForUpdate(ctx, now, limit)
		if err != nil {
			return err
		}
		if len(expired) == 0 {
			return nil
		}
		byCompany := map[uint64]float64{}
		for _, g := range expired {
			if g.RemainingStock <= 0 {
				continue
			}
			if err := s.repo.DeleteStockGroup(ctx, g.Id); err != nil {
				return err
			}
			byCompany[g.CompanyId] += g.RemainingStock
			processed++
		}
		for cid, amt := range byCompany {
			ce, err := s.getOrInitForUpdate(ctx, cid)
			if err != nil {
				return err
			}
			newStock := ce.AvailableStock - amt
			if newStock < 0 {
				newStock = 0
			}
			if err := s.repo.UpdateCompanyEcoinStock(ctx, cid, newStock); err != nil {
				return err
			}
			tx := &companyecoin_model.CompanyEcoinTransaction{
				CompanyId:      cid,
				OperatorUserId: 0,
				Amount:         -amt,
				BeforeStock:    ce.AvailableStock,
				AfterStock:     newStock,
				TxType:         ecoin_model.TransactionTypeDeduct,
				SourceType:     ecoin_model.SourceTypeExpire,
				SourceId:       fmt.Sprintf("expire_task:%d", now),
				Description:    "积分已过期自动失效",
				Status:         ecoin_model.TransactionStatusCompleted,
			}
			if err := s.repo.AddTransaction(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	})
	return processed, err
}

func (s *companyEcoinServiceImpl) getOrInitForUpdate(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error) {
	ce, err := s.repo.GetCompanyEcoinForUpdate(ctx, companyId)
	if err != nil {
		return nil, err
	}
	if ce != nil {
		return ce, nil
	}
	ce = &companyecoin_model.CompanyEcoin{CompanyId: companyId, AvailableStock: 0}
	if err := s.repo.CreateCompanyEcoin(ctx, ce); err != nil {
		return nil, err
	}
	return ce, nil
}

func (s *companyEcoinServiceImpl) ensureLegacyStockGroup(ctx context.Context, ce *companyecoin_model.CompanyEcoin, groups []*companyecoin_model.CompanyEcoinStockGroup, now uint32) ([]*companyecoin_model.CompanyEcoinStockGroup, error) {
	if ce.AvailableStock <= 0 || len(groups) > 0 {
		return groups, nil
	}
	g := &companyecoin_model.CompanyEcoinStockGroup{
		CompanyId:      ce.CompanyId,
		TotalStock:     ce.AvailableStock,
		RemainingStock: ce.AvailableStock,
		ExpireTime:     ecoin.CalcEcoinExpireTimeFromUnix(now),
		SourceType:     ecoin_model.SourceTypeSystem,
		SourceId:       "legacy_migration",
	}
	if err := s.repo.CreateStockGroup(ctx, g); err != nil {
		return nil, err
	}
	return append(groups, g), nil
}

func (s *companyEcoinServiceImpl) expireGroups(ctx context.Context, ce *companyecoin_model.CompanyEcoin, groups []*companyecoin_model.CompanyEcoinStockGroup, now uint32, sourceID string) (float64, error) {
	var expiredAmount float64
	for _, g := range groups {
		if g.RemainingStock <= 0 {
			continue
		}
		if g.ExpireTime > 0 && g.ExpireTime <= now {
			expiredAmount += g.RemainingStock
			if err := s.repo.DeleteStockGroup(ctx, g.Id); err != nil {
				return 0, err
			}
		}
	}
	if expiredAmount <= 0 {
		return 0, nil
	}
	newStock := ce.AvailableStock - expiredAmount
	if newStock < 0 {
		newStock = 0
	}
	if err := s.repo.UpdateCompanyEcoinStock(ctx, ce.CompanyId, newStock); err != nil {
		return 0, err
	}
	tx := &companyecoin_model.CompanyEcoinTransaction{
		CompanyId:      ce.CompanyId,
		OperatorUserId: 0,
		Amount:         -expiredAmount,
		BeforeStock:    ce.AvailableStock,
		AfterStock:     newStock,
		TxType:         ecoin_model.TransactionTypeDeduct,
		SourceType:     ecoin_model.SourceTypeExpire,
		SourceId:       sourceID,
		Description:    "积分已过期自动失效",
		Status:         ecoin_model.TransactionStatusCompleted,
	}
	if err := s.repo.AddTransaction(ctx, tx); err != nil {
		return 0, err
	}
	ce.AvailableStock = newStock
	return expiredAmount, nil
}
