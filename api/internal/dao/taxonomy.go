package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/lib/pq"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/foundation/go/identifier"
	"github.com/yueli-official/resource/api/internal/model"
)

type TaxonomyListFilter struct {
	Kind      string
	Q         string
	Sort      string
	Direction string
}

const (
	tTerms                 = "terms"
	tTaxonomies            = "taxonomies"
	tObjTax                = "object_taxonomies"
	tClassificationCatalog = "resource_classification_catalogs"
	tTagLookup             = "resource_tag_lookup_entries"
)

// UpsertTermTaxonomyWithHook creates an active Category or Tag and advances the
// consumer-owned Catalog revision in the same transaction. Existing
// (kind,slug) identities are returned without changing the revision.
func (p *PG) UpsertTermTaxonomyWithHook(
	ctx context.Context,
	name, slug, kind, description, parentID, tagLookupKey string,
	hook TransactionHook,
) (string, error) {
	var taxonomyID string
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var existing *model.Taxonomy
		if err := tx.Model(tTaxonomies+" taxonomy").Ctx(ctx).
			InnerJoin(tTerms+" term", "term.id = taxonomy.term_id").
			Fields("taxonomy.*, term.name, term.slug").
			Where("taxonomy.taxonomy", kind).
			Where("term.slug", slug).
			Where("taxonomy.status <>", classification.StatusReplaced).
			Limit(1).Scan(&existing); err != nil {
			return err
		}
		if existing != nil {
			taxonomyID = existing.ID
			return nil
		}

		catalogID, err := resourceClassificationCatalogID(ctx, tx)
		if err != nil {
			return err
		}
		var term *model.Term
		if err := tx.Model(tTerms).Ctx(ctx).Where("slug", slug).Limit(1).Scan(&term); err != nil {
			return err
		}
		if term != nil {
			return fmt.Errorf("resource classification slug %q is already used by another identity", slug)
		}
		term = &model.Term{ID: identifier.MustNew().String(), Name: strings.TrimSpace(name), Slug: slug}
		if _, err := tx.Model(tTerms).Ctx(ctx).Data(g.Map{
			"id": term.ID, "name": term.Name, "slug": term.Slug,
		}).Insert(); err != nil {
			return err
		}
		taxonomyID = identifier.MustNew().String()
		data := g.Map{
			"id": taxonomyID, "catalog_id": catalogID, "term_id": term.ID,
			"taxonomy": kind, "description": description, "status": classification.StatusActive,
			"first_activated_at": gdb.Raw("NOW()"),
		}
		if parentID != "" {
			data["parent_id"] = parentID
		}
		if _, err := tx.Model(tTaxonomies).Ctx(ctx).Data(data).Insert(); err != nil {
			return err
		}
		if kind == "tag" {
			if strings.TrimSpace(tagLookupKey) == "" {
				return fmt.Errorf("resource canonical tag lookup key is empty")
			}
			if _, err := tx.Model(tTagLookup).Ctx(ctx).Data(g.Map{
				"catalog_id": catalogID, "lookup_key": tagLookupKey,
				"target_taxonomy_id": taxonomyID, "kind": classification.TagMatchCanonical,
				"display_value": term.Name,
			}).Insert(); err != nil {
				return err
			}
		}
		if err := bumpResourceClassificationRevision(ctx, tx, catalogID); err != nil {
			return err
		}
		return runTransactionHook(ctx, tx, hook)
	})
	return taxonomyID, err
}

// GetTaxonomy returns one identity, including replaced governance records.
func (p *PG) GetTaxonomy(ctx context.Context, id string) (*model.Taxonomy, error) {
	var value *model.Taxonomy
	err := p.db.Model(tTaxonomies+" taxonomy").Ctx(ctx).
		InnerJoin(tTerms+" term", "term.id = taxonomy.term_id").
		Fields("taxonomy.*, term.name, term.slug").
		Where("taxonomy.id", id).Limit(1).Scan(&value)
	return value, err
}

