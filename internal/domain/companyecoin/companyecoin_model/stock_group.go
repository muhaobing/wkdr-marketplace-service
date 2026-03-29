package companyecoin_model

const CompanyEcoinStockGroupTabName = "company_ecoin_stock_group_tab"

// CompanyEcoinStockGroup 企业积分批次
type CompanyEcoinStockGroup struct {
	Id             uint64  `gorm:"column:id" json:"id"`
	CompanyId      uint64  `gorm:"column:company_id" json:"company_id"`
	TotalStock     float64 `gorm:"column:total_stock" json:"total_stock"`
	RemainingStock float64 `gorm:"column:remaining_stock" json:"remaining_stock"`
	ExpireTime     uint32  `gorm:"column:expire_time" json:"expire_time"`
	SourceType     string  `gorm:"column:source_type" json:"source_type"`
	SourceId       string  `gorm:"column:source_id" json:"source_id"`
	Ctime          uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (g *CompanyEcoinStockGroup) TableName() string {
	return CompanyEcoinStockGroupTabName
}
