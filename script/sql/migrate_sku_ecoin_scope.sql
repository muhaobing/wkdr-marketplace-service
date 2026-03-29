-- 积分商品归属：0=个人积分包 1=企业积分包；已有数据默认 0（个人）
ALTER TABLE `sku_tab`
  ADD COLUMN `ecoin_scope` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '积分包：0=个人 1=企业' AFTER `multi_select`;

UPDATE `sku_tab` SET `ecoin_scope` = 0 WHERE `ecoin_scope` > 1;
