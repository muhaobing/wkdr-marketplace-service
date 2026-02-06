package ecoin

import (
	"context"
	"errors"
	"fmt"

	"github.com/muhaobing-eng/std-go/go-common/database"

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

	userEcoin, err := s.ecoinRepo.GetUserEcoin(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ecoin: %w", err)
	}

	// 如果用户积分记录不存在，自动初始化
	if userEcoin == nil {
		return s.InitUserEcoin(ctx, userId)
	}

	return userEcoin, nil
}

// AddEcoin 增加积分
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

	var transaction *ecoin_model.EcoinTransaction
	err := database.Transaction(ctx, func(ctx context.Context) error {
		// 获取用户当前积分（加锁）
		userEcoin, err := s.ecoinRepo.GetUserEcoinForUpdate(ctx, req.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user ecoin for update: %w", err)
		}

		// 如果用户积分记录不存在，先创建
		if userEcoin == nil {
			userEcoin = &ecoin_model.UserEcoin{
				UserId:         req.UserId,
				AvailableStock: 0,
			}
			if err = s.ecoinRepo.CreateUserEcoin(ctx, userEcoin); err != nil {
				return fmt.Errorf("failed to create user ecoin: %w", err)
			}
		}

		// 计算新的积分余额
		newStock := userEcoin.AvailableStock + req.Amount

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

// DeductEcoin 扣除积分
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

	var transaction *ecoin_model.EcoinTransaction
	err := database.Transaction(ctx, func(ctx context.Context) error {
		// 获取用户当前积分（加锁）
		userEcoin, err := s.ecoinRepo.GetUserEcoinForUpdate(ctx, req.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user ecoin for update: %w", err)
		}
		// 检查用户积分记录是否存在
		if userEcoin == nil {
			return errors.New("user ecoin record not found")
		}

		// 检查积分余额是否足够
		if userEcoin.AvailableStock < req.Amount {
			return errors.New("insufficient ecoin balance")
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
