-- 已有库升级：为 sku_tab 增加履约模式字段（可重复执行；列已存在则跳过）
-- 用法：mysql -h <host> -u <user> -p <database> < alter_sku_tab_fulfill.sql

DELIMITER $$

DROP PROCEDURE IF EXISTS sp_add_sku_fulfill_columns$$
CREATE PROCEDURE sp_add_sku_fulfill_columns()
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'sku_tab'
      AND COLUMN_NAME = 'fulfill_mode'
  ) THEN
    ALTER TABLE `sku_tab`
      ADD COLUMN `fulfill_mode` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '履约模式: 0-接口回调, 1-积分发放'
      AFTER `delivery_method`;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'sku_tab'
      AND COLUMN_NAME = 'fulfill_ecoin_amount'
  ) THEN
    ALTER TABLE `sku_tab`
      ADD COLUMN `fulfill_ecoin_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '积分发放模式下每件发放的积分数'
      AFTER `fulfill_mode`;
  END IF;
END$$

DELIMITER ;

CALL sp_add_sku_fulfill_columns();
DROP PROCEDURE IF EXISTS sp_add_sku_fulfill_columns;
