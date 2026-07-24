-- Resource keeps its classification truth in the content database while using
-- Foundation Classification as the deterministic rules authority.

CREATE TABLE resource_classification_catalogs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    catalog_key TEXT NOT NULL UNIQUE,
    revision    BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO resource_classification_catalogs (catalog_key)
VALUES ('resource');

CREATE TABLE resource_classification_policy_profiles (
    catalog_id       UUID NOT NULL REFERENCES resource_classification_catalogs(id) ON DELETE RESTRICT,
    policy_key       TEXT NOT NULL,
    schema_version   SMALLINT NOT NULL CHECK (schema_version = 1),
    policy_revision  BIGINT NOT NULL CHECK (policy_revision > 0),
    category_policy  JSONB NOT NULL,
    facet_policies   JSONB NOT NULL DEFAULT '[]'::jsonb,
    tag_policy       JSONB NOT NULL,
    discovery_policy JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (catalog_id, policy_key),
    CONSTRAINT resource_classification_category_policy_object_check
        CHECK (jsonb_typeof(category_policy) = 'object'),
    CONSTRAINT resource_classification_facet_policies_array_check
        CHECK (jsonb_typeof(facet_policies) = 'array'),
    CONSTRAINT resource_classification_tag_policy_object_check
        CHECK (jsonb_typeof(tag_policy) = 'object'),
    CONSTRAINT resource_classification_discovery_policy_object_check
        CHECK (jsonb_typeof(discovery_policy) = 'object')
);

INSERT INTO resource_classification_policy_profiles (
    catalog_id, policy_key, schema_version, policy_revision,
    category_policy, facet_policies, tag_policy, discovery_policy
)
SELECT id, 'resource.item.default', 1, 1,
       '{"minAssignments":0,"maxAssignments":8,"requirePrimary":false,"leafOnly":false,"maxDepth":0}'::jsonb,
       '[]'::jsonb,
       '{"unknown":"reject","minAssignments":0,"maxAssignments":20}'::jsonb,
       '{"defaultSort":"name_asc"}'::jsonb
FROM resource_classification_catalogs
WHERE catalog_key = 'resource';

ALTER TABLE taxonomies
    ADD COLUMN catalog_id UUID,
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN editorial_position INTEGER,
    ADD COLUMN replacement_id UUID,
    ADD COLUMN first_activated_at TIMESTAMPTZ;

UPDATE taxonomies
SET catalog_id = catalog.id,
    first_activated_at = NOW()
FROM resource_classification_catalogs catalog
WHERE catalog.catalog_key = 'resource';

ALTER TABLE taxonomies
    ALTER COLUMN catalog_id SET NOT NULL,
    ADD CONSTRAINT resource_taxonomy_catalog_fk
        FOREIGN KEY (catalog_id) REFERENCES resource_classification_catalogs(id) ON DELETE RESTRICT,
    ADD CONSTRAINT resource_taxonomy_catalog_identity_key UNIQUE (catalog_id, id),
    ADD CONSTRAINT resource_taxonomy_catalog_kind_identity_key UNIQUE (catalog_id, taxonomy, id),
    ADD CONSTRAINT resource_taxonomy_kind_check CHECK (taxonomy IN ('category', 'tag')),
    ADD CONSTRAINT resource_taxonomy_status_check
        CHECK (status IN ('draft', 'active', 'inactive', 'replaced')),
    ADD CONSTRAINT resource_tag_status_check
        CHECK (taxonomy = 'category' OR status IN ('active', 'inactive', 'replaced')),
    ADD CONSTRAINT resource_taxonomy_editorial_position_check
        CHECK (editorial_position IS NULL OR editorial_position >= 0),
    ADD CONSTRAINT resource_taxonomy_tag_flat_check
        CHECK (taxonomy = 'category' OR parent_id IS NULL),
    ADD CONSTRAINT resource_taxonomy_replacement_state_check CHECK (
        (status = 'replaced' AND replacement_id IS NOT NULL) OR
        (status <> 'replaced' AND replacement_id IS NULL)
    ),
    ADD CONSTRAINT resource_taxonomy_self_replacement_check
        CHECK (replacement_id IS NULL OR replacement_id <> id);

ALTER TABLE taxonomies
    DROP CONSTRAINT taxonomies_parent_id_fkey,
    ADD CONSTRAINT resource_taxonomy_parent_fk
        FOREIGN KEY (catalog_id, taxonomy, parent_id)
        REFERENCES taxonomies(catalog_id, taxonomy, id)
        ON DELETE RESTRICT
        DEFERRABLE INITIALLY IMMEDIATE,
    ADD CONSTRAINT resource_taxonomy_replacement_fk
        FOREIGN KEY (catalog_id, taxonomy, replacement_id)
        REFERENCES taxonomies(catalog_id, taxonomy, id)
        ON DELETE RESTRICT
        DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE object_taxonomies
    DROP CONSTRAINT object_taxonomies_taxonomy_id_fkey,
    ADD CONSTRAINT resource_object_taxonomy_identity_fk
        FOREIGN KEY (taxonomy_id) REFERENCES taxonomies(id) ON DELETE RESTRICT;

CREATE INDEX resource_taxonomies_active_kind_idx
    ON taxonomies (catalog_id, taxonomy, status, id);
CREATE INDEX resource_taxonomies_parent_idx
    ON taxonomies (catalog_id, parent_id)
    WHERE parent_id IS NOT NULL;
CREATE INDEX resource_taxonomies_replacement_idx
    ON taxonomies (catalog_id, replacement_id)
    WHERE replacement_id IS NOT NULL;

CREATE TABLE resource_tag_lookup_entries (
    catalog_id         UUID NOT NULL REFERENCES resource_classification_catalogs(id) ON DELETE RESTRICT,
    taxonomy           TEXT NOT NULL DEFAULT 'tag' CHECK (taxonomy = 'tag'),
    lookup_key         TEXT NOT NULL,
    target_taxonomy_id UUID NOT NULL,
    kind               TEXT NOT NULL CHECK (kind IN ('canonical', 'alias', 'replacement')),
    source_taxonomy_id UUID,
    display_value      TEXT NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (catalog_id, lookup_key),
    FOREIGN KEY (catalog_id, taxonomy, target_taxonomy_id)
        REFERENCES taxonomies(catalog_id, taxonomy, id) ON DELETE RESTRICT,
    FOREIGN KEY (catalog_id, taxonomy, source_taxonomy_id)
        REFERENCES taxonomies(catalog_id, taxonomy, id) ON DELETE RESTRICT,
    CONSTRAINT resource_tag_lookup_source_check CHECK (
        (kind = 'replacement' AND source_taxonomy_id IS NOT NULL) OR
        (kind <> 'replacement' AND source_taxonomy_id IS NULL)
    )
);

CREATE INDEX resource_tag_lookup_target_idx
    ON resource_tag_lookup_entries (catalog_id, target_taxonomy_id);

-- Existing values predate the Go NFKC normalizer. This migration handles the
-- common normalized form; every subsequent create/rename stores the exact
-- lookup key requested by Foundation Classification.
INSERT INTO resource_tag_lookup_entries (
    catalog_id, lookup_key, target_taxonomy_id, kind, display_value
)
SELECT taxonomy.catalog_id,
       lower(regexp_replace(btrim(term.name), '\s+', ' ', 'g')),
       taxonomy.id,
       'canonical',
       term.name
FROM taxonomies taxonomy
JOIN terms term ON term.id = taxonomy.term_id
WHERE taxonomy.taxonomy = 'tag'
ON CONFLICT (catalog_id, lookup_key) DO NOTHING;