func (p *PG) ListTaxonomies(ctx context.Context, kind string) ([]*model.Taxonomy, error) {
	items, _, err := p.ListTaxonomiesPage(ctx, TaxonomyListFilter{Kind: kind}, 0, 0)
	return items, err
}

// ListTaxonomiesPage is the public active candidate projection. Category counts
// include active descendants; Tag counts are direct.
func (p *PG) ListTaxonomiesPage(ctx context.Context, filter TaxonomyListFilter, limit, offset int) ([]*model.Taxonomy, int, error) {
	query := p.db.Model(tTaxonomies+" taxonomy").Ctx(ctx).
		InnerJoin(tTerms+" term", "term.id = taxonomy.term_id").
		Where("taxonomy.status", classification.StatusActive)
	if filter.Kind != "" {
		query = query.Where("taxonomy.taxonomy", filter.Kind)
	}
	if keyword := strings.TrimSpace(filter.Q); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(term.name ILIKE ? OR term.slug ILIKE ? OR taxonomy.description ILIKE ?)", like, like, like)
	}
	total, err := query.Clone().Count()
	if err != nil {
		return nil, 0, err
	}
	query = query.Fields(`
taxonomy.id, taxonomy.catalog_id, taxonomy.term_id, taxonomy.taxonomy,
taxonomy.description, taxonomy.parent_id, taxonomy.status,
taxonomy.editorial_position, taxonomy.replacement_id, term.name, term.slug,
CASE taxonomy.taxonomy
WHEN 'category' THEN (
    WITH RECURSIVE descendants AS (
        SELECT taxonomy.id
        UNION ALL
        SELECT child.id
        FROM taxonomies child
        JOIN descendants parent ON child.parent_id = parent.id
        WHERE child.status = 'active'
    )
    SELECT COUNT(DISTINCT assignment.object_id)
    FROM object_taxonomies assignment
    JOIN resources resource ON resource.id = assignment.object_id
    WHERE assignment.taxonomy_id IN (SELECT id FROM descendants)
      AND resource.status = 'published'
)
ELSE (
    SELECT COUNT(DISTINCT assignment.object_id)
    FROM object_taxonomies assignment
    JOIN resources resource ON resource.id = assignment.object_id
    WHERE assignment.taxonomy_id = taxonomy.id
      AND resource.status = 'published'
)
END AS post_count`).Order(taxonomyListOrder(filter.Sort, filter.Direction))
	if limit > 0 {
		query = query.Limit(offset, limit)
	}
	var items []*model.Taxonomy
	err = query.Scan(&items)
	if items == nil {
		items = []*model.Taxonomy{}
	}
	return items, total, err
}

func taxonomyListOrder(sortBy, direction string) string {
	dir := "ASC"
	if strings.EqualFold(direction, "desc") {
		dir = "DESC"
	}
	switch strings.TrimSpace(sortBy) {
	case "count", "postCount":
		return "post_count " + dir + ", term.name ASC"
	case "slug":
		return "term.slug " + dir + ", term.name ASC"
	default:
		return "term.name " + dir
	}
}

func (p *PG) TaxonomiesByIDs(ctx context.Context, ids []string) ([]*model.Taxonomy, error) {
	if len(ids) == 0 {
		return []*model.Taxonomy{}, nil
	}
	var items []*model.Taxonomy
	err := p.db.Model(tTaxonomies+" taxonomy").Ctx(ctx).
		InnerJoin(tTerms+" term", "term.id = taxonomy.term_id").
		Fields("taxonomy.*, term.name, term.slug").
		WhereIn("taxonomy.id", ids).Scan(&items)
	return items, err
}

func (p *PG) CountTaxonomiesByIDs(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return p.db.Model(tTaxonomies).Ctx(ctx).WhereIn("id", ids).Count()
}

// SetResourceClassification persists only assignments accepted and normalized
// by Foundation Classification.
func (p *PG) SetResourceClassification(ctx context.Context, resourceID string, assignments classification.ClassificationAssignments) error {
	ids := make([]string, 0, len(assignments.Categories)+len(assignments.Tags))
	for _, assignment := range assignments.Categories {
		ids = append(ids, assignment.CategoryID)
	}
	for _, assignment := range assignments.Tags {
		ids = append(ids, assignment.TagID)
	}
	return p.SetResourceTaxonomies(ctx, resourceID, ids)
}

