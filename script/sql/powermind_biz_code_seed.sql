-- PowerMind 商城库专用 biz_code（在 power_marketplace_db 执行，勿在 marketplace 库执行）
DELETE FROM `biz_code_enum_tab` WHERE `code` IN ('LawMind_ToC', 'LawMind_Enterprise');

INSERT IGNORE INTO `biz_code_enum_tab` (`code`, `name`, `scope`, `ctime`, `mtime`) VALUES
('PowerMind_ToC', 'PowerMind 个人用户', 0, 0, 0),
('PowerMind_Enterprise', 'PowerMind 企业用户', 1, 0, 0),
('PowerMind_Ops', 'PowerMind 运营/超管', 0, 0, 0);
