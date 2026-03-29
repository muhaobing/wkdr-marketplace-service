-- 新库：直接增加 sku_scope（若曾用旧脚本建过 ecoin_scope，请改跑 migrate_rename_ecoin_scope_to_sku_scope.sql）
ALTER TABLE `sku_tab`
  ADD COLUMN `sku_scope` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'SKU可见范围：0=通用 1=仅个人 2=仅企业' AFTER `multi_select`;

UPDATE `sku_tab` SET `sku_scope` = 0 WHERE `sku_scope` > 2;