func (p *PG) SetResourceTaxonomies(ctx context.Context, resourceID string, taxonomyIDs []string) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(tObjTax).Ctx(ctx).Where("object_id", resourceID).Delete(); err != nil {
			return err
		}
		for index, taxonomyID := range taxonomyIDs {
			if _, err := tx.Model(tObjTax).Ctx(ctx).Data(g.Map{
				"object_id": resourceID, "taxonomy_id": taxonomyID, "sort_order": index,
			}).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

// ResourceIDsByTaxonomySlug resolves replacement and expands active Category
// descendants. Tags remain flat.
func (p *PG) ResourceIDsByTaxonomySlug(ctx context.Context, slug string) ([]string, error) {
	var rows []struct {
		ID string `orm:"id"`
	}
	err := p.db.Ctx(ctx).Raw(`
WITH RECURSIVE requested AS (
    SELECT taxonomy.id, taxonomy.taxonomy, taxonomy.replacement_id
    FROM taxonomies taxonomy
    JOIN terms term ON term.id = taxonomy.term_id
    WHERE term.slug = ?
), root AS (
    SELECT COALESCE(replacement_id, id) AS id, taxonomy
    FROM requested
), matches AS (
    SELECT id, taxonomy FROM root
    UNION ALL
    SELECT child.id, child.taxonomy
    FROM taxonomies child
    JOIN matches parent ON child.parent_id = parent.id
    WHERE parent.taxonomy = 'category' AND child.status = 'active'
)
SELECT DISTINCT assignment.object_id::text AS id
FROM object_taxonomies assignment
JOIN matches ON matches.id = assignment.taxonomy_id
ORDER BY id`, slug).Scan(&rows)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (p *PG) GetTaxonomyBySlug(ctx context.Context, kind, slug string) (*model.Taxonomy, error) {
	var value *model.Taxonomy
	query := p.db.Model(tTaxonomies+" taxonomy").Ctx(ctx).
		InnerJoin(tTerms+" term", "term.id = taxonomy.term_id").
		Fields("taxonomy.*, term.name, term.slug").
		Where("term.slug", slug).
		Where("taxonomy.status", classification.StatusActive)
	if kind != "" {
		query = query.Where("taxonomy.taxonomy", kind)
	}
	err := query.Limit(1).Scan(&value)
	return value, err
}

func (p *PG) GetResourceTaxonomies(ctx context.Context, resourceID string) ([]*model.Taxonomy, error) {
	var items []*model.Taxonomy
	err := p.db.Model(tObjTax+" assignment").Ctx(ctx).
		InnerJoin(tTaxonomies+" taxonomy", "taxonomy.id = assignment.taxonomy_id").
		InnerJoin(tTerms+" term", "term.id = taxonomy.term_id").
		Fields("taxonomy.*, term.name, term.slug").
		Where("assignment.object_id", resourceID).
		OrderAsc("assignment.sort_order").Scan(&items)
	if items == nil {
		items = []*model.Taxonomy{}
	}
	return items, err
}

func (p *PG) TaxonomyResourceCounts(ctx context.Context) (map[string]int, error) {
	var rows []struct {
		TaxonomyID string `orm:"taxonomy_id"`
		Count      int    `orm:"count"`
	}
	err := p.db.Model(tObjTax+" assignment").Ctx(ctx).
		InnerJoin(tResources+" resource", "resource.id = assignment.object_id").
		Where("resource.status", model.StatusPublished).
		Fields("assignment.taxonomy_id, COUNT(DISTINCT assignment.object_id) AS count").
		Group("assignment.taxonomy_id").Scan(&rows)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.TaxonomyID] = row.Count
	}
	return counts, nil
}

func (p *PG) TaxonomyChildCount(ctx context.Context, id string) (int, error) {
	return p.db.Model(tTaxonomies).Ctx(ctx).Where("parent_id", id).Count()
}

