package repo

import (
	"context"

	"wdkr-marketplace-service/internal/domain/companyecoin/companyecoin_model"
)

// CompanyEcoinRepo 企业积分仓储
type CompanyEcoinRepo interface {
	GetCompanyEcoin(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error)
	CreateCompanyEcoin(ctx context.Context, row *companyecoin_model.CompanyEcoin) error
	UpdateCompanyEcoinStock(ctx context.Context, companyId uint64, stock float64) error
	GetCompanyEcoinForUpdate(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error)

	AddTransaction(ctx context.Context, tx *companyecoin_model.CompanyEcoinTransaction) error
	GetTransactionsByCompanyId(ctx context.Context, companyId uint64, offset, limit int) ([]*companyecoin_model.CompanyEcoinTransaction, error)
	CountTransactionsByCompanyId(ctx context.Context, companyId uint64) (int64, error)

	CreateStockGroup(ctx context.Context, g *companyecoin_model.CompanyEcoinStockGroup) error
	UpdateStockGroupRemaining(ctx context.Context, groupId uint64, remaining float64) error
	DeleteStockGroup(ctx context.Context, groupId uint64) error
	GetAvailableStockGroupsForUpdate(ctx context.Context, companyId uint64, now uint32) ([]*companyecoin_model.CompanyEcoinStockGroup, error)
	GetStockGroupsByCompanyId(ctx context.Context, companyId uint64) ([]*companyecoin_model.CompanyEcoinStockGroup, error)
	GetStockGroupsByCompanyIdForUpdate(ctx context.Context, companyId uint64) ([]*companyecoin_model.CompanyEcoinStockGroup, error)
	GetExpiredStockGroupsForUpdate(ctx context.Context, now uint32, limit int) ([]*companyecoin_model.CompanyEcoinStockGroup, error)
}
