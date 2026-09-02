ALTER TABLE resource_assets
    ADD COLUMN media_key TEXT NOT NULL DEFAULT '';

ALTER TABLE resource_assets
    DROP COLUMN cdn_url;
