-- migrate:irreversible
ALTER TABLE resources
  ADD COLUMN IF NOT EXISTS delivery_kind TEXT NOT NULL DEFAULT 'asset_file',
  ADD COLUMN IF NOT EXISTS delivery_payload JSONB NOT NULL DEFAULT '{}'::jsonb;
