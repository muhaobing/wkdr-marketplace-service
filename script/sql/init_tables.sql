-- ============================================================
-- wkdr-marketplace-service 数据库表结构
-- 生成时间: 2026-02-01
-- 说明: 重新创建所有表（会删除已有表）
-- ============================================================

-- -----------------------------------------------------------
-- 1. 用户表 (user_tab)
-- 商城中心用户表
-- role: 0-普通用户, 1-管理员
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `user_tab`;
CREATE TABLE `user_tab` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `tel_no` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
    `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `secret_key` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户密钥(SHA256加密)',
    `role` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户角色: 0-User, 1-Admin',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tel_no` (`tel_no`),
    UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- -----------------------------------------------------------
-- 1.1 用户绑定表 (user_binding_tab)
-- 记录用户与业务平台的绑定关系
-- 唯一键: user_id + biz_code + biz_user_id（同业务域同账号只允许绑定一个商城账号）
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `user_binding_tab`;
CREATE TABLE `user_binding_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` INT UNSIGNED NOT NULL COMMENT '商城用户ID',
    `biz_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '业务平台代码',
    `biz_user_id` BIGINT UNSIGNED NOT NULL COMMENT '业务平台用户ID',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_biz_binding` (`biz_code`, `biz_user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户绑定表';

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

-- -----------------------------------------------------------
-- 4. 商品表 (sku_tab)
-- 记录商城商品信息
-- sku_status: 0-未上架, 1-已上架
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `sku_tab`;
CREATE TABLE `sku_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '商品ID',
    `biz_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '商品所属的业务编码',
    `sku_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商品代码，在biz_code下唯一',
    `sku_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '商品名称',
    `sku_avatar` MEDIUMTEXT COMMENT '商品图片(Base64)',
    `sku_desc` TEXT COMMENT '商品描述',
    `sku_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上架状态: 0-未上架, 1-已上架',
    `cost` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '商品售价(积分)',
    `delivery_method` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '商品履约回调接口URL',
    `multi_select` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否支持多选下单: 0-不支持, 1-支持',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_biz_sku` (`biz_code`, `sku_code`),
    KEY `idx_biz_code` (`biz_code`),
    KEY `idx_sku_status` (`sku_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';

-- -----------------------------------------------------------
-- 5. 订单表 (order_tab)
-- 记录商城订单信息，支持同时下单多个SKU
-- status: 0-待支付, 1-已支付, 2-已履约, 3-已取消, 4-已退款
-- pay_type: ecoin-积分支付, money-货币支付
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `order_tab`;
CREATE TABLE `order_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `item_count` INT NOT NULL DEFAULT 0 COMMENT '商品种类数量',
    `total_quantity` INT NOT NULL DEFAULT 0 COMMENT '商品总数量',
    `original_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '原价',
    `pay_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '实付金额',
    `pay_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '支付类型: ecoin/money',
    `payment_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '支付订单号(货币支付)',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0-待支付, 1-已支付, 2-已履约, 3-已取消, 4-已退款',
    `pay_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付时间戳',
    `fulfill_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '履约时间戳',
    `cancel_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '取消时间戳',
    `cancel_reason` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '取消原因',
    `remark` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_ctime` (`ctime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- -----------------------------------------------------------
-- 6. 订单明细表 (order_item_tab)
-- 记录订单包含的SKU明细
-- fulfill_status: 0-待履约, 1-履约成功, 2-履约失败
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `order_item_tab`;
CREATE TABLE `order_item_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
    `sku_id` BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    `sku_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'SKU编码',
    `sku_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'SKU名称',
    `sku_avatar` MEDIUMTEXT COMMENT 'SKU图片(Base64)',
    `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
    `unit_price` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '单价',
    `total_price` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '小计',
    `fulfill_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '履约状态: 0-待履约, 1-成功, 2-失败',
    `fulfill_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '履约时间戳',
    `fulfill_msg` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '履约消息',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_order_no` (`order_no`),
    KEY `idx_sku_id` (`sku_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单明细表';

-- -----------------------------------------------------------
-- 7. 支付订单表 (payment_order_tab)
-- 记录支付订单信息
-- status: 0-待支付, 1-已支付, 2-已关闭, 3-已退款
-- biz_type: recharge-充值积分, purchase-购买商品
-- channel: wechat-微信支付, alipay-支付宝
-- pay_method: native-扫码支付, jsapi-JSAPI支付, h5-H5支付
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `payment_order_tab`;
CREATE TABLE `payment_order_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '支付订单号',
    `biz_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '业务订单号',
    `biz_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '业务类型: recharge/purchase',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `channel` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '支付渠道: wechat/alipay',
    `pay_method` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '支付方式: native/jsapi/h5',
    `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '支付金额(分)',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0-待支付, 1-已支付, 2-已关闭, 3-已退款',
    `channel_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道订单号(如微信交易号)',
    `pay_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付时间戳',
    `expire_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '过期时间戳',
    `notify_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '回调通知URL',
    `extra` JSON COMMENT '扩展信息',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_biz_order_no` (`biz_order_no`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_ctime` (`ctime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单表';

-- -----------------------------------------------------------
-- 8. 退款记录表 (payment_refund_tab)
-- 记录退款信息
-- status: 0-处理中, 1-退款成功, 2-退款失败
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `payment_refund_tab`;
CREATE TABLE `payment_refund_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `refund_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '退款单号',
    `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '关联支付订单号',
    `channel` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '支付渠道',
    `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '退款金额(分)',
    `reason` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '退款原因',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0-处理中, 1-成功, 2-失败',
    `channel_refund_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道退款号',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_refund_no` (`refund_no`),
    KEY `idx_order_no` (`order_no`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款记录表';

-- -----------------------------------------------------------
-- 9. 购物车表 (cart_item_tab)
-- 记录用户购物车商品
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `cart_item_tab`;
CREATE TABLE `cart_item_tab` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `sku_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_sku` (`user_id`, `sku_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车表';
