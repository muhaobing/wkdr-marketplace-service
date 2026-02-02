package usermodel

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

const (
	UserTabName = "user_tab"
)

type User struct {
	Id        uint       `gorm:"column:id" json:"id"`
	TelNo     string     `gorm:"column:tel_no" json:"tel_no"`
	Email     string     `gorm:"column:email" json:"email"`
	SecretKey string     `gorm:"column:secret_key" json:"secret_key"`
	Binding   BindingMap `gorm:"column:binding" json:"binding"`
	Ctime     uint32     `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime     uint32     `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

// GenerateSecretKey 根据secret和用户ID生成加密后的密钥
func GenerateSecretKey(secret string, userId uint) string {
	combined := fmt.Sprintf("%s:%d", secret, userId)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// VerifySecretKey 验证密钥是否正确
func (u *User) VerifySecretKey(secret string) bool {
	expectedKey := GenerateSecretKey(secret, u.Id)
	return u.SecretKey == expectedKey
}

// HasBinding 检查用户是否已绑定指定业务平台
func (u *User) HasBinding(bizCode string) bool {
	if u.Binding == nil {
		return false
	}
	_, exists := u.Binding[bizCode]
	return exists
}

// GetBinding 获取指定业务平台的绑定信息
func (u *User) GetBinding(bizCode string) (BindingInfo, bool) {
	if u.Binding == nil {
		return BindingInfo{}, false
	}
	info, exists := u.Binding[bizCode]
	return info, exists
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
	BizUserId uint64 `json:"biz_user_id"`
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
