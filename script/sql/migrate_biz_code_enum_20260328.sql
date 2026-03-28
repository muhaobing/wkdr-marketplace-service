-- 已有库增量：业务平台编码枚举表 + user_binding.biz_code 扩长
-- 执行前请备份。旧数据若使用 lawmind，请按需 UPDATE user_binding_tab SET biz_code='LawMind_ToC' WHERE biz_code='lawmind';

CREATE TABLE IF NOT EXISTS `biz_code_enum_tab` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `code` VARCHAR(64) NOT NULL COMMENT '编码，如 LawMind_ToC',
    `name` VARCHAR(128) NOT NULL COMMENT '展示名称',
    `ctime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `mtime` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='业务平台编码枚举';

INSERT IGNORE INTO `biz_code_enum_tab` (`code`, `name`, `ctime`, `mtime`) VALUES
('LawMind_ToC', 'LawMind C端用户', 0, 0),
('LawMind_Enterprise', 'LawMind 企业用户', 0, 0),
('LawMind_Admin', 'LawMind 运营人员', 0, 0);

ALTER TABLE `user_binding_tab` MODIFY COLUMN `biz_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '业务平台代码（见 biz_code_enum_tab）';
