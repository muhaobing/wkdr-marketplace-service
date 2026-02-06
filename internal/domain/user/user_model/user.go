package usermodel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

const (
	UserTabName = "user_tab"
)

// 用户角色
const (
	RoleUser  uint8 = 0 // 普通用户
	RoleAdmin uint8 = 1 // 管理员
)

type User struct {
	Id        uint   `gorm:"column:id" json:"id"`
	TelNo     string `gorm:"column:tel_no" json:"tel_no"`
	Email     string `gorm:"column:email" json:"email"`
	SecretKey string `gorm:"column:secret_key" json:"-"` // 密钥不对外暴露
	Role      uint8  `gorm:"column:role" json:"role"`
	Ctime     uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime     uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
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

// IsAdmin 判断是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) TableName() string {
	return UserTabName
}

// ToJSON 序列化为 JSON 字符串（用于 session）
func (u *User) ToJSON() (string, error) {
	bytes, err := jsoniter.Marshal(u)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// UserFromJSON 从 JSON 字符串反序列化用户
func UserFromJSON(jsonStr string) (*User, error) {
	var user User
	if err := jsoniter.Unmarshal([]byte(jsonStr), &user); err != nil {
		return nil, err
	}
	return &user, nil
}
