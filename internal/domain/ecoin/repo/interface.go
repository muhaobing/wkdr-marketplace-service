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

	// CountEcoinTransactionsByUserId 统计用户积分流水数量
	CountEcoinTransactionsByUserId(ctx context.Context, userId uint64) (int64, error)

	// GetEcoinTransactionById 根据ID获取积分流水记录
	GetEcoinTransactionById(ctx context.Context, id uint64) (*ecoin_model.EcoinTransaction, error)

	// UpdateEcoinTransactionStatus 更新积分流水状态
	UpdateEcoinTransactionStatus(ctx context.Context, id uint64, status int) error

	// GetUserEcoinForUpdate 获取用户积分信息（加锁）
	GetUserEcoinForUpdate(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error)

	// CreateUserEcoinStockGroup 新增积分库存分组
	CreateUserEcoinStockGroup(ctx context.Context, group *ecoin_model.UserEcoinStockGroup) error

	// UpdateUserEcoinStockGroupRemaining 更新积分库存分组剩余数量
	UpdateUserEcoinStockGroupRemaining(ctx context.Context, groupId uint64, remainingStock float64) error

	// DeleteUserEcoinStockGroup 删除积分库存分组
	DeleteUserEcoinStockGroup(ctx context.Context, groupId uint64) error

	// GetAvailableStockGroupsForUpdate 获取可用积分分组（按最早过期优先，0=永不过期排最后）
	GetAvailableStockGroupsForUpdate(ctx context.Context, userId uint64, now uint32) ([]*ecoin_model.UserEcoinStockGroup, error)

	// GetStockGroupsByUserId 获取用户积分分组（用于展示）
	GetStockGroupsByUserId(ctx context.Context, userId uint64) ([]*ecoin_model.UserEcoinStockGroup, error)

	// GetStockGroupsByUserIdForUpdate 获取用户积分分组（加锁）
	GetStockGroupsByUserIdForUpdate(ctx context.Context, userId uint64) ([]*ecoin_model.UserEcoinStockGroup, error)

	// GetExpiredStockGroupsForUpdate 获取已过期且仍有余额的积分分组（加锁）
	GetExpiredStockGroupsForUpdate(ctx context.Context, now uint32, limit int) ([]*ecoin_model.UserEcoinStockGroup, error)
}
