package dao

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"
)

type DiscoveryRow struct {
	Key          string      `orm:"key"`
	Path         string      `orm:"path"`
	Kind         string      `orm:"kind"`
	Title        string      `orm:"title"`
	Description  string      `orm:"description"`
	ImageURL     string      `orm:"image_url"`
	CanonicalURL string      `orm:"canonical_url"`
	Robots       string      `orm:"robots"`
	UpdatedAt    *gtime.Time `orm:"updated_at"`
}

func (p *PG) ListDiscoveryPages(ctx context.Context, origin, afterURL string, limit int) ([]DiscoveryRow, error) {
	var rows []DiscoveryRow
	err := p.db.Ctx(ctx).Raw(`
WITH pages AS (
    SELECT 'resource:' || resource.id::text AS key,
           '/resources/' || resource.id::text AS path,
           'resource'::text AS kind,
           COALESCE(NULLIF(seo.meta_title, ''), resource.title) AS title,
           COALESCE(NULLIF(seo.meta_desc, ''), NULLIF(resource.summary, ''), LEFT(resource.description, 300)) AS description,
           COALESCE(NULLIF(seo.og_image, ''), resource.cover_url, '') AS image_url,
           COALESCE(seo.canonical_url, '') AS canonical_url,
           COALESCE(seo.robots, '') AS robots,
           resource.updated_at
    FROM resources resource
    LEFT JOIN resource_seo seo ON seo.resource_id = resource.id
    WHERE resource.status = 'published'
    UNION ALL
    SELECT taxonomy.taxonomy || ':' || taxonomy.id::text,
           CASE taxonomy.taxonomy
               WHEN 'category' THEN '/category/' || term.slug
               ELSE '/tags/' || term.slug
           END,
           'collection',
           term.name,
           taxonomy.description,
           '',
           '',
           '',
           NULL::timestamptz
    FROM taxonomies taxonomy
    JOIN terms term ON term.id = taxonomy.term_id
)
SELECT * FROM pages
WHERE COALESCE(NULLIF(canonical_url, ''), ? || path) > ?
ORDER BY COALESCE(NULLIF(canonical_url, ''), ? || path) ASC
LIMIT ?`, origin, afterURL, origin, limit).Scan(&rows)
	return rows, err
}
