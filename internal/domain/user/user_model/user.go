package usermodel

import (
	"database/sql/driver"
	"errors"

	jsoniter "github.com/json-iterator/go"
)

const (
	UserTabName = "user_tab"
)

type User struct {
	Id       uint       `gorm:"id" json:"id"`
	Username string     `gorm:"username" json:"username"`
	Password string     `gorm:"password" json:"password"`
	Email    string     `gorm:"email" json:"email"`
	Binding  BindingMap `gorm:"binding" json:"binding"`
	Ctime    uint32     `gorm:"ctime,autoCreateTime" json:"ctime"`
	Mtime    uint32     `gorm:"mtime,autoUpdateTime" json:"mtime"`
}

func (u *User) TableName() string {
	return UserTabName
}

func (u *User) Bind(bizType string, info BindingInfo) {
	u.Binding[bizType] = info
}

func (u *User) Unbind(bizType string) {
	delete(u.Binding, bizType)
}

type BindingInfo struct {
	BizId uint64 `json:"biz_id"`
}

type BindingMap map[string]BindingInfo

func (b BindingMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type")
	}
	return jsoniter.Unmarshal(bytes, b)
}

func (b BindingMap) Value() (driver.Value, error) {
	return jsoniter.Marshal(b)
}
