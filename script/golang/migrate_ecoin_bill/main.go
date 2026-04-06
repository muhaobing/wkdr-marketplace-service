// 一次性迁移：创建 ecoin_bill_tab
// 用法（仓库根目录）：go run ./script/golang/migrate_ecoin_bill -conf conf/restserver.yaml
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

type yamlRoot struct {
	Restserver struct {
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
	confPath := flag.String("conf", "conf/restserver.yaml", "restserver.yaml path")
	sqlPath := flag.String("sql", "script/sql/ecoin_bill_tab.sql", "DDL file path")
	flag.Parse()

	absConf, err := filepath.Abs(*confPath)
	if err != nil {
		log.Fatalf("conf path: %v", err)
	}
	raw, err := os.ReadFile(absConf)
	if err != nil {
		log.Fatalf("read %s: %v", absConf, err)
	}
	var root yamlRoot
	if err := yaml.Unmarshal(raw, &root); err != nil {
		log.Fatalf("yaml: %v", err)
	}
	d := root.Restserver.Database
	port := d.Port
	if port == 0 {
		port = 3306
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		d.User, d.Password, d.IP, port, d.DB)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("sql open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	absSQL, err := filepath.Abs(*sqlPath)
	if err != nil {
		log.Fatalf("sql path: %v", err)
	}
	ddl, err := os.ReadFile(absSQL)
	if err != nil {
		log.Fatalf("read ddl: %v", err)
	}
	if _, err := db.Exec(string(ddl)); err != nil {
		log.Fatalf("exec ddl: %v", err)
	}
	log.Printf("ok: applied %s", absSQL)
}
