-- 将 user_tab.tel_no / email 改为可 NULL，空串改为 NULL 写入，避免 uk_company_tel / uk_company_email 在空串上冲突
-- 执行前建议备份；已有全空串行可先改为 NULL

ALTER TABLE `user_tab`
  MODIFY COLUMN `tel_no` VARCHAR(20) NULL DEFAULT NULL COMMENT '手机号',
  MODIFY COLUMN `email` VARCHAR(128) NULL DEFAULT NULL COMMENT '邮箱';

UPDATE `user_tab` SET `tel_no` = NULL WHERE `tel_no` = '';
UPDATE `user_tab` SET `email` = NULL WHERE `email` = '';
