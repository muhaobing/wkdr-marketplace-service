-- 企业用户与 company_id：在已有库上执行（请先备份）
-- 1) 企业表
CREATE TABLE IF NOT EXISTS `company_tab` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` VARCHAR(256) NOT NULL COMMENT '企业名称（唯一）',
  `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
  `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商城企业';

-- 2) 用户表：company_id + 唯一键调整
ALTER TABLE `user_tab` ADD COLUMN `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=个人用户，>0=企业用户' AFTER `email`;
ALTER TABLE `user_tab` DROP INDEX `uk_tel_no`;
ALTER TABLE `user_tab` DROP INDEX `uk_email`;
ALTER TABLE `user_tab` ADD UNIQUE KEY `uk_company_tel` (`company_id`, `tel_no`);
ALTER TABLE `user_tab` ADD UNIQUE KEY `uk_company_email` (`company_id`, `email`);

-- 3) biz_code 范围
ALTER TABLE `biz_code_enum_tab` ADD COLUMN `scope` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=个人 1=企业' AFTER `name`;
UPDATE `biz_code_enum_tab` SET `scope` = 1 WHERE `code` = 'LawMind_Enterprise';

-- 4) 企业积分（独立表）
CREATE TABLE IF NOT EXISTS `company_ecoin_tab` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `company_id` BIGINT UNSIGNED NOT NULL COMMENT '企业ID',
  `available_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '可用积分余额',
  `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
  `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业积分账户';

CREATE TABLE IF NOT EXISTS `company_ecoin_stock_group_tab` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `company_id` BIGINT UNSIGNED NOT NULL COMMENT '企业ID',
  `total_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '该批次总积分',
  `remaining_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '该批次剩余积分',
  `expire_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '过期时间戳(0表示不过期)',
  `source_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '来源类型',
  `source_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源业务ID',
  `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
  `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_company_expire` (`company_id`, `expire_time`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业积分库存分组';

CREATE TABLE IF NOT EXISTS `company_ecoin_transaction_tab` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
  `company_id` BIGINT UNSIGNED NOT NULL COMMENT '企业ID',
  `operator_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作人商城用户ID',
  `amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '积分变动',
  `before_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '操作前企业积分余额',
  `after_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '操作后企业积分余额',
  `tx_type` TINYINT NOT NULL DEFAULT 0 COMMENT '1增加 2扣除',
  `source_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '来源类型',
  `source_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源业务ID',
  `description` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '交易状态',
  `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
  `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_company_id` (`company_id`),
  KEY `idx_operator` (`operator_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业积分流水';
