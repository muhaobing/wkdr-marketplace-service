#!/usr/bin/env bash
# 商城合并：company_tab 增加主站 companyId 映射列；确保 biz_code 枚举存在
set -euo pipefail

MYSQL_HOST="${MYSQL_HOST:-rm-2ze7ce8lnejdjnt70xo.mysql.rds.aliyuncs.com}"
MYSQL_USER="${MYSQL_USER:-DBManager}"
MYSQL_PASS="${MYSQL_PASS:-DBA_powermind}"
MYSQL_DB="${MYSQL_DB:-marketplace}"

mysql_exec() {
  mysql -h "$MYSQL_HOST" -u "$MYSQL_USER" -p"$MYSQL_PASS" "$MYSQL_DB" -e "$1"
}

echo "[migrate] adding biz_company_id to company_tab if missing..."
mysql_exec "ALTER TABLE company_tab ADD COLUMN IF NOT EXISTS biz_company_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '主站 companyId';" || \
  mysql_exec "ALTER TABLE company_tab ADD COLUMN biz_company_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '主站 companyId';" 2>/dev/null || true

mysql_exec "CREATE INDEX idx_company_biz_company_id ON company_tab (biz_company_id);" 2>/dev/null || true

echo "[migrate] ensuring biz_code_enum rows..."
mysql_exec "INSERT IGNORE INTO biz_code_enum_tab (code, name, scope, ctime, mtime) VALUES
  ('LawMind_ToC', 'LawMind 个人用户', 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  ('LawMind_Enterprise', 'LawMind 企业用户', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());"

echo "[migrate] done."
