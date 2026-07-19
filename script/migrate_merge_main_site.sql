-- 商城合并主站：company_tab 主站企业 ID 映射
ALTER TABLE company_tab
  ADD COLUMN biz_company_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '主站 companyId' AFTER name;

CREATE INDEX idx_company_biz_company_id ON company_tab (biz_company_id);

-- 确保 biz_code 枚举
INSERT IGNORE INTO biz_code_enum_tab (code, name, scope, ctime, mtime) VALUES
  ('LawMind_ToC', 'LawMind 个人用户', 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  ('LawMind_Enterprise', 'LawMind 企业用户', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
