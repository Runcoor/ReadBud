-- ReadBud Initial Schema Migration
-- Covers all 18 tables from spec Sections 11.1-11.18
-- All time fields: timestamptz
-- Primary key: bigserial, external ID: ULID varchar(26)
-- No database enum, use text + check
-- Migrations are additive only

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================================
-- 1. users (Section 11.1)
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    public_id       VARCHAR(26) NOT NULL UNIQUE,
    username        VARCHAR(64) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    nickname        VARCHAR(64) NOT NULL,
    role            VARCHAR(32) NOT NULL DEFAULT 'editor',
    status          SMALLINT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      BIGINT,
    updated_by      BIGINT
);

-- ============================================================
-- 2. provider_configs (Section 11.2)
-- ============================================================
CREATE TABLE IF NOT EXISTS provider_configs (
    id              BIGSERIAL PRIMARY KEY,
    public_id       VARCHAR(26) NOT NULL UNIQUE,
    provider_type   VARCHAR(32) NOT NULL,
    provider_name   VARCHAR(64) NOT NULL,
    config_json     JSONB,
    secret_json_enc TEXT,
    status          SMALLINT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      BIGINT,
    updated_by      BIGINT
);

CREATE INDEX IF NOT EXISTS idx_provider_configs_type ON provider_configs(provider_type);

-- ============================================================
-- 3. wechat_accounts (Section 11.3)
-- ============================================================
CREATE TABLE IF NOT EXISTS wechat_accounts (
    id                  BIGSERIAL PRIMARY KEY,
    public_id           VARCHAR(26) NOT NULL UNIQUE,
    name                VARCHAR(64) NOT NULL,
    app_id              VARCHAR(64) NOT NULL,
    app_secret_enc      TEXT NOT NULL,
    token_mode          VARCHAR(32) NOT NULL DEFAULT 'direct',
    stable_access_token TEXT,
    token_expire_at     TIMESTAMPTZ,
    gateway_user_api_id VARCHAR(128),
    is_default          SMALLINT NOT NULL DEFAULT 0,
    status              SMALLINT NOT NULL DEFAULT 1,
    remark              VARCHAR(255),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    created_by          BIGINT,
    updated_by          BIGINT
);

-- ============================================================
-- 4. content_tasks (Section 11.4)
-- ============================================================
CREATE TABLE IF NOT EXISTS content_tasks (
    id                BIGSERIAL PRIMARY KEY,
    public_id         VARCHAR(26) NOT NULL UNIQUE,
    task_no           VARCHAR(64) NOT NULL UNIQUE,
    keyword           VARCHAR(255) NOT NULL,
    audience          VARCHAR(255),
    tone              VARCHAR(64),
    target_words      INT NOT NULL DEFAULT 2000,
    image_mode        VARCHAR(32) NOT NULL DEFAULT 'auto',
    chart_mode        SMALLINT NOT NULL DEFAULT 1,
    publish_mode      VARCHAR(32) NOT NULL DEFAULT 'manual',
    publish_at        TIMESTAMPTZ,
    wechat_account_id BIGINT REFERENCES wechat_accounts(id),
    status            VARCHAR(32) NOT NULL DEFAULT 'pending',
    progress          INT NOT NULL DEFAULT 0,
    current_stage     VARCHAR(64),
    error_message     TEXT,
    result_draft_id   BIGINT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        BIGINT,
    updated_by        BIGINT
);

CREATE INDEX IF NOT EXISTS idx_task_status_created ON content_tasks(status, created_at);

-- ============================================================
-- 5. source_documents (Section 11.5)
-- ============================================================
CREATE TABLE IF NOT EXISTS source_documents (
    id                BIGSERIAL PRIMARY KEY,
    public_id         VARCHAR(26) NOT NULL UNIQUE,
    task_id           BIGINT NOT NULL REFERENCES content_tasks(id),
    source_type       VARCHAR(32) NOT NULL,
    site_name         VARCHAR(128),
    source_url        TEXT NOT NULL,
    title             VARCHAR(512) NOT NULL,
    author            VARCHAR(128),
    published_at      TIMESTAMPTZ,
    crawled_at        TIMESTAMPTZ NOT NULL,
    raw_html          TEXT,
    plain_text        TEXT,
    summary_json      JSONB,
    hot_score         DECIMAL(10,2) NOT NULL DEFAULT 0,
    relevance_score   DECIMAL(10,2) NOT NULL DEFAULT 0,
    data_points_json  JSONB,
    image_clues_json  JSONB,
    license_note      VARCHAR(255),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        BIGINT,
    updated_by        BIGINT
);

