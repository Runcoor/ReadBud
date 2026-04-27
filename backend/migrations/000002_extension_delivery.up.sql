-- ReadBud Migration 000002: Browser-extension delivery mode + extension tokens
--
-- Adds:
--   1. wechat_accounts.delivery_mode — how articles reach WeChat:
--        api       = direct draft/add API (requires verified service account)
--        extension = browser plugin auto-fills the editor (default for personal accounts)
--        manual    = copy/paste by hand
--   2. extension_tokens — long-lived bearer tokens for the browser extension to
--      authenticate against the ReadBud API. One user can have multiple tokens
--      (e.g., one per browser/device). Tokens are stored hashed (sha256).

-- ============================================================
-- 1. wechat_accounts.delivery_mode
-- ============================================================
ALTER TABLE wechat_accounts
    ADD COLUMN IF NOT EXISTS delivery_mode VARCHAR(32) NOT NULL DEFAULT 'extension';

-- ============================================================
-- 2. extension_tokens
-- ============================================================
CREATE TABLE IF NOT EXISTS extension_tokens (
    id           BIGSERIAL PRIMARY KEY,
    public_id    VARCHAR(26) NOT NULL UNIQUE,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    name         VARCHAR(64) NOT NULL,
    token_hash   VARCHAR(128) NOT NULL UNIQUE,
    token_prefix VARCHAR(16) NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    created_by   BIGINT,
    updated_by   BIGINT
);

CREATE INDEX IF NOT EXISTS idx_extension_tokens_user ON extension_tokens(user_id) WHERE revoked_at IS NULL;
