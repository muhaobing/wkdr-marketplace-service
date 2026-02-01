package ecoin_model

type UserEcoin struct {
	Id             uint64  `gorm:"id" json:"id"`
	UserId         uint64  `gorm:"user_id" json:"user_id"`
	AvailableStock float64 `gorm:"available_stock" json:"available_stock"`
	Ctime          uint32  `gorm:"ctime,autoCreateTime" json:"ctime"`
	Mtime          uint32  `gorm:"mtime,autoUpdateTime" json:"mtime"`
}
