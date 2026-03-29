// 执行 script/sql/migrate_enterprise_user_company.sql（仓库根目录运行）
// go run ./script/golang/migrate_enterprise -conf conf/restserver.yaml
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
}

func main() {
	confPath := flag.String("conf", "conf/restserver.yaml", "path to restserver.yaml")
	sqlPath := flag.String("sql", "script/sql/migrate_enterprise_user_company.sql", "migration sql file")
	flag.Parse()

	raw, err := os.ReadFile(*confPath)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	var c confFile
	if err := yaml.Unmarshal(raw, &c); err != nil {
		log.Fatalf("parse yaml: %v", err)
	}
	d := c.RestServer.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		d.User, d.Password, d.IP, d.Port, d.DB)

	sqlBytes, err := os.ReadFile(*sqlPath)
	if err != nil {
		log.Fatalf("read sql: %v", err)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	sqldb, err := db.DB()
	if err != nil {
		log.Fatalf("sql db: %v", err)
	}
	defer sqldb.Close()

	if _, err := sqldb.Exec(string(sqlBytes)); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	fmt.Println("migrate_enterprise_user_company: ok")
}
