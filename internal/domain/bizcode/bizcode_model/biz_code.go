package bizcodemodel

const BizCodeEnumTabName = "biz_code_enum_tab"

// 业务线范围：与商城用户类型一致方可绑定
const (
	ScopePersonal   uint8 = 0
	ScopeEnterprise uint8 = 1
)

// BizCode 业务平台编码枚举（LawMind 多租户类型等）
type BizCode struct {
	Id    uint   `gorm:"column:id" json:"id"`
	Code  string `gorm:"column:code" json:"code"`
	Name  string `gorm:"column:name" json:"name"`
	Scope uint8  `gorm:"column:scope" json:"scope"` // 0 个人 1 企业
	Ctime uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (b *BizCode) TableName() string {
	return BizCodeEnumTabName
}
