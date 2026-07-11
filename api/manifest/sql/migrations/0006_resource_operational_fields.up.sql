-- migrate:irreversible
ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS view_count BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS ix_resources_status_published_at ON resources (status, published_at DESC);
