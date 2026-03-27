package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	"wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
)

type ecoinRepoImpl struct{}

// NewEcoinRepo 创建积分仓储实现
func NewEcoinRepo() EcoinRepo {
	return &ecoinRepoImpl{}
}

// GetUserEcoin 获取用户积分信息
func (r *ecoinRepoImpl) GetUserEcoin(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error) {
	var userEcoin ecoin_model.UserEcoin
	err := database.FromContext(ctx).Where("user_id = ?", userId).First(&userEcoin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userEcoin, nil
}

// CreateUserEcoin 创建用户积分记录
func (r *ecoinRepoImpl) CreateUserEcoin(ctx context.Context, userEcoin *ecoin_model.UserEcoin) error {
	return database.FromContext(ctx).Create(userEcoin).Error
}

// UpdateUserEcoinStock 更新用户积分余额
func (r *ecoinRepoImpl) UpdateUserEcoinStock(ctx context.Context, userId uint64, stock float64) error {
	return database.FromContext(ctx).Model(&ecoin_model.UserEcoin{}).
		Where("user_id = ?", userId).
		Update("available_stock", stock).Error
}

// AddEcoinTransaction 新增积分流水记录
func (r *ecoinRepoImpl) AddEcoinTransaction(ctx context.Context, transaction *ecoin_model.EcoinTransaction) error {
	return database.FromContext(ctx).Create(transaction).Error
}

// GetEcoinTransactionsByUserId 根据用户ID获取积分流水记录
func (r *ecoinRepoImpl) GetEcoinTransactionsByUserId(ctx context.Context, userId uint64, offset, limit int) ([]*ecoin_model.EcoinTransaction, error) {
	var transactions []*ecoin_model.EcoinTransaction
	err := database.FromContext(ctx).
		Where("user_id = ?", userId).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}

// CountEcoinTransactionsByUserId 统计用户积分流水数量
func (r *ecoinRepoImpl) CountEcoinTransactionsByUserId(ctx context.Context, userId uint64) (int64, error) {
	var count int64
	err := database.FromContext(ctx).
		Model(&ecoin_model.EcoinTransaction{}).
		Where("user_id = ?", userId).
		Count(&count).Error
	return count, err
}

// GetEcoinTransactionById 根据ID获取积分流水记录
func (r *ecoinRepoImpl) GetEcoinTransactionById(ctx context.Context, id uint64) (*ecoin_model.EcoinTransaction, error) {
	var transaction ecoin_model.EcoinTransaction
	err := database.FromContext(ctx).Where("id = ?", id).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

// UpdateEcoinTransactionStatus 更新积分流水状态
func (r *ecoinRepoImpl) UpdateEcoinTransactionStatus(ctx context.Context, id uint64, status int) error {
	return database.FromContext(ctx).WithContext(ctx).Model(&ecoin_model.EcoinTransaction{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetUserEcoinForUpdate 获取用户积分信息（加锁）
func (r *ecoinRepoImpl) GetUserEcoinForUpdate(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error) {
	var userEcoin ecoin_model.UserEcoin
	err := database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userId).First(&userEcoin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userEcoin, nil
}

// CreateUserEcoinStockGroup 新增积分库存分组
func (r *ecoinRepoImpl) CreateUserEcoinStockGroup(ctx context.Context, group *ecoin_model.UserEcoinStockGroup) error {
	return database.FromContext(ctx).Create(group).Error
}

// UpdateUserEcoinStockGroupRemaining 更新积分库存分组剩余数量
func (r *ecoinRepoImpl) UpdateUserEcoinStockGroupRemaining(ctx context.Context, groupId uint64, remainingStock float64) error {
	return database.FromContext(ctx).Model(&ecoin_model.UserEcoinStockGroup{}).
		Where("id = ?", groupId).
		Update("remaining_stock", remainingStock).Error
}

// DeleteUserEcoinStockGroup 删除积分库存分组
func (r *ecoinRepoImpl) DeleteUserEcoinStockGroup(ctx context.Context, groupId uint64) error {
	return database.FromContext(ctx).Where("id = ?", groupId).Delete(&ecoin_model.UserEcoinStockGroup{}).Error
}

// GetAvailableStockGroupsForUpdate 获取可用积分分组（按最早过期优先）
func (r *ecoinRepoImpl) GetAvailableStockGroupsForUpdate(ctx context.Context, userId uint64, now uint32) ([]*ecoin_model.UserEcoinStockGroup, error) {
	var groups []*ecoin_model.UserEcoinStockGroup
	err := database.FromContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		Where("user_id = ? AND remaining_stock > 0 AND (expire_time = 0 OR expire_time > ?)", userId, now).
		Order("CASE WHEN expire_time = 0 THEN 1 ELSE 0 END ASC").
		Order("expire_time ASC").
		Order("id ASC").
		Find(&groups).Error
	return groups, err
}

// GetStockGroupsByUserId 获取用户积分分组（用于展示）
func (r *ecoinRepoImpl) GetStockGroupsByUserId(ctx context.Context, userId uint64) ([]*ecoin_model.UserEcoinStockGroup, error) {
	var groups []*ecoin_model.UserEcoinStockGroup
	err := database.FromContext(ctx).
		Where("user_id = ?", userId).
		Order("CASE WHEN expire_time = 0 THEN 1 ELSE 0 END ASC").
		Order("expire_time ASC").
		Order("id DESC").
		Find(&groups).Error
	return groups, err
}

// GetStockGroupsByUserIdForUpdate 获取用户积分分组（加锁）
func (r *ecoinRepoImpl) GetStockGroupsByUserIdForUpdate(ctx context.Context, userId uint64) ([]*ecoin_model.UserEcoinStockGroup, error) {
	var groups []*ecoin_model.UserEcoinStockGroup
	err := database.FromContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		Where("user_id = ?", userId).
		Order("CASE WHEN expire_time = 0 THEN 1 ELSE 0 END ASC").
		Order("expire_time ASC").
		Order("id ASC").
		Find(&groups).Error
	return groups, err
}

// GetExpiredStockGroupsForUpdate 获取已过期且仍有余额的积分分组（加锁）
func (r *ecoinRepoImpl) GetExpiredStockGroupsForUpdate(ctx context.Context, now uint32, limit int) ([]*ecoin_model.UserEcoinStockGroup, error) {
	var groups []*ecoin_model.UserEcoinStockGroup
	query := database.FromContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		Where("expire_time > 0 AND expire_time <= ? AND remaining_stock > 0", now).
		Order("expire_time ASC").
		Order("id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&groups).Error
	return groups, err
}