CREATE INDEX IF NOT EXISTS idx_source_task_score ON source_documents(task_id, hot_score DESC);

-- ============================================================
-- 6. article_drafts (Section 11.6)
-- ============================================================
CREATE TABLE IF NOT EXISTS article_drafts (
    id                  BIGSERIAL PRIMARY KEY,
    public_id           VARCHAR(26) NOT NULL UNIQUE,
    task_id             BIGINT NOT NULL REFERENCES content_tasks(id),
    wechat_account_id   BIGINT REFERENCES wechat_accounts(id),
    title               VARCHAR(255) NOT NULL,
    subtitle            VARCHAR(255),
    digest              VARCHAR(512),
    author_name         VARCHAR(64),
    content_source_url  TEXT,
    cover_asset_id      BIGINT,
    compiled_html       TEXT,
    outline_json        JSONB,
    quality_score       DECIMAL(10,2) NOT NULL DEFAULT 0,
    similarity_score    DECIMAL(10,2) NOT NULL DEFAULT 0,
    risk_level          VARCHAR(32) NOT NULL DEFAULT 'low',
    review_status       VARCHAR(32) NOT NULL DEFAULT 'pending',
    version             INT NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    created_by          BIGINT,
    updated_by          BIGINT
);

CREATE INDEX IF NOT EXISTS idx_draft_task_version ON article_drafts(task_id, version DESC);

-- ============================================================
-- 7. article_blocks (Section 11.7)
-- ============================================================
CREATE TABLE IF NOT EXISTS article_blocks (
    id                BIGSERIAL PRIMARY KEY,
    public_id         VARCHAR(26) NOT NULL UNIQUE,
    draft_id          BIGINT NOT NULL REFERENCES article_drafts(id),
    sort_no           INT NOT NULL,
    block_type        VARCHAR(32) NOT NULL,
    heading           VARCHAR(255),
    text_md           TEXT,
    html_fragment     TEXT,
    asset_id          BIGINT,
    source_refs_json  JSONB,
    prompt_text       TEXT,
    status            VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        BIGINT,
    updated_by        BIGINT
);

CREATE INDEX IF NOT EXISTS idx_block_draft_sort ON article_blocks(draft_id, sort_no);

-- ============================================================
-- 8. assets (Section 11.8)
-- ============================================================
CREATE TABLE IF NOT EXISTS assets (
    id                    BIGSERIAL PRIMARY KEY,
    public_id             VARCHAR(26) NOT NULL UNIQUE,
    asset_type            VARCHAR(32) NOT NULL,
    source_kind           VARCHAR(32) NOT NULL,
    mime_type             VARCHAR(64) NOT NULL,
    storage_provider      VARCHAR(32) NOT NULL,
    bucket                VARCHAR(128) NOT NULL,
    object_key            VARCHAR(512) NOT NULL,
    local_path            VARCHAR(512),
    width                 INT,
    height                INT,
    size_bytes            BIGINT,
    sha256                VARCHAR(128) NOT NULL,
    source_url            TEXT,
    source_page_url       TEXT,
    source_site           VARCHAR(128),
    source_author         VARCHAR(128),
    license_type          VARCHAR(64),
    attribution_text      VARCHAR(255),
    prompt_text           TEXT,
    is_ai_generated       SMALLINT NOT NULL DEFAULT 0,
    wechat_url            TEXT,
    wechat_media_id       VARCHAR(128),
    wechat_upload_status  VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,
    created_by            BIGINT,
    updated_by            BIGINT
);

CREATE INDEX IF NOT EXISTS idx_assets_sha256 ON assets(sha256);

-- Add FK for article_blocks.asset_id and article_drafts.cover_asset_id
ALTER TABLE article_blocks
    ADD CONSTRAINT fk_blocks_asset FOREIGN KEY (asset_id) REFERENCES assets(id);
ALTER TABLE article_drafts
    ADD CONSTRAINT fk_drafts_cover_asset FOREIGN KEY (cover_asset_id) REFERENCES assets(id);
