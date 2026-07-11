-- migrate:irreversible
-- Cover image public URL, snapshotted at cover finalize (spec §6 ④). The resource
-- service has no "resolve public URL by asset id" path, so we store the asset
-- service's public CDN URL redundantly (same pattern as resource_assets.cdn_url)
-- to render covers on browse/detail without a per-row asset-service round trip.
ALTER TABLE resources ADD COLUMN cover_url TEXT NOT NULL DEFAULT '';