func (p *PG) TaxonomyAssignmentCount(ctx context.Context, id string) (int, error) {
	return p.db.Model(tObjTax).Ctx(ctx).Where("taxonomy_id", id).Count()
}

func (p *PG) UpdateTaxonomyWithHook(
	ctx context.Context,
	current *model.Taxonomy,
	termFields, taxonomyFields g.Map,
	tagLookupKey string,
	hook TransactionHook,
) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if len(termFields) > 0 {
			if _, err := tx.Model(tTerms).Ctx(ctx).Where("id", current.TermID).Data(termFields).Update(); err != nil {
				return err
			}
		}
		if len(taxonomyFields) > 0 {
			if _, err := tx.Model(tTaxonomies).Ctx(ctx).Where("id", current.ID).Data(taxonomyFields).Update(); err != nil {
				return err
			}
		}
		if current.Taxonomy == "tag" && tagLookupKey != "" {
			var existing *struct {
				TargetID string `orm:"target_taxonomy_id"`
			}
			if err := tx.Model(tTagLookup).Ctx(ctx).
				Fields("target_taxonomy_id").
				Where("catalog_id", current.CatalogID).
				Where("lookup_key", tagLookupKey).
				Limit(1).Scan(&existing); err != nil {
				return err
			}
			if existing != nil && existing.TargetID != current.ID {
				return fmt.Errorf("resource tag lookup key already resolves to another identity")
			}
			if _, err := tx.Model(tTagLookup).Ctx(ctx).
				Where("catalog_id", current.CatalogID).
				Where("target_taxonomy_id", current.ID).
				Where("kind", classification.TagMatchCanonical).
				Where("lookup_key <>", tagLookupKey).
				Data(g.Map{"kind": classification.TagMatchAlias}).
				Update(); err != nil {
				return err
			}
			display := current.Name
			if name, ok := termFields["name"].(string); ok {
				display = name
			}
			if existing == nil {
				if _, err := tx.Model(tTagLookup).Ctx(ctx).Data(g.Map{
					"catalog_id": current.CatalogID, "lookup_key": tagLookupKey,
					"target_taxonomy_id": current.ID, "kind": classification.TagMatchCanonical,
					"display_value": display,
				}).Insert(); err != nil {
					return err
				}
			} else if _, err := tx.Model(tTagLookup).Ctx(ctx).
				Where("catalog_id", current.CatalogID).
				Where("lookup_key", tagLookupKey).
				Data(g.Map{
					"kind": classification.TagMatchCanonical, "source_taxonomy_id": nil,
					"display_value": display,
				}).Update(); err != nil {
				return err
			}
		}
		if err := bumpResourceClassificationRevision(ctx, tx, current.CatalogID); err != nil {
			return err
		}
		return runTransactionHook(ctx, tx, hook)
	})
}

func (p *PG) DeleteTaxonomyWithHook(ctx context.Context, current *model.Taxonomy, hook TransactionHook) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if current.Taxonomy == "tag" {
			if _, err := tx.Model(tTagLookup).Ctx(ctx).
				Where("target_taxonomy_id", current.ID).
				WhereOr("source_taxonomy_id", current.ID).
				Delete(); err != nil {
				return err
			}
		}
		if _, err := tx.Model(tTaxonomies).Ctx(ctx).Where("id", current.ID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model(tTerms).Ctx(ctx).Where("id", current.TermID).Delete(); err != nil {
			return err
		}
		if err := bumpResourceClassificationRevision(ctx, tx, current.CatalogID); err != nil {
			return err
		}
		return runTransactionHook(ctx, tx, hook)
	})
}

