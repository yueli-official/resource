ALTER TABLE resource_assets
    ADD COLUMN cdn_url TEXT NOT NULL DEFAULT '';

ALTER TABLE resource_assets
    DROP COLUMN media_key;
