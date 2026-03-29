// 将 user_ecoin_stock_group_tab 中所有批次的 expire_time 统一为「运行时刻所在自然月」最后一天 23:59:59（Asia/Shanghai）。
//
// 在 marketplace 模块根目录执行：
//
//	go run ./script/golang/backfill_ecoin_expire_end_of_month
//
// 仅打印将要写入的值而不更新数据库：
//
//	$env:DRY_RUN="1"; go run ./script/golang/backfill_ecoin_expire_end_of_month
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
}

func endOfCurrentMonthUnixShanghai(now time.Time) uint32 {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.UTC
	}
	t := now.In(loc)
	end := time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 0, loc)
	return uint32(end.Unix())
}

func main() {
	confPath := "conf/restserver.yaml"
	if len(os.Args) > 1 {
		confPath = os.Args[1]
	}
	raw, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatalf("read config %q: %v", confPath, err)
	}

	var conf confFile
	if err := yaml.Unmarshal(raw, &conf); err != nil {
		log.Fatalf("parse config: %v", err)
	}

	dbConf := conf.RestServer.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConf.User, dbConf.Password, dbConf.IP, dbConf.Port, dbConf.DB)

	expire := endOfCurrentMonthUnixShanghai(time.Now())
	nowUnix := uint32(time.Now().Unix())

	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.UTC
	}
	fmt.Printf("expire_time target: %d (%s Asia/Shanghai)\n", expire,
		time.Unix(int64(expire), 0).In(loc).Format("2006-01-02 15:04:05"))

	if os.Getenv("DRY_RUN") == "1" {
		fmt.Println("DRY_RUN=1, skip UPDATE")
		return
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	res := db.Exec(`UPDATE user_ecoin_stock_group_tab SET expire_time = ?, mtime = ?`, expire, nowUnix)
	if res.Error != nil {
		log.Fatalf("update: %v", res.Error)
	}
	fmt.Printf("rows affected: %d\n", res.RowsAffected)
}