-- Add FK for content_tasks.result_draft_id
ALTER TABLE content_tasks
    ADD CONSTRAINT fk_tasks_result_draft FOREIGN KEY (result_draft_id) REFERENCES article_drafts(id);

-- ============================================================
-- 9. publish_jobs (Section 11.9)
-- ============================================================
CREATE TABLE IF NOT EXISTS publish_jobs (
    id                      BIGSERIAL PRIMARY KEY,
    public_id               VARCHAR(26) NOT NULL UNIQUE,
    draft_id                BIGINT NOT NULL REFERENCES article_drafts(id),
    wechat_account_id       BIGINT NOT NULL REFERENCES wechat_accounts(id),
    publish_mode            VARCHAR(32) NOT NULL,
    schedule_at             TIMESTAMPTZ,
    status                  VARCHAR(32) NOT NULL DEFAULT 'queued',
    retry_count             INT NOT NULL DEFAULT 0,
    last_error              TEXT,
    provider_request_json   JSONB,
    provider_response_json  JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    created_by              BIGINT,
    updated_by              BIGINT
);

CREATE INDEX IF NOT EXISTS idx_job_status_schedule ON publish_jobs(status, schedule_at);

-- ============================================================
-- 10. publish_records (Section 11.10)
-- ============================================================
CREATE TABLE IF NOT EXISTS publish_records (
    id                BIGSERIAL PRIMARY KEY,
    public_id         VARCHAR(26) NOT NULL UNIQUE,
    publish_job_id    BIGINT NOT NULL REFERENCES publish_jobs(id),
    draft_id          BIGINT NOT NULL REFERENCES article_drafts(id),
    wechat_account_id BIGINT NOT NULL REFERENCES wechat_accounts(id),
    draft_media_id    VARCHAR(128),
    publish_id        VARCHAR(128),
    article_id        VARCHAR(128),
    article_url       TEXT,
    published_at      TIMESTAMPTZ,
    publish_status    VARCHAR(32) NOT NULL,
    extra_json        JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        BIGINT,
    updated_by        BIGINT
);

-- ============================================================
-- 11. metrics_snapshots (Section 11.11) — partitioned by month
-- ============================================================
CREATE TABLE IF NOT EXISTS metrics_snapshots (
    id                BIGSERIAL,
    public_id         VARCHAR(26) NOT NULL,
    wechat_account_id BIGINT NOT NULL,
    article_id        VARCHAR(128) NOT NULL,
    metric_date       DATE NOT NULL,
    read_count        INT,
    read_user_count   INT,
    share_count       INT,
    share_user_count  INT,
    add_fans_count    INT,
    cancel_fans_count INT,
    net_fans_count    INT,
    raw_json          JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        BIGINT,
    updated_by        BIGINT,
    PRIMARY KEY (id, metric_date)
) PARTITION BY RANGE (metric_date);

-- Create partitions for 2026
CREATE TABLE IF NOT EXISTS metrics_snapshots_2026_01 PARTITION OF metrics_snapshots
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE IF NOT EXISTS metrics_snapshots_2026_02 PARTITION OF metrics_snapshots
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE IF NOT EXISTS metrics_snapshots_2026_03 PARTITION OF metrics_snapshots
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE IF NOT EXISTS metrics_snapshots_2026_04 PARTITION OF metrics_snapshots
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE IF NOT EXISTS metrics_snapshots_2026_05 PARTITION OF metrics_snapshots
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE IF NOT EXISTS metrics_snapshots_2026_06 PARTITION OF metrics_snapshots
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE INDEX IF NOT EXISTS idx_metrics_article_date ON metrics_snapshots(article_id, metric_date);

-- ============================================================
-- 12. brand_profiles (Section 11.12)
-- ============================================================
CREATE TABLE IF NOT EXISTS brand_profiles (
    id                BIGSERIAL PRIMARY KEY,
    public_id         VARCHAR(26) NOT NULL UNIQUE,
    name              VARCHAR(64) NOT NULL,
    brand_tone        TEXT,
    forbidden_words   JSONB,
    preferred_words   JSONB,
    cta_rules         JSONB,
    cover_style_rules JSONB,
    image_style_rules JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        BIGINT,
    updated_by        BIGINT
);

