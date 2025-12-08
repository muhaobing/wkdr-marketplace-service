package conf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	// AppConfig 全局配置实例
	AppConfig *Config
)

// Init 初始化配置
func Init(configPath ...string) error {
	var err error
	
	// 如果没有指定配置文件路径，尝试从几个默认位置加载
	var cfgPath string
	if len(configPath) > 0 && configPath[0] != "" {
		cfgPath = configPath[0]
	} else {
		// 默认配置文件路径
		possiblePaths := []string{
			"./conf/config.yaml",
			"./config.yaml",
			"../conf/config.yaml",
		}
		
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				cfgPath = path
				break
			}
		}
	}
	
	if cfgPath != "" && fileExists(cfgPath) {
		// 从配置文件加载
		AppConfig, err = LoadFromFile(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config from file %s: %v", cfgPath, err)
		}
		fmt.Printf("配置已从文件加载: %s\n", cfgPath)
	} else {
		// 使用默认配置
		AppConfig = GetDefaultConfig()
		fmt.Println("使用默认配置")
	}
	
	return nil
}

// LoadFromFile 从文件加载配置
func LoadFromFile(configPath string) (*Config, error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}
	
	// 读取文件
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	// 解析YAML
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}
	
	return config, nil
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	if AppConfig == nil {
		AppConfig = GetDefaultConfig()
	}
	return AppConfig
}

// GetDBDSN 获取数据库DSN连接字符串
func GetDBDSN() string {
	cfg := GetConfig()
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Database,
	)
}

// GetRedisAddr 获取Redis地址
func GetRedisAddr() string {
	cfg := GetConfig()
	return fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
