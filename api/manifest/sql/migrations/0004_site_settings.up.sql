-- migrate:irreversible
CREATE TABLE IF NOT EXISTS resource_home_settings (
    id      INT PRIMARY KEY DEFAULT 1,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT ck_resource_home_settings_singleton CHECK (id = 1)
);

CREATE TABLE IF NOT EXISTS resource_site_settings (
    key     TEXT PRIMARY KEY,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT ck_resource_site_settings_key CHECK (key <> '')
);
