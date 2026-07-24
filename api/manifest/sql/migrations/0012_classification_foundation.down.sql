DROP TABLE IF EXISTS resource_tag_lookup_entries;

DROP INDEX IF EXISTS resource_taxonomies_replacement_idx;
DROP INDEX IF EXISTS resource_taxonomies_parent_idx;
DROP INDEX IF EXISTS resource_taxonomies_active_kind_idx;

ALTER TABLE object_taxonomies
    DROP CONSTRAINT IF EXISTS resource_object_taxonomy_identity_fk,
    ADD CONSTRAINT object_taxonomies_taxonomy_id_fkey
        FOREIGN KEY (taxonomy_id) REFERENCES taxonomies(id) ON DELETE CASCADE;

ALTER TABLE taxonomies
    DROP CONSTRAINT IF EXISTS resource_taxonomy_replacement_fk,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_parent_fk,
    ADD CONSTRAINT taxonomies_parent_id_fkey
        FOREIGN KEY (parent_id) REFERENCES taxonomies(id) ON DELETE SET NULL;

ALTER TABLE taxonomies
    DROP CONSTRAINT IF EXISTS resource_taxonomy_self_replacement_check,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_replacement_state_check,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_tag_flat_check,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_editorial_position_check,
    DROP CONSTRAINT IF EXISTS resource_tag_status_check,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_status_check,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_kind_check,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_catalog_kind_identity_key,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_catalog_identity_key,
    DROP CONSTRAINT IF EXISTS resource_taxonomy_catalog_fk,
    DROP COLUMN IF EXISTS first_activated_at,
    DROP COLUMN IF EXISTS replacement_id,
    DROP COLUMN IF EXISTS editorial_position,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS catalog_id;

DROP TABLE IF EXISTS resource_classification_policy_profiles;
DROP TABLE IF EXISTS resource_classification_catalogs;
