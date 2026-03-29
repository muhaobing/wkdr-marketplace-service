package ecoin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/muhaobing/std-go/go-common/database"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/constant/sys_err"
	"wdkr-marketplace-service/internal/common/utils"
	"wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
	"wdkr-marketplace-service/internal/domain/ecoin/repo"
)

type ecoinServiceImpl struct {
	ecoinRepo repo.EcoinRepo
}

// NewEcoinService 创建积分服务实例
func NewEcoinService(ecoinRepo repo.EcoinRepo) EcoinService {
	return &ecoinServiceImpl{
		ecoinRepo: ecoinRepo,
	}
}

// GetUserEcoin 获取用户积分信息
func (s *ecoinServiceImpl) GetUserEcoin(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error) {
	if userId == 0 {
		return nil, errors.New("user id is required")
	}

	var userEcoin *ecoin_model.UserEcoin
	now := uint32(time.Now().Unix())
	err := database.Transaction(ctx, func(txCtx context.Context) error {
		var err error
		userEcoin, err = s.getOrInitUserEcoinForUpdate(txCtx, userId)
		if err != nil {
			return err
		}
		groups, err := s.ecoinRepo.GetStockGroupsByUserIdForUpdate(txCtx, userId)
		if err != nil {
			return fmt.Errorf("failed to get stock groups: %w", err)
		}
		groups, err = s.ensureLegacyStockGroup(txCtx, userEcoin, groups, now)
		if err != nil {
			return err
		}
		_, err = s.expireGroupsForUser(txCtx, userEcoin, groups, now, "auto_expire_read")
		if err != nil {
			return err
		}
		displayGroups, err := s.ecoinRepo.GetStockGroupsByUserId(txCtx, userId)
		if err != nil {
			return fmt.Errorf("failed to list stock groups: %w", err)
		}
		userEcoin.StockGroups = displayGroups
		return nil
	})
	if err != nil {
		return nil, err
	}

	return userEcoin, nil
}

// AddEcoin 增加积分（source_id 非空时 Redis SETNX 幂等键仅含 source_type+source_id；未抢到键视为已成功并返回 nil,nil；TTL 见 ecoin_idempotency_ttl_seconds）
func (s *ecoinServiceImpl) AddEcoin(ctx context.Context, req *AddEcoinRequest) (*ecoin_model.EcoinTransaction, error) {
	if req.UserId == 0 {
		return nil, errors.New("user id is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if req.SourceType == "" {
		return nil, errors.New("source type is required")
	}

	if strings.TrimSpace(req.SourceId) == "" {
		return s.addEcoinOnce(ctx, req)
	}

	key := utils.EcoinIdempotencyRedisKey(req.SourceType, req.SourceId)
	ttl := config.GetEcoinIdempotencyTTL()
	acquired, err := utils.TryAcquireIdempotencyKey(ctx, key, ttl)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, nil
	}
	tx, err := s.addEcoinOnce(ctx, req)
	if err != nil {
		utils.ReleaseIdempotencyKey(ctx, key)
		return nil, err
	}
	return tx, nil
}

