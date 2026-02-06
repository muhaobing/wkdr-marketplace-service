package usermodel

const (
	UserBindingTabName = "user_binding_tab"
)

// UserBinding 用户绑定信息
type UserBinding struct {
	Id        uint64 `gorm:"column:id" json:"id"`
	UserId    uint   `gorm:"column:user_id" json:"user_id"`
	BizCode   string `gorm:"column:biz_code" json:"biz_code"`
	BizUserId uint64 `gorm:"column:biz_user_id" json:"biz_user_id"`
	Ctime     uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime     uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (b *UserBinding) TableName() string {
	return UserBindingTabName
}
