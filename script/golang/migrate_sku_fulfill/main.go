// 读取 conf/restserver.yaml，为 sku_tab 按需添加 fulfill_mode / fulfill_ecoin_amount（列已存在则跳过）
// 用法：在项目根目录执行 go run ./script/golang/migrate_sku_fulfill
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
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
}

func main() {
	cfgPath := "conf/restserver.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Fatalf("read config %s: %v", cfgPath, err)
	}

	var conf confFile
	if err := yaml.Unmarshal(raw, &conf); err != nil {
		log.Fatalf("parse config: %v", err)
	}

	dc := conf.RestServer.Database
	if dc.IP == "" || dc.DB == "" || dc.User == "" {
		log.Fatal("invalid database config in yaml")
	}
	port := dc.Port
	if port == 0 {
		port = 3306
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		dc.User, dc.Password, dc.IP, port, dc.DB)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("sql open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	schema := dc.DB

	fulfillModeExists, err := columnExists(db, schema, "sku_tab", "fulfill_mode")
	if err != nil {
		log.Fatalf("check fulfill_mode: %v", err)
	}
	if !fulfillModeExists {
		_, err = db.Exec(`
ALTER TABLE sku_tab
  ADD COLUMN fulfill_mode TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '履约模式: 0-接口回调, 1-积分发放'
  AFTER delivery_method`)
		if err != nil {
			log.Fatalf("add fulfill_mode: %v", err)
		}
		fmt.Println("added column: sku_tab.fulfill_mode")
	} else {
		fmt.Println("skip: sku_tab.fulfill_mode already exists")
	}

	ecoinAmtExists, err := columnExists(db, schema, "sku_tab", "fulfill_ecoin_amount")
	if err != nil {
		log.Fatalf("check fulfill_ecoin_amount: %v", err)
	}
	if !ecoinAmtExists {
		_, err = db.Exec(`
ALTER TABLE sku_tab
  ADD COLUMN fulfill_ecoin_amount DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '积分发放模式下每件发放的积分数'
  AFTER fulfill_mode`)
		if err != nil {
			log.Fatalf("add fulfill_ecoin_amount: %v", err)
		}
		fmt.Println("added column: sku_tab.fulfill_ecoin_amount")
	} else {
		fmt.Println("skip: sku_tab.fulfill_ecoin_amount already exists")
	}

	fmt.Println("migrate_sku_fulfill: ok")
}

func columnExists(db *sql.DB, schema, table, column string) (bool, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, schema, table, column).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
