package companymodel

const CompanyTabName = "company_tab"

// Company 商城企业（按名称唯一）
type Company struct {
	Id    uint64 `gorm:"column:id" json:"id"`
	Name  string `gorm:"column:name" json:"name"`
	Ctime uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (c *Company) TableName() string {
	return CompanyTabName
}