func (s *ecoinServiceImpl) addEcoinOnce(ctx context.Context, req *AddEcoinRequest) (*ecoin_model.EcoinTransaction, error) {
	var transaction *ecoin_model.EcoinTransaction
	now := uint32(time.Now().Unix())
	err := database.Transaction(ctx, func(ctx context.Context) error {
		userEcoin, err := s.getOrInitUserEcoinForUpdate(ctx, req.UserId)
		if err != nil {
			return err
		}
		groups, err := s.ecoinRepo.GetStockGroupsByUserIdForUpdate(ctx, req.UserId)
		if err != nil {
			return fmt.Errorf("failed to get stock groups: %w", err)
		}
		groups, err = s.ensureLegacyStockGroup(ctx, userEcoin, groups, now)
		if err != nil {
			return err
		}
		_, err = s.expireGroupsForUser(ctx, userEcoin, groups, now, "auto_expire_add")
		if err != nil {
			return err
		}

		// 计算新的积分余额
		newStock := userEcoin.AvailableStock + req.Amount

		// 入账分组
		group := &ecoin_model.UserEcoinStockGroup{
			UserId:         req.UserId,
			TotalStock:     req.Amount,
			RemainingStock: req.Amount,
			ExpireTime:     s.calcExpireTime(now),
			SourceType:     req.SourceType,
			SourceId:       req.SourceId,
		}
		if err = s.ecoinRepo.CreateUserEcoinStockGroup(ctx, group); err != nil {
			return fmt.Errorf("failed to create stock group: %w", err)
		}

		// 创建积分流水记录
		transaction = &ecoin_model.EcoinTransaction{
			UserId:      req.UserId,
			Amount:      req.Amount,
			BeforeStock: userEcoin.AvailableStock,
			AfterStock:  newStock,
			TxType:      ecoin_model.TransactionTypeAdd,
			SourceType:  req.SourceType,
			SourceId:    req.SourceId,
			Description: req.Description,
			Status:      ecoin_model.TransactionStatusCompleted,
		}
		if err = s.ecoinRepo.AddEcoinTransaction(ctx, transaction); err != nil {
			return fmt.Errorf("failed to add ecoin transaction: %w", err)
		}

		// 更新用户积分余额
		if err = s.ecoinRepo.UpdateUserEcoinStock(ctx, req.UserId, newStock); err != nil {
			return fmt.Errorf("failed to update user ecoin stock: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// DeductEcoin 扣除积分（幂等键与 AddEcoin 相同规则，同一 source_type+source_id 在 TTL 内先执行的接口会占用键）
func (s *ecoinServiceImpl) DeductEcoin(ctx context.Context, req *DeductEcoinRequest) (*ecoin_model.EcoinTransaction, error) {
	if req.UserId == 0 {
		return nil, errors.New("user id is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if req.SourceType == "" {
		return nil, errors.New("source type is required")
	}

	if strings.TrimSpace(req.SourceId) == "" {
		return s.deductEcoinOnce(ctx, req)
	}

	key := utils.EcoinIdempotencyRedisKey(req.SourceType, req.SourceId)
	ttl := config.GetEcoinIdempotencyTTL()
	acquired, err := utils.TryAcquireIdempotencyKey(ctx, key, ttl)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, nil
	}
	tx, err := s.deductEcoinOnce(ctx, req)
	if err != nil {
		utils.ReleaseIdempotencyKey(ctx, key)
		return nil, err
	}
	return tx, nil
}

func (s *ecoinServiceImpl) deductEcoinOnce(ctx context.Context, req *DeductEcoinRequest) (*ecoin_model.EcoinTransaction, error) {
	var transaction *ecoin_model.EcoinTransaction
	now := uint32(time.Now().Unix())
	err := database.Transaction(ctx, func(ctx context.Context) error {
		userEcoin, err := s.getOrInitUserEcoinForUpdate(ctx, req.UserId)
		if err != nil {
			return err
		}
		groups, err := s.ecoinRepo.GetStockGroupsByUserIdForUpdate(ctx, req.UserId)
		if err != nil {
			return fmt.Errorf("failed to get stock groups: %w", err)
		}
		groups, err = s.ensureLegacyStockGroup(ctx, userEcoin, groups, now)
		if err != nil {
			return err
		}
		_, err = s.expireGroupsForUser(ctx, userEcoin, groups, now, "auto_expire_deduct")
		if err != nil {
			return err
		}

		availableGroups, err := s.ecoinRepo.GetAvailableStockGroupsForUpdate(ctx, req.UserId, now)
		if err != nil {
			return fmt.Errorf("failed to get available stock groups: %w", err)
		}
		availableStock := 0.0
		for _, g := range availableGroups {
			availableStock += g.RemainingStock
		}
		if availableStock < req.Amount {
			return sys_err.ErrInsufficientEcoin
		}

		// 按最早过期优先扣减
		need := req.Amount
		for _, g := range availableGroups {
			if need <= 0 {
				break
			}
			use := g.RemainingStock
			if use > need {
				use = need
			}
			newRemaining := g.RemainingStock - use
			if newRemaining <= 0 {
				if err = s.ecoinRepo.DeleteUserEcoinStockGroup(ctx, g.Id); err != nil {
					return fmt.Errorf("failed to delete stock group %d: %w", g.Id, err)
				}
			} else {
				if err = s.ecoinRepo.UpdateUserEcoinStockGroupRemaining(ctx, g.Id, newRemaining); err != nil {
					return fmt.Errorf("failed to update stock group %d: %w", g.Id, err)
				}
			}
			need -= use
		}

		// 计算新的积分余额
		newStock := userEcoin.AvailableStock - req.Amount

		// 创建积分流水记录
		transaction = &ecoin_model.EcoinTransaction{
			UserId:      req.UserId,
			Amount:      -req.Amount, // 扣除积分，使用负数
			BeforeStock: userEcoin.AvailableStock,
			AfterStock:  newStock,
			TxType:      ecoin_model.TransactionTypeDeduct,
			SourceType:  req.SourceType,
			SourceId:    req.SourceId,
			Description: req.Description,
			Status:      ecoin_model.TransactionStatusCompleted,
		}
		if err = s.ecoinRepo.AddEcoinTransaction(ctx, transaction); err != nil {
			return fmt.Errorf("failed to add ecoin transaction: %w", err)
		}

		// 更新用户积分余额
		if err = s.ecoinRepo.UpdateUserEcoinStock(ctx, req.UserId, newStock); err != nil {
			return fmt.Errorf("failed to update user ecoin stock: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetEcoinTransactionList 获取积分流水列表
func (s *ecoinServiceImpl) GetEcoinTransactionList(ctx context.Context, req *EcoinTransactionListRequest) (*EcoinTransactionListResponse, error) {
	if req.UserId == 0 {
		return nil, errors.New("user id is required")
	}

	// 设置默认分页参数
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 获取总数
	total, err := s.ecoinRepo.CountEcoinTransactionsByUserId(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to count ecoin transactions: %w", err)
	}

	// 获取列表
	transactions, err := s.ecoinRepo.GetEcoinTransactionsByUserId(ctx, req.UserId, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get ecoin transactions: %w", err)
	}

	return &EcoinTransactionListResponse{
		Total: total,
		List:  transactions,
	}, nil
}

// GetEcoinTransaction 根据ID获取积分流水
func (s *ecoinServiceImpl) GetEcoinTransaction(ctx context.Context, id uint64) (*ecoin_model.EcoinTransaction, error) {
	if id == 0 {
		return nil, errors.New("transaction id is required")
	}

	transaction, err := s.ecoinRepo.GetEcoinTransactionById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ecoin transaction: %w", err)
	}

	if transaction == nil {
		return nil, errors.New("transaction not found")
	}

	return transaction, nil
}

// InitUserEcoin 初始化用户积分账户
func (s *ecoinServiceImpl) InitUserEcoin(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error) {
	if userId == 0 {
		return nil, errors.New("user id is required")
	}

	// 检查是否已存在
	existingEcoin, err := s.ecoinRepo.GetUserEcoin(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user ecoin: %w", err)
	}

	if existingEcoin != nil {
		return existingEcoin, nil
	}

	// 创建新的用户积分记录
	userEcoin := &ecoin_model.UserEcoin{
		UserId:         userId,
		AvailableStock: 0,
	}

	if err := s.ecoinRepo.CreateUserEcoin(ctx, userEcoin); err != nil {
		return nil, fmt.Errorf("failed to create user ecoin: %w", err)
	}

	return userEcoin, nil
}

// GetEcoinStockGroupList 获取积分库存分组
func (s *ecoinServiceImpl) GetEcoinStockGroupList(ctx context.Context, req *EcoinStockGroupListRequest) (*EcoinStockGroupListResponse, error) {
	if req.UserId == 0 {
		return nil, errors.New("user id is required")
	}
	ecoinInfo, err := s.GetUserEcoin(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &EcoinStockGroupListResponse{
		TotalStock: ecoinInfo.AvailableStock,
		List:       ecoinInfo.StockGroups,
	}, nil
}

// ExpireEcoinStock 过期积分处理（定时任务调用）
func (s *ecoinServiceImpl) ExpireEcoinStock(ctx context.Context, now uint32, limit int) (int, error) {
	if limit <= 0 {
		limit = 500
	}
	processed := 0
	err := database.Transaction(ctx, func(ctx context.Context) error {
		expiredGroups, err := s.ecoinRepo.GetExpiredStockGroupsForUpdate(ctx, now, limit)
		if err != nil {
			return fmt.Errorf("failed to query expired groups: %w", err)
		}
		if len(expiredGroups) == 0 {
			return nil
		}

		userExpired := map[uint64]float64{}
		for _, g := range expiredGroups {
			if g.RemainingStock <= 0 {
				continue
			}
			if err := s.ecoinRepo.DeleteUserEcoinStockGroup(ctx, g.Id); err != nil {
				return fmt.Errorf("failed to expire group %d: %w", g.Id, err)
			}
			userExpired[g.UserId] += g.RemainingStock
			processed++
		}

		for userId, amount := range userExpired {
			userEcoin, err := s.getOrInitUserEcoinForUpdate(ctx, userId)
			if err != nil {
				return err
			}
			newStock := userEcoin.AvailableStock - amount
			if newStock < 0 {
				newStock = 0
			}
			if err := s.ecoinRepo.UpdateUserEcoinStock(ctx, userId, newStock); err != nil {
				return fmt.Errorf("failed to update user stock: %w", err)
			}
			tx := &ecoin_model.EcoinTransaction{
				UserId:      userId,
				Amount:      -amount,
				BeforeStock: userEcoin.AvailableStock,
				AfterStock:  newStock,
				TxType:      ecoin_model.TransactionTypeDeduct,
				SourceType:  ecoin_model.SourceTypeExpire,
				SourceId:    fmt.Sprintf("expire_task:%d", now),
				Description: "积分已过期自动失效",
				Status:      ecoin_model.TransactionStatusCompleted,
			}
			if err := s.ecoinRepo.AddEcoinTransaction(ctx, tx); err != nil {
				return fmt.Errorf("failed to add expire transaction: %w", err)
			}
		}
		return nil
	})
	return processed, err
}

func (s *ecoinServiceImpl) getOrInitUserEcoinForUpdate(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error) {
	userEcoin, err := s.ecoinRepo.GetUserEcoinForUpdate(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ecoin for update: %w", err)
	}
	if userEcoin != nil {
		return userEcoin, nil
	}
	userEcoin = &ecoin_model.UserEcoin{
		UserId:         userId,
		AvailableStock: 0,
	}
	if err = s.ecoinRepo.CreateUserEcoin(ctx, userEcoin); err != nil {
		return nil, fmt.Errorf("failed to create user ecoin: %w", err)
	}
	return userEcoin, nil
}

func (s *ecoinServiceImpl) calcExpireTime(now uint32) uint32 {
	return CalcEcoinExpireTimeFromUnix(now)
}

func (s *ecoinServiceImpl) ensureLegacyStockGroup(ctx context.Context, userEcoin *ecoin_model.UserEcoin, groups []*ecoin_model.UserEcoinStockGroup, now uint32) ([]*ecoin_model.UserEcoinStockGroup, error) {
	if userEcoin.AvailableStock <= 0 || len(groups) > 0 {
		return groups, nil
	}
	legacy := &ecoin_model.UserEcoinStockGroup{
		UserId:         userEcoin.UserId,
		TotalStock:     userEcoin.AvailableStock,
		RemainingStock: userEcoin.AvailableStock,
		ExpireTime:     s.calcExpireTime(now),
		SourceType:     ecoin_model.SourceTypeSystem,
		SourceId:       "legacy_migration",
	}
	if err := s.ecoinRepo.CreateUserEcoinStockGroup(ctx, legacy); err != nil {
		return nil, fmt.Errorf("failed to create legacy stock group: %w", err)
	}
	return append(groups, legacy), nil
}

func (s *ecoinServiceImpl) expireGroupsForUser(ctx context.Context, userEcoin *ecoin_model.UserEcoin, groups []*ecoin_model.UserEcoinStockGroup, now uint32, sourceID string) (float64, error) {
	expiredAmount := 0.0
	for _, g := range groups {
		if g.RemainingStock <= 0 {
			continue
		}
		if g.ExpireTime > 0 && g.ExpireTime <= now {
			expiredAmount += g.RemainingStock
			if err := s.ecoinRepo.DeleteUserEcoinStockGroup(ctx, g.Id); err != nil {
				return 0, fmt.Errorf("failed to expire stock group %d: %w", g.Id, err)
			}
		}
	}
	if expiredAmount <= 0 {
		return 0, nil
	}
	newStock := userEcoin.AvailableStock - expiredAmount
	if newStock < 0 {
		newStock = 0
	}
	if err := s.ecoinRepo.UpdateUserEcoinStock(ctx, userEcoin.UserId, newStock); err != nil {
		return 0, fmt.Errorf("failed to update user ecoin stock: %w", err)
	}
	tx := &ecoin_model.EcoinTransaction{
		UserId:      userEcoin.UserId,
		Amount:      -expiredAmount,
		BeforeStock: userEcoin.AvailableStock,
		AfterStock:  newStock,
		TxType:      ecoin_model.TransactionTypeDeduct,
		SourceType:  ecoin_model.SourceTypeExpire,
		SourceId:    sourceID,
		Description: "积分已过期自动失效",
		Status:      ecoin_model.TransactionStatusCompleted,
	}
	if err := s.ecoinRepo.AddEcoinTransaction(ctx, tx); err != nil {
		return 0, fmt.Errorf("failed to add expire transaction: %w", err)
	}
	userEcoin.AvailableStock = newStock
	return expiredAmount, nil
}
