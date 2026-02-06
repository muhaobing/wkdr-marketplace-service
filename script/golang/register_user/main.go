package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ==================== 在这里修改注册信息 ====================

const (
	Email    = "1642990705@qq.com" // 用户邮箱（可为空）
	Phone    = "18281695061"       // 用户手机号（可为空，但至少填一个）
	Password = "ayaka"             // 用户密码（必填，用于生成密钥）
	Role     = 1                   // 用户角色：0-普通用户，1-管理员
)

// ============================================================

// Config 配置结构
type Config struct {
	RestServer struct {
		Database struct {
			IP       string `yaml:"ip"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DB       string `yaml:"db"`
		} `yaml:"database"`
	} `yaml:"restserver"`
}

// User 用户模型
type User struct {
	Id        uint   `gorm:"column:id;primaryKey;autoIncrement"`
	TelNo     string `gorm:"column:tel_no"`
	Email     string `gorm:"column:email"`
	SecretKey string `gorm:"column:secret_key"`
	Role      uint8  `gorm:"column:role"`
	Ctime     uint32 `gorm:"column:ctime"`
	Mtime     uint32 `gorm:"column:mtime"`
}

func (User) TableName() string {
	return "user_tab"
}

// GenerateSecretKey 生成密钥
func GenerateSecretKey(secret string, userId uint) string {
	combined := fmt.Sprintf("%s:%d", secret, userId)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

func main() {
	// 参数校验
	if Email == "" && Phone == "" {
		log.Fatal("错误: Email 和 Phone 至少需要填写一个")
	}
	if Password == "" {
		log.Fatal("错误: Password 不能为空")
	}

	// 读取配置文件
	configData, err := os.ReadFile("conf/restserver.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	// 构建 DSN
	dbConfig := config.RestServer.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.IP,
		dbConfig.Port,
		dbConfig.DB,
	)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 检查邮箱是否已存在
	if Email != "" {
		var existingUser User
		if err := db.Where("email = ?", Email).First(&existingUser).Error; err == nil {
			log.Fatalf("错误: 邮箱 %s 已被注册", Email)
		}
	}

	// 检查手机号是否已存在
	if Phone != "" {
		var existingUser User
		if err := db.Where("tel_no = ?", Phone).First(&existingUser).Error; err == nil {
			log.Fatalf("错误: 手机号 %s 已被注册", Phone)
		}
	}

	// 创建用户
	now := uint32(time.Now().Unix())
	user := &User{
		TelNo: Phone,
		Email: Email,
		Role:  Role,
		Ctime: now,
		Mtime: now,
	}

	// 先创建用户获取 ID
	if err := db.Create(user).Error; err != nil {
		log.Fatalf("创建用户失败: %v", err)
	}

	// 生成并更新密钥
	secretKey := GenerateSecretKey(Password, user.Id)
	if err := db.Model(user).Update("secret_key", secretKey).Error; err != nil {
		log.Fatalf("更新密钥失败: %v", err)
	}

	// 打印结果
	fmt.Println("========================================")
	fmt.Println("用户注册成功!")
	fmt.Println("========================================")
	fmt.Printf("用户ID:   %d\n", user.Id)
	fmt.Printf("邮箱:     %s\n", user.Email)
	fmt.Printf("手机号:   %s\n", user.TelNo)
	fmt.Printf("角色:     %d (%s)\n", user.Role, getRoleName(user.Role))
	fmt.Printf("密钥:     %s\n", secretKey)
	fmt.Println("========================================")
	fmt.Println("请妥善保管您的密码，登录时需要使用!")
}

func getRoleName(role uint8) string {
	switch role {
	case 0:
		return "普通用户"
	case 1:
		return "管理员"
	default:
		return "未知"
	}
}
