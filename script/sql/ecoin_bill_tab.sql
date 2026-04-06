-- 积分账单表（OpenAPI 预扣 / 确认 / 取消）
CREATE TABLE IF NOT EXISTS `ecoin_bill_tab` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '商城用户ID',
  `company_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '0=个人积分；>0=企业积分账户',
  `amount` decimal(20,4) NOT NULL COMMENT '预扣积分数量',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '1=incomplete 2=completed 3=cancelled',
  `deduct_user_tx_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '个人积分扣款流水ID',
  `deduct_company_tx_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '企业积分扣款流水ID',
  `ctime` int unsigned NOT NULL,
  `mtime` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status_ctime` (`status`,`ctime`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分账单';
