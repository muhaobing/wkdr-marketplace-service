-- 已有库升级：为 sku_tab 增加履约模式字段（若已存在可跳过对应语句）
ALTER TABLE `sku_tab`
  ADD COLUMN `fulfill_mode` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '履约模式: 0-接口回调, 1-积分发放' AFTER `delivery_method`,
  ADD COLUMN `fulfill_ecoin_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '积分发放模式下每件发放的积分数' AFTER `fulfill_mode`;
