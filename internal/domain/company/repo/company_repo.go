package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	companymodel "wdkr-marketplace-service/internal/domain/company/company_model"
)

type companyRepoImpl struct{}

func NewCompanyRepo() CompanyRepo {
	return &companyRepoImpl{}
}

func (r *companyRepoImpl) GetById(ctx context.Context, id uint64) (*companymodel.Company, error) {
	if id == 0 {
		return nil, nil
	}
	var c companymodel.Company
	err := database.FromContext(ctx).Where("id = ?", id).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *companyRepoImpl) GetByName(ctx context.Context, name string) (*companymodel.Company, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	var c companymodel.Company
	err := database.FromContext(ctx).Where("name = ?", name).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *companyRepoImpl) Create(ctx context.Context, c *companymodel.Company) error {
	return database.FromContext(ctx).Create(c).Error
}

func (r *companyRepoImpl) SearchByName(ctx context.Context, q string, limit int) ([]*companymodel.Company, error) {
	q = strings.TrimSpace(q)
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	var list []*companymodel.Company
	db := database.FromContext(ctx).Model(&companymodel.Company{})
	if q != "" {
		db = db.Where("name LIKE ?", "%"+q+"%")
	}
	err := db.Order("id DESC").Limit(limit).Find(&list).Error
	return list, err
}
