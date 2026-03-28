// 一次性迁移：创建 biz_code_enum_tab、写入三条枚举、扩展 user_binding_tab.biz_code
// 用法（仓库根目录）：go run ./script/golang/migrate_biz_code -conf conf/restserver.yaml
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
	confPath := flag.String("conf", "conf/restserver.yaml", "restserver.yaml 路径（相对当前工作目录）")
	flag.Parse()

	abs, err := filepath.Abs(*confPath)
	if err != nil {
		log.Fatalf("conf path: %v", err)
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		log.Fatalf("read %s: %v", abs, err)
	}

	var root yamlRoot
	if err := yaml.Unmarshal(raw, &root); err != nil {
		log.Fatalf("yaml: %v", err)
	}
	d := root.Restserver.Database
	if d.IP == "" || d.User == "" || d.DB == "" {
		log.Fatal("database.ip / user / db 不能为空")
	}
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
		log.Fatalf("ping database: %v", err)
	}

	steps := []string{
		`CREATE TABLE IF NOT EXISTS biz_code_enum_tab (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    code VARCHAR(64) NOT NULL COMMENT '编码，如 LawMind_ToC',
    name VARCHAR(128) NOT NULL COMMENT '展示名称',
    ctime INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    mtime INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (id),
    UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='业务平台编码枚举'`,

		`INSERT IGNORE INTO biz_code_enum_tab (code, name, ctime, mtime) VALUES
('LawMind_ToC', 'LawMind C端用户', 0, 0),
('LawMind_Enterprise', 'LawMind 企业用户', 0, 0),
('LawMind_Admin', 'LawMind 运营人员', 0, 0)`,

		`ALTER TABLE user_binding_tab MODIFY COLUMN biz_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '业务平台代码（见 biz_code_enum_tab）'`,
	}

	for i, q := range steps {
		if _, err := db.Exec(q); err != nil {
			log.Fatalf("step %d: %v", i+1, err)
		}
	}

	dropSortOrderIfExists(db)

	log.Println("migrate_biz_code: ok (table + 3 rows + alter user_binding_tab.biz_code; list order by id)")
}

// 旧版表曾含 sort_order，与按 id 排序重复，存在则删除
func dropSortOrderIfExists(db *sql.DB) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'biz_code_enum_tab'
		  AND COLUMN_NAME = 'sort_order'
	`).Scan(&n)
	if err != nil || n == 0 {
		return
	}
	if _, err := db.Exec(`ALTER TABLE biz_code_enum_tab DROP COLUMN sort_order`); err != nil {
		log.Printf("migrate_biz_code: warn drop sort_order: %v", err)
		return
	}
	log.Println("migrate_biz_code: dropped legacy column sort_order")
}
