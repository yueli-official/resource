CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE resources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        TEXT NOT NULL,
    title           TEXT NOT NULL,
    slug            TEXT NOT NULL,
    summary         TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    type            TEXT NOT NULL,
    cover_asset_id  TEXT NOT NULL DEFAULT '',
    access          TEXT NOT NULL DEFAULT 'public',
    price_cents     INT NOT NULL DEFAULT 0,
    points_cost     INT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'draft',
    download_count  BIGINT NOT NULL DEFAULT 0,
    tags            JSONB NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_resources_slug              ON resources (slug);
CREATE INDEX        ix_resources_status_type_time  ON resources (status, type, created_at);
CREATE INDEX        ix_resources_owner             ON resources (owner_id);

-- One asset belongs to exactly one resource (spec §3); CASCADE keeps rows tidy.
CREATE TABLE resource_assets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID NOT NULL REFERENCES resources (id) ON DELETE CASCADE,
    asset_id    TEXT NOT NULL,
    label       TEXT NOT NULL DEFAULT '',
    cdn_url     TEXT NOT NULL DEFAULT '',   -- captured at finalize for public delivery
    size        BIGINT NOT NULL DEFAULT 0,
    mime        TEXT NOT NULL DEFAULT '',
    filename    TEXT NOT NULL DEFAULT '',
    sort        INT NOT NULL DEFAULT 0
);

CREATE INDEX ix_resource_assets_resource ON resource_assets (resource_id, sort);
