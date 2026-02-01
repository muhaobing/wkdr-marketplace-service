package ecoin_model

const (
	UserEcoinTabName = "user_ecoin_tab"
)

type UserEcoin struct {
	Id             uint64  `gorm:"column:id" json:"id"`
	UserId         uint64  `gorm:"column:user_id" json:"user_id"`
	AvailableStock float64 `gorm:"column:available_stock" json:"available_stock"`
	Ctime          uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (u *UserEcoin) TableName() string {
	return UserEcoinTabName
}