// MergeTaxonomyWithHook preserves source identity as replaced, migrates all
// assignments and direct children, and advances the Catalog revision.
func (p *PG) MergeTaxonomyWithHook(ctx context.Context, source, target *model.Taxonomy, hook TransactionHook) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO object_taxonomies (object_id, taxonomy_id, sort_order)
SELECT object_id, ?::uuid, sort_order
FROM object_taxonomies
WHERE taxonomy_id = ?::uuid
ON CONFLICT (object_id, taxonomy_id) DO NOTHING`, target.ID, source.ID); err != nil {
			return err
		}
		if _, err := tx.Model(tObjTax).Ctx(ctx).Where("taxonomy_id", source.ID).Delete(); err != nil {
			return err
		}
		if source.Taxonomy == "category" {
			if _, err := tx.Model(tTaxonomies).Ctx(ctx).
				Where("parent_id", source.ID).
				Data(g.Map{"parent_id": target.ID}).
				Update(); err != nil {
				return err
			}
		} else {
			if _, err := tx.Ctx(ctx).Exec(`
UPDATE resource_tag_lookup_entries
SET target_taxonomy_id = ?::uuid
WHERE target_taxonomy_id = ?::uuid AND kind = 'alias'`, target.ID, source.ID); err != nil {
				return err
			}
			if _, err := tx.Ctx(ctx).Exec(`
UPDATE resource_tag_lookup_entries
SET target_taxonomy_id = ?::uuid,
    source_taxonomy_id = ?::uuid,
    kind = 'replacement'
WHERE target_taxonomy_id = ?::uuid AND kind = 'canonical'`, target.ID, source.ID, source.ID); err != nil {
				return err
			}
		}
		if _, err := tx.Model(tTaxonomies).Ctx(ctx).Where("replacement_id", source.ID).
			Data(g.Map{"replacement_id": target.ID}).Update(); err != nil {
			return err
		}
		if _, err := tx.Model(tTaxonomies).Ctx(ctx).Where("id", source.ID).Data(g.Map{
			"parent_id": nil, "status": classification.StatusReplaced,
			"replacement_id": target.ID,
		}).Update(); err != nil {
			return err
		}
		if err := bumpResourceClassificationRevision(ctx, tx, source.CatalogID); err != nil {
			return err
		}
		return runTransactionHook(ctx, tx, hook)
	})
}

func (p *PG) ClassificationSnapshot(ctx context.Context) (classification.Snapshot, error) {
	type catalogRow struct {
		ID       string `orm:"id"`
		Revision uint64 `orm:"revision"`
	}
	type policyRow struct {
		Key             string `orm:"policy_key"`
		SchemaVersion   uint16 `orm:"schema_version"`
		PolicyRevision  uint64 `orm:"policy_revision"`
		CategoryPolicy  string `orm:"category_policy"`
		FacetPolicies   string `orm:"facet_policies"`
		TagPolicy       string `orm:"tag_policy"`
		DiscoveryPolicy string `orm:"discovery_policy"`
	}
	var snapshot classification.Snapshot
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		scoped := tx.Ctx(ctx)
		if _, err := scoped.Exec(`SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
			return err
		}
		var catalog *catalogRow
		if err := scoped.Raw(`
SELECT id::text AS id, revision
FROM resource_classification_catalogs
WHERE catalog_key = 'resource'`).Scan(&catalog); err != nil {
			return err
		}
		if catalog == nil {
			return fmt.Errorf("resource classification catalog is not seeded")
		}
		snapshot = classification.Snapshot{
			CatalogID: catalog.ID, Revision: catalog.Revision,
			Facets: []classification.Facet{}, FacetValues: []classification.FacetValue{},
		}
		if err := scoped.Raw(`
SELECT taxonomy.id::text AS id, COALESCE(taxonomy.parent_id::text, '') AS parent_id,
       term.slug, term.name, taxonomy.status, taxonomy.editorial_position,
       COALESCE(taxonomy.replacement_id::text, '') AS replacement_id
FROM taxonomies taxonomy
JOIN terms term ON term.id = taxonomy.term_id
WHERE taxonomy.catalog_id = ?::uuid AND taxonomy.taxonomy = 'category'
ORDER BY taxonomy.id`, catalog.ID).Scan(&snapshot.Categories); err != nil {
			return err
		}
		var policies []policyRow
		if err := scoped.Raw(`
SELECT policy_key, schema_version, policy_revision,
       category_policy::text AS category_policy,
       facet_policies::text AS facet_policies,
       tag_policy::text AS tag_policy,
       discovery_policy::text AS discovery_policy
FROM resource_classification_policy_profiles
WHERE catalog_id = ?::uuid
ORDER BY policy_key`, catalog.ID).Scan(&policies); err != nil {
			return err
		}
		for _, row := range policies {
			policy := classification.PolicyProfile{
				Key: row.Key, SchemaVersion: row.SchemaVersion, PolicyRevision: row.PolicyRevision,
			}
			if err := json.Unmarshal([]byte(row.CategoryPolicy), &policy.Category); err != nil {
				return err
			}
			if err := json.Unmarshal([]byte(row.FacetPolicies), &policy.Facets); err != nil {
				return err
			}
			if err := json.Unmarshal([]byte(row.TagPolicy), &policy.Tags); err != nil {
				return err
			}
			if err := json.Unmarshal([]byte(row.DiscoveryPolicy), &policy.Discovery); err != nil {
				return err
			}
			snapshot.Policies = append(snapshot.Policies, policy)
		}
		return nil
	})
	return snapshot, err
}

