package companyecoin_model

const CompanyEcoinTabName = "company_ecoin_tab"

// CompanyEcoin 企业积分账户
type CompanyEcoin struct {
	Id             uint64                    `gorm:"column:id" json:"id"`
	CompanyId      uint64                    `gorm:"column:company_id" json:"company_id"`
	AvailableStock float64                   `gorm:"column:available_stock" json:"available_stock"`
	StockGroups    []*CompanyEcoinStockGroup `gorm:"-" json:"stock_groups,omitempty"`
	Ctime          uint32                    `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          uint32                    `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (u *CompanyEcoin) TableName() string {
	return CompanyEcoinTabName
}
