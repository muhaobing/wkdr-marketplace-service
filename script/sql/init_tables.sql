-- ============================================================
-- wkdr-marketplace-service 数据库表结构
-- 生成时间: 2026-02-01
-- 说明: 重新创建所有表（会删除已有表）
-- ============================================================

-- -----------------------------------------------------------
-- 1. 用户表 (user_tab)
-- 商城中心用户表，支持多业务平台绑定
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `user_tab`;
CREATE TABLE `user_tab` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `tel_no` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
    `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `secret_key` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户密钥(SHA256加密)',
    `binding` JSON COMMENT '业务平台绑定信息，格式: {"biz_code": {"biz_id": 123}}',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tel_no` (`tel_no`),
    UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- -----------------------------------------------------------
-- 2. 用户积分表 (user_ecoin_tab)
-- 记录用户积分余额
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `user_ecoin_tab`;
CREATE TABLE `user_ecoin_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `available_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '可用积分余额',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户积分表';

-- -----------------------------------------------------------
-- 3. 积分流水表 (ecoin_transaction_tab)
-- 记录积分变动明细
-- tx_type: 1-增加, 2-扣除
-- source_type: system-系统赠送, order-订单奖励, consume-积分消费, refund-退款返还, transfer-转账
-- status: 0-处理中, 1-已完成, 2-已失败
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `ecoin_transaction_tab`;
CREATE TABLE `ecoin_transaction_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '积分变动数量(正数增加,负数扣除)',
    `before_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '操作前积分余额',
    `after_stock` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '操作后积分余额',
    `tx_type` TINYINT NOT NULL DEFAULT 0 COMMENT '交易类型: 1-增加, 2-扣除',
    `source_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '来源类型',
    `source_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源业务ID',
    `description` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '描述信息',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '交易状态: 0-处理中, 1-已完成, 2-已失败',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_source` (`source_type`, `source_id`),
    KEY `idx_ctime` (`ctime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分流水表';
