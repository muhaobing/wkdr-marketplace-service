package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type confFile struct {
	RestServer struct {
		Database struct {
			IP       string `yaml:"ip"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DB       string `yaml:"db"`
		} `yaml:"database"`
	} `yaml:"restserver"`
	Config struct {
		EcoinExpireSeconds uint32 `yaml:"ecoin_expire_seconds"`
	} `yaml:"config"`
}

func main() {
	raw, err := os.ReadFile("conf/restserver.yaml")
	if err != nil {
		log.Fatalf("read config failed: %v", err)
	}

	var conf confFile
	if err := yaml.Unmarshal(raw, &conf); err != nil {
		log.Fatalf("parse config failed: %v", err)
	}

	dbConf := conf.RestServer.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConf.User, dbConf.Password, dbConf.IP, dbConf.Port, dbConf.DB)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db failed: %v", err)
	}

	createSQL := `
CREATE TABLE IF NOT EXISTS user_ecoin_stock_group_tab (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  total_stock DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '该批次总积分',
  remaining_stock DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '该批次剩余积分',
  expire_time INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '过期时间戳(0表示不过期)',
  source_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '来源类型',
  source_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源业务ID',
  ctime INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
  mtime INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
  PRIMARY KEY (id),
  KEY idx_user_expire (user_id, expire_time),
  KEY idx_expire_time (expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户积分库存分组表';`
	if err := db.Exec(createSQL).Error; err != nil {
		log.Fatalf("create table failed: %v", err)
	}

	now := uint32(time.Now().Unix())
	expire := uint32(0)
	if conf.Config.EcoinExpireSeconds > 0 {
		expire = now + conf.Config.EcoinExpireSeconds
	}

	backfillSQL := `
INSERT INTO user_ecoin_stock_group_tab
  (user_id, total_stock, remaining_stock, expire_time, source_type, source_id, ctime, mtime)
SELECT
  ue.user_id,
  ue.available_stock,
  ue.available_stock,
  ?,
  'system',
  'legacy_migration',
  ?,
  ?
FROM user_ecoin_tab ue
WHERE ue.available_stock > 0
  AND NOT EXISTS (
    SELECT 1
    FROM user_ecoin_stock_group_tab sg
    WHERE sg.user_id = ue.user_id
  );`
	if err := db.Exec(backfillSQL, expire, now, now).Error; err != nil {
		log.Fatalf("backfill stock groups failed: %v", err)
	}

	var count int64
	if err := db.Raw("SELECT COUNT(1) FROM user_ecoin_stock_group_tab").Scan(&count).Error; err != nil {
		log.Fatalf("query count failed: %v", err)
	}

	fmt.Printf("migration success, stock group rows=%d, default_expire_time=%d\n", count, expire)
}
