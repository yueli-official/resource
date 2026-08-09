-- Resource site goes fully free (paid/variation moves to the mall ⑤, per the
-- content-kit-resource decision): drop the access gate and price/points columns.
ALTER TABLE resources DROP COLUMN IF EXISTS access;
ALTER TABLE resources DROP COLUMN IF EXISTS price_cents;
ALTER TABLE resources DROP COLUMN IF EXISTS points_cost;

-- Taxonomy (category/tag), mirrored from the blog service schema; objects are
-- resources. Thin-mode copy (each service owns its own taxonomy tables).
CREATE TABLE terms (
    id   UUID PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL
);
CREATE UNIQUE INDEX uq_terms_slug ON terms (slug);

CREATE TABLE taxonomies (
    id          UUID PRIMARY KEY,
    term_id     UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    taxonomy    TEXT NOT NULL,                                  -- 'category' | 'tag'
    description TEXT NOT NULL DEFAULT '',
    parent_id   UUID REFERENCES taxonomies(id) ON DELETE SET NULL,
    post_count  BIGINT NOT NULL DEFAULT 0,
    extra       JSONB NOT NULL DEFAULT '{}'
);
CREATE UNIQUE INDEX uq_taxonomies_term_taxonomy ON taxonomies (term_id, taxonomy);
CREATE INDEX        ix_taxonomies_taxonomy      ON taxonomies (taxonomy);

CREATE TABLE object_taxonomies (
    object_id   UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    taxonomy_id UUID NOT NULL REFERENCES taxonomies(id) ON DELETE CASCADE,
    sort_order  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (object_id, taxonomy_id)
);
CREATE INDEX ix_object_taxonomies_tax ON object_taxonomies (taxonomy_id);

-- SEO metadata, mirrored from post_seo (1:1 with a resource).
CREATE TABLE resource_seo (
    resource_id     UUID PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    meta_title      TEXT NOT NULL DEFAULT '',
    meta_desc       TEXT NOT NULL DEFAULT '',
    og_title        TEXT NOT NULL DEFAULT '',
    og_image        TEXT NOT NULL DEFAULT '',
    canonical_url   TEXT NOT NULL DEFAULT '',
    robots          TEXT NOT NULL DEFAULT '',
    structured_data JSONB NOT NULL DEFAULT '{}'
);
