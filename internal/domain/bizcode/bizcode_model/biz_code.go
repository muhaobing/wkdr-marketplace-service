package bizcodemodel

const BizCodeEnumTabName = "biz_code_enum_tab"

// BizCode 业务平台编码枚举（LawMind 多租户类型等）
type BizCode struct {
	Id    uint   `gorm:"column:id" json:"id"`
	Code  string `gorm:"column:code" json:"code"`
	Name  string `gorm:"column:name" json:"name"`
	Ctime uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (b *BizCode) TableName() string {
	return BizCodeEnumTabName
}
