-- Rollback for 000002_extension_delivery
DROP TABLE IF EXISTS extension_tokens;
ALTER TABLE wechat_accounts DROP COLUMN IF EXISTS delivery_mode;