-- ============================================================
-- 13. style_profiles (Section 11.13)
-- ============================================================
CREATE TABLE IF NOT EXISTS style_profiles (
    id                  BIGSERIAL PRIMARY KEY,
    public_id           VARCHAR(26) NOT NULL UNIQUE,
    name                VARCHAR(64) NOT NULL,
    applicable_scene    VARCHAR(255),
    opening_template    TEXT,
    structure_template  JSONB,
    closing_template    TEXT,
    sentence_preference JSONB,
    title_preference    JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    created_by          BIGINT,
    updated_by          BIGINT
);

-- ============================================================
-- 14. draft_versions (Section 11.14)
-- ============================================================
CREATE TABLE IF NOT EXISTS draft_versions (
    id              BIGSERIAL PRIMARY KEY,
    public_id       VARCHAR(26) NOT NULL UNIQUE,
    draft_id        BIGINT NOT NULL REFERENCES article_drafts(id),
    version_no      INT NOT NULL,
    title           VARCHAR(255),
    digest          VARCHAR(512),
    blocks_json     JSONB,
    html_snapshot   TEXT,
    operator_id     BIGINT,
    change_reason   VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      BIGINT,
    updated_by      BIGINT
);

CREATE INDEX IF NOT EXISTS idx_draft_versions_draft ON draft_versions(draft_id);

-- ============================================================
-- 15. content_citations (Section 11.15)
-- ============================================================
CREATE TABLE IF NOT EXISTS content_citations (
    id                  BIGSERIAL PRIMARY KEY,
    public_id           VARCHAR(26) NOT NULL UNIQUE,
    draft_id            BIGINT NOT NULL REFERENCES article_drafts(id),
    block_id            BIGINT NOT NULL REFERENCES article_blocks(id),
    source_document_id  BIGINT NOT NULL REFERENCES source_documents(id),
    citation_type       VARCHAR(32) NOT NULL,
    citation_text       TEXT,
    source_link         TEXT,
    source_note         VARCHAR(255),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    created_by          BIGINT,
    updated_by          BIGINT
);

CREATE INDEX IF NOT EXISTS idx_citations_draft ON content_citations(draft_id);

-- ============================================================
-- 16. review_rules (Section 11.16)
-- ============================================================
CREATE TABLE IF NOT EXISTS review_rules (
    id            BIGSERIAL PRIMARY KEY,
    public_id     VARCHAR(26) NOT NULL UNIQUE,
    rule_type     VARCHAR(32) NOT NULL,
    rule_content  TEXT NOT NULL,
    risk_level    VARCHAR(32) NOT NULL,
    is_enabled    SMALLINT NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    created_by    BIGINT,
    updated_by    BIGINT
);

-- ============================================================
-- 17. distribution_packages (Section 11.17)
-- ============================================================
CREATE TABLE IF NOT EXISTS distribution_packages (
    id                      BIGSERIAL PRIMARY KEY,
    public_id               VARCHAR(26) NOT NULL UNIQUE,
    draft_id                BIGINT NOT NULL UNIQUE REFERENCES article_drafts(id),
    community_copy          TEXT,
    moments_copy            TEXT,
    summary_card            TEXT,
    comment_guide           TEXT,
    next_topic_suggestion   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    created_by              BIGINT,
    updated_by              BIGINT
);

-- ============================================================
-- 18. topic_library (Section 11.18)
-- ============================================================
CREATE TABLE IF NOT EXISTS topic_library (
    id                BIGSERIAL PRIMARY KEY,
    public_id         VARCHAR(26) NOT NULL UNIQUE,
    keyword           VARCHAR(255) NOT NULL,
    audience          VARCHAR(255),
    article_goal      VARCHAR(64),
    historical_score  DECIMAL(10,2) NOT NULL DEFAULT 0,
    last_used_at      TIMESTAMPTZ,
    recommend_weight  DECIMAL(10,2) NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        BIGINT,
    updated_by        BIGINT
);

CREATE INDEX IF NOT EXISTS idx_topic_keyword_trgm ON topic_library USING gin (keyword gin_trgm_ops);

-- ============================================================
-- Seed: default admin user (password: admin123, bcrypt hash)
-- ============================================================
INSERT INTO users (public_id, username, password_hash, nickname, role, status)
VALUES ('01JAAAAAAAAAAAAAAAAAAAAAAA', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '管理员', 'admin', 1)
ON CONFLICT (username) DO NOTHING;
