// 一次性迁移：创建 ecoin_bill_tab_0..9，删除旧表 ecoin_bill_tab、points_bill_tab
// 用法（仓库根目录）：go run ./script/golang/migrate_ecoin_bill_shards -conf conf/restserver.yaml
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

func ddlForShard(shard int) string {
	return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  bill_id varchar(32) NOT NULL COMMENT '业务账单ID',
  user_id bigint unsigned NOT NULL COMMENT '商城用户ID',
  company_id bigint unsigned NOT NULL DEFAULT '0' COMMENT '0=个人积分；>0=企业积分账户',
  amount decimal(20,4) NOT NULL COMMENT '预扣积分数量',
  status tinyint unsigned NOT NULL DEFAULT '1' COMMENT '1=incomplete 2=completed 3=cancelled',
  deduct_user_tx_id bigint unsigned NOT NULL DEFAULT '0' COMMENT '个人积分扣款流水ID',
  deduct_company_tx_id bigint unsigned NOT NULL DEFAULT '0' COMMENT '企业积分扣款流水ID',
  ctime int unsigned NOT NULL,
  mtime int unsigned NOT NULL,
  PRIMARY KEY (bill_id),
  UNIQUE KEY uk_bill_id (bill_id),
  KEY idx_status_ctime (status,ctime),
  KEY idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分账单分表%d';
`, tableName(shard), shard)
}

func tableName(shard int) string {
	return fmt.Sprintf("`ecoin_bill_tab_%08d`", shard)
}

func main() {
	confPath := flag.String("conf", "conf/restserver.yaml", "restserver.yaml path")
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

	for _, drop := range []string{
		"DROP TABLE IF EXISTS `ecoin_bill_tab`",
		"DROP TABLE IF EXISTS `points_bill_tab`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_0`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_1`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_2`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_3`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_4`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_5`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_6`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_7`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_8`",
		"DROP TABLE IF EXISTS `ecoin_bill_tab_9`",
	} {
		if _, err := db.Exec(drop); err != nil {
			log.Fatalf("exec %s: %v", drop, err)
		}
		log.Printf("ok: %s", drop)
	}

	for i := 0; i < 10; i++ {
		if _, err := db.Exec(ddlForShard(i)); err != nil {
			log.Fatalf("create shard %d: %v", i, err)
		}
		log.Printf("ok: created %s", tableName(i))
	}
	log.Print("migrate_ecoin_bill_shards: done")
}
