package ecoin_model

const (
	UserEcoinStockGroupTabName = "user_ecoin_stock_group_tab"
)

// UserEcoinStockGroup 用户积分库存分组（不同批次可有不同过期时间）
type UserEcoinStockGroup struct {
	Id             uint64  `gorm:"column:id" json:"id"`
	UserId         uint64  `gorm:"column:user_id" json:"user_id"`
	TotalStock     float64 `gorm:"column:total_stock" json:"total_stock"`
	RemainingStock float64 `gorm:"column:remaining_stock" json:"remaining_stock"`
	ExpireTime     uint32  `gorm:"column:expire_time" json:"expire_time"` // 0 表示不过期
	SourceType     string  `gorm:"column:source_type" json:"source_type"`
	SourceId       string  `gorm:"column:source_id" json:"source_id"`
	Ctime          uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (g *UserEcoinStockGroup) TableName() string {
	return UserEcoinStockGroupTabName
}
