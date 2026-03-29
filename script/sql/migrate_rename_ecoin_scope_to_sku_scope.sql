-- 已有 ecoin_scope 列的库：重命名为 sku_scope（MySQL 8.0+）
ALTER TABLE `sku_tab`
  CHANGE COLUMN `ecoin_scope` `sku_scope` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'SKU可见范围：0=通用 1=仅个人 2=仅企业';

UPDATE `sku_tab` SET `sku_scope` = 0 WHERE `sku_scope` > 2;
