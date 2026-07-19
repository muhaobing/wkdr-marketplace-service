package repo

import (
	"context"

	companymodel "wdkr-marketplace-service/internal/domain/company/company_model"
)

// CompanyRepo 企业表
type CompanyRepo interface {
	GetById(ctx context.Context, id uint64) (*companymodel.Company, error)
	GetByName(ctx context.Context, name string) (*companymodel.Company, error)
	GetByBizCompanyId(ctx context.Context, bizCompanyId uint64) (*companymodel.Company, error)
	Create(ctx context.Context, c *companymodel.Company) error
	SearchByName(ctx context.Context, q string, limit int) ([]*companymodel.Company, error)
}
