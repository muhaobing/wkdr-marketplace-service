package companymodel

const CompanyTabName = "company_tab"

// Company 商城企业（按名称唯一；biz_company_id 映射主站 companyId）
type Company struct {
	Id            uint64 `gorm:"column:id" json:"id"`
	Name          string `gorm:"column:name" json:"name"`
	BizCompanyId  uint64 `gorm:"column:biz_company_id" json:"biz_company_id"` // 主站 companyId，0=未绑定
	Ctime         uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime         uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (c *Company) TableName() string {
	return CompanyTabName
}
