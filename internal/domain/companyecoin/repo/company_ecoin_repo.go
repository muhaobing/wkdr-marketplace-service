package repo

import (
	"context"
	"errors"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	"wdkr-marketplace-service/internal/domain/companyecoin/companyecoin_model"
)

type companyEcoinRepoImpl struct{}

func NewCompanyEcoinRepo() CompanyEcoinRepo {
	return &companyEcoinRepoImpl{}
}

func (r *companyEcoinRepoImpl) GetCompanyEcoin(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error) {
	var row companyecoin_model.CompanyEcoin
	err := database.FromContext(ctx).Where("company_id = ?", companyId).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *companyEcoinRepoImpl) CreateCompanyEcoin(ctx context.Context, row *companyecoin_model.CompanyEcoin) error {
	return database.FromContext(ctx).Create(row).Error
}

func (r *companyEcoinRepoImpl) UpdateCompanyEcoinStock(ctx context.Context, companyId uint64, stock float64) error {
	return database.FromContext(ctx).Model(&companyecoin_model.CompanyEcoin{}).
		Where("company_id = ?", companyId).
		Update("available_stock", stock).Error
}

func (r *companyEcoinRepoImpl) GetCompanyEcoinForUpdate(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error) {
	var row companyecoin_model.CompanyEcoin
	err := database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("company_id = ?", companyId).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *companyEcoinRepoImpl) AddTransaction(ctx context.Context, tx *companyecoin_model.CompanyEcoinTransaction) error {
	return database.FromContext(ctx).Create(tx).Error
}

func (r *companyEcoinRepoImpl) GetTransactionsByCompanyId(ctx context.Context, companyId uint64, offset, limit int) ([]*companyecoin_model.CompanyEcoinTransaction, error) {
	var list []*companyecoin_model.CompanyEcoinTransaction
	err := database.FromContext(ctx).
		Where("company_id = ?", companyId).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *companyEcoinRepoImpl) CountTransactionsByCompanyId(ctx context.Context, companyId uint64) (int64, error) {
	var n int64
	err := database.FromContext(ctx).
		Model(&companyecoin_model.CompanyEcoinTransaction{}).
		Where("company_id = ?", companyId).
		Count(&n).Error
	return n, err
}

func (r *companyEcoinRepoImpl) CreateStockGroup(ctx context.Context, g *companyecoin_model.CompanyEcoinStockGroup) error {
	return database.FromContext(ctx).Create(g).Error
}

func (r *companyEcoinRepoImpl) UpdateStockGroupRemaining(ctx context.Context, groupId uint64, remaining float64) error {
	return database.FromContext(ctx).Model(&companyecoin_model.CompanyEcoinStockGroup{}).
		Where("id = ?", groupId).
		Update("remaining_stock", remaining).Error
}

func (r *companyEcoinRepoImpl) DeleteStockGroup(ctx context.Context, groupId uint64) error {
	return database.FromContext(ctx).Where("id = ?", groupId).Delete(&companyecoin_model.CompanyEcoinStockGroup{}).Error
}

func (r *companyEcoinRepoImpl) GetAvailableStockGroupsForUpdate(ctx context.Context, companyId uint64, now uint32) ([]*companyecoin_model.CompanyEcoinStockGroup, error) {
	var groups []*companyecoin_model.CompanyEcoinStockGroup
	err := database.FromContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		Where("company_id = ? AND remaining_stock > 0 AND (expire_time = 0 OR expire_time > ?)", companyId, now).
		Order("CASE WHEN expire_time = 0 THEN 1 ELSE 0 END ASC").
		Order("expire_time ASC").
		Order("id ASC").
		Find(&groups).Error
	return groups, err
}

func (r *companyEcoinRepoImpl) GetStockGroupsByCompanyId(ctx context.Context, companyId uint64) ([]*companyecoin_model.CompanyEcoinStockGroup, error) {
	var groups []*companyecoin_model.CompanyEcoinStockGroup
	err := database.FromContext(ctx).
		Where("company_id = ?", companyId).
		Order("CASE WHEN expire_time = 0 THEN 1 ELSE 0 END ASC").
		Order("expire_time ASC").
		Order("id DESC").
		Find(&groups).Error
	return groups, err
}

func (r *companyEcoinRepoImpl) GetStockGroupsByCompanyIdForUpdate(ctx context.Context, companyId uint64) ([]*companyecoin_model.CompanyEcoinStockGroup, error) {
	var groups []*companyecoin_model.CompanyEcoinStockGroup
	err := database.FromContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		Where("company_id = ?", companyId).
		Order("CASE WHEN expire_time = 0 THEN 1 ELSE 0 END ASC").
		Order("expire_time ASC").
		Order("id ASC").
		Find(&groups).Error
	return groups, err
}

func (r *companyEcoinRepoImpl) GetExpiredStockGroupsForUpdate(ctx context.Context, now uint32, limit int) ([]*companyecoin_model.CompanyEcoinStockGroup, error) {
	var groups []*companyecoin_model.CompanyEcoinStockGroup
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
