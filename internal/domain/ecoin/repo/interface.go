package repo

import (
	"context"

	"wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
)

// EcoinRepo 积分仓储接口
type EcoinRepo interface {
	// GetUserEcoin 获取用户积分信息
	GetUserEcoin(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error)

	// CreateUserEcoin 创建用户积分记录
	CreateUserEcoin(ctx context.Context, userEcoin *ecoin_model.UserEcoin) error

	// UpdateUserEcoinStock 更新用户积分余额
	UpdateUserEcoinStock(ctx context.Context, userId uint64, stock float64) error

	// AddEcoinTransaction 新增积分流水记录
	AddEcoinTransaction(ctx context.Context, transaction *ecoin_model.EcoinTransaction) error

	// GetEcoinTransactionsByUserId 根据用户ID获取积分流水记录
	GetEcoinTransactionsByUserId(ctx context.Context, userId uint64, offset, limit int) ([]*ecoin_model.EcoinTransaction, error)

	// GetEcoinTransactionById 根据ID获取积分流水记录
	GetEcoinTransactionById(ctx context.Context, id uint64) (*ecoin_model.EcoinTransaction, error)

	// UpdateEcoinTransactionStatus 更新积分流水状态
	UpdateEcoinTransactionStatus(ctx context.Context, id uint64, status int) error

	// GetUserEcoinForUpdate 获取用户积分信息（加锁）
	GetUserEcoinForUpdate(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error)
}