func (p *PG) ClassificationTagMatches(ctx context.Context, lookups []classification.TagLookupRequest) ([]classification.TagMatch, string, error) {
	if len(lookups) == 0 {
		return []classification.TagMatch{}, "", nil
	}
	keys := make([]string, 0, len(lookups))
	for _, lookup := range lookups {
		keys = append(keys, lookup.LookupKey)
	}
	value, err := pq.StringArray(keys).Value()
	if err != nil {
		return nil, "", err
	}
	var rows []struct {
		LookupKey string `orm:"lookup_key"`
		Kind      string `orm:"kind"`
		TagID     string `orm:"tag_id"`
		Revision  uint64 `orm:"revision"`
	}
	if err := p.db.Ctx(ctx).Raw(`
WITH requested AS (
    SELECT lookup_key, ordinal
    FROM unnest(?::text[]) WITH ORDINALITY AS value(lookup_key, ordinal)
), catalog AS (
    SELECT id, revision
    FROM resource_classification_catalogs
    WHERE catalog_key = 'resource'
)
SELECT requested.lookup_key,
       CASE
           WHEN target.id IS NULL THEN 'not_found'
           WHEN target.status = 'inactive' THEN 'inactive'
           WHEN target.status = 'replaced' THEN 'replacement'
           ELSE entry.kind
       END AS kind,
       COALESCE(
           CASE WHEN target.status = 'replaced' THEN target.replacement_id ELSE target.id END::text,
           ''
       ) AS tag_id,
       catalog.revision
FROM requested
CROSS JOIN catalog
LEFT JOIN resource_tag_lookup_entries entry
       ON entry.catalog_id = catalog.id AND entry.lookup_key = requested.lookup_key
LEFT JOIN taxonomies target ON target.id = entry.target_taxonomy_id
ORDER BY requested.ordinal`, value).Scan(&rows); err != nil {
		return nil, "", err
	}
	matches := make([]classification.TagMatch, 0, len(rows))
	var revision uint64
	for _, row := range rows {
		revision = row.Revision
		matches = append(matches, classification.TagMatch{
			LookupKey: row.LookupKey,
			Kind:      classification.TagMatchKind(row.Kind),
			TagID:     row.TagID,
		})
	}
	return matches, fmt.Sprintf("%d", revision), nil
}

func resourceClassificationCatalogID(ctx context.Context, tx gdb.TX) (string, error) {
	value, err := tx.Ctx(ctx).GetValue(`
SELECT id::text
FROM resource_classification_catalogs
WHERE catalog_key = 'resource'`)
	if err != nil {
		return "", err
	}
	if value.IsEmpty() {
		return "", fmt.Errorf("resource classification catalog is not seeded")
	}
	return value.String(), nil
}

func bumpResourceClassificationRevision(ctx context.Context, tx gdb.TX, catalogID string) error {
	_, err := tx.Ctx(ctx).Exec(`
UPDATE resource_classification_catalogs
SET revision = revision + 1, updated_at = NOW()
WHERE id = ?::uuid`, catalogID)
	return err
}
