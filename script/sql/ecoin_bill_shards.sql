-- 分表 ecoin_bill_tab_00000000 .. ecoin_bill_tab_00000009（与 migrate_ecoin_bill_shards/main.go 中 DDL 一致）
-- bill_id varchar(32) UNIQUE；主键 bill_id

CREATE TABLE IF NOT EXISTS `ecoin_bill_tab_00000000` (
  `bill_id` varchar(32) NOT NULL COMMENT '业务账单ID（EB+时间+哈希+分片）',
  `user_id` bigint unsigned NOT NULL COMMENT '商城用户ID',
  `company_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '0=个人积分；>0=企业积分账户',
  `amount` decimal(20,4) NOT NULL COMMENT '预扣积分数量',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '1=incomplete 2=completed 3=cancelled',
  `deduct_user_tx_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '个人积分扣款流水ID',
  `deduct_company_tx_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '企业积分扣款流水ID',
  `ctime` int unsigned NOT NULL,
  `mtime` int unsigned NOT NULL,
  PRIMARY KEY (`bill_id`),
  UNIQUE KEY `uk_bill_id` (`bill_id`),
  KEY `idx_status_ctime` (`status`,`ctime`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分账单分表00000000';
