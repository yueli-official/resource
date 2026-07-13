package dao

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"

	"platform/products/resource/api/internal/model"
)

type TaxonomyListFilter struct {
	Kind      string
	Q         string
	Sort      string
	Direction string
}

const (
	tTerms      = "terms"
	tTaxonomies = "taxonomies"
	tObjTax     = "object_taxonomies"
)

// UpsertTerm returns the id of the term with this slug, inserting it if absent.
func (p *PG) UpsertTerm(ctx context.Context, name, slug string) (string, error) {
	var t *model.Term
	if err := p.db.Model(tTerms).Ctx(ctx).Where("slug", slug).Limit(1).Scan(&t); err != nil {
		return "", err
	}
	if t != nil {
		return t.ID, nil
	}
	id := uuid.NewString()
	if _, err := p.db.Model(tTerms).Ctx(ctx).Data(g.Map{"id": id, "name": name, "slug": slug}).Insert(); err != nil {
		// lost a race? the slug now exists — re-read.
		var t2 *model.Term
		if e := p.db.Model(tTerms).Ctx(ctx).Where("slug", slug).Limit(1).Scan(&t2); e == nil && t2 != nil {
			return t2.ID, nil
		}
		return "", err
	}
	return id, nil
}

// UpsertTaxonomy returns the id of the (term, kind) taxonomy, inserting if absent.
func (p *PG) UpsertTaxonomy(ctx context.Context, termID, kind, description, parentID string) (string, error) {
	var tx *model.Taxonomy
	if err := p.db.Model(tTaxonomies).Ctx(ctx).
		Where("term_id", termID).Where("taxonomy", kind).Limit(1).Scan(&tx); err != nil {
		return "", err
	}
	if tx != nil {
		return tx.ID, nil
	}
	id := uuid.NewString()
	data := g.Map{"id": id, "term_id": termID, "taxonomy": kind, "description": description}
	if parentID != "" {
		data["parent_id"] = parentID
	}
	if _, err := p.db.Model(tTaxonomies).Ctx(ctx).Data(data).Insert(); err != nil {
		return "", err
	}
	return id, nil
}

// GetTaxonomy returns one taxonomy joined with its term (name/slug), or (nil,nil).
func (p *PG) GetTaxonomy(ctx context.Context, id string) (*model.Taxonomy, error) {
	var tx *model.Taxonomy
	err := p.db.Model(tTaxonomies+" tx").Ctx(ctx).
		LeftJoin(tTerms+" t", "t.id = tx.term_id").
		Fields("tx.*, t.name, t.slug").
		Where("tx.id", id).Limit(1).Scan(&tx)
	return tx, err
}

// ListTaxonomies returns taxonomies (optionally filtered by kind), joined with
// their term name/slug, ordered by name.
func (p *PG) ListTaxonomies(ctx context.Context, kind string) ([]*model.Taxonomy, error) {
	items, _, err := p.ListTaxonomiesPage(ctx, TaxonomyListFilter{Kind: kind}, 0, 0)
	return items, err
}

func (p *PG) ListTaxonomiesPage(ctx context.Context, filter TaxonomyListFilter, limit, offset int) ([]*model.Taxonomy, int, error) {
	m := p.db.Model(tTaxonomies+" tx").Ctx(ctx).
		LeftJoin(tTerms+" t", "t.id = tx.term_id")
	if filter.Kind != "" {
		m = m.Where("tx.taxonomy", filter.Kind)
	}
	if keyword := strings.TrimSpace(filter.Q); keyword != "" {
		like := "%" + keyword + "%"
		m = m.Where("(t.name ILIKE ? OR t.slug ILIKE ? OR tx.description ILIKE ?)", like, like, like)
	}
	total, err := m.Count()
	if err != nil {
		return nil, 0, err
	}
	m = m.Fields(`tx.id, tx.term_id, tx.taxonomy, tx.description, tx.parent_id, t.name, t.slug,
(SELECT COUNT(*) FROM object_taxonomies ot
 JOIN resources r ON r.id = ot.object_id
 WHERE ot.taxonomy_id = tx.id AND r.status = 'published') AS post_count`).
		Order(taxonomyListOrder(filter.Sort, filter.Direction))
	if limit > 0 {
		m = m.Limit(offset, limit)
	}
	var out []*model.Taxonomy
	err = m.Scan(&out)
	return out, total, err
}

func taxonomyListOrder(sortBy, direction string) string {
	dir := "ASC"
	if strings.EqualFold(direction, "desc") {
		dir = "DESC"
	}
	switch strings.TrimSpace(sortBy) {
	case "count", "postCount":
		return "post_count " + dir + ", t.name ASC"
	case "slug":
		return "t.slug " + dir + ", t.name ASC"
	default:
		return "t.name " + dir
	}
}

// CountTaxonomiesByIDs returns how many of the given taxonomy ids exist.
func (p *PG) CountTaxonomiesByIDs(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return p.db.Model(tTaxonomies).Ctx(ctx).WhereIn("id", ids).Count()
}

// SetResourceTaxonomies replaces a resource's taxonomy assignments (delete then insert).
func (p *PG) SetResourceTaxonomies(ctx context.Context, resourceID string, taxIDs []string) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(tObjTax).Ctx(ctx).Where("object_id", resourceID).Delete(); err != nil {
			return err
		}
		for i, tid := range taxIDs {
			if _, err := tx.Model(tObjTax).Ctx(ctx).Data(g.Map{
				"object_id": resourceID, "taxonomy_id": tid, "sort_order": i,
			}).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

// ResourceIDsByTaxonomySlug returns the ids of resources tagged with the term slug.
func (p *PG) ResourceIDsByTaxonomySlug(ctx context.Context, slug string) ([]string, error) {
	vals, err := p.db.Model(tObjTax+" ot").Ctx(ctx).
		LeftJoin(tTaxonomies+" tx", "tx.id = ot.taxonomy_id").
		LeftJoin(tTerms+" t", "t.id = tx.term_id").
		Where("t.slug", slug).
		Fields("ot.object_id").Array()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(vals))
	for _, v := range vals {
		ids = append(ids, v.String())
	}
	return ids, nil
}

// GetTaxonomyBySlug returns the taxonomy of the given kind whose term has this
// slug, joined with the term name/slug, or (nil, nil) when absent.
func (p *PG) GetTaxonomyBySlug(ctx context.Context, kind, slug string) (*model.Taxonomy, error) {
	var tx *model.Taxonomy
	err := p.db.Model(tTaxonomies+" tx").Ctx(ctx).
		LeftJoin(tTerms+" t", "t.id = tx.term_id").
		Fields("tx.*, t.name, t.slug").
		Where("tx.taxonomy", kind).Where("t.slug", slug).Limit(1).Scan(&tx)
	return tx, err
}

// GetResourceTaxonomies returns a resource's assigned taxonomies, joined with term
// name/slug, ordered by the saved sort_order. Never returns nil.
func (p *PG) GetResourceTaxonomies(ctx context.Context, resourceID string) ([]*model.Taxonomy, error) {
	var out []*model.Taxonomy
	err := p.db.Model(tObjTax+" ot").Ctx(ctx).
		LeftJoin(tTaxonomies+" tx", "tx.id = ot.taxonomy_id").
		LeftJoin(tTerms+" t", "t.id = tx.term_id").
		Fields("tx.*, t.name, t.slug").
		Where("ot.object_id", resourceID).
		OrderAsc("ot.sort_order").Scan(&out)
	if out == nil {
		out = []*model.Taxonomy{}
	}
	return out, err
}

// TaxonomyResourceCounts returns published-resource counts keyed by taxonomy id
// (for tag-cloud sizing + governance display). Drafts/archived are excluded.
func (p *PG) TaxonomyResourceCounts(ctx context.Context) (map[string]int, error) {
	var rows []struct {
		TaxonomyID string `orm:"taxonomy_id"`
		C          int    `orm:"c"`
	}
	err := p.db.Model(tObjTax+" ot").Ctx(ctx).
		LeftJoin(tResources+" r", "r.id = ot.object_id").
		Where("r.status", string(model.StatusPublished)).
		Fields("ot.taxonomy_id, COUNT(*) AS c").
		Group("ot.taxonomy_id").Scan(&rows)
	if err != nil {
		return nil, err
	}
	m := make(map[string]int, len(rows))
	for _, r := range rows {
		m[r.TaxonomyID] = r.C
	}
	return m, nil
}

// TaxonomyChildCount returns how many taxonomies have this one as their parent.
func (p *PG) TaxonomyChildCount(ctx context.Context, id string) (int, error) {
	return p.db.Model(tTaxonomies).Ctx(ctx).Where("parent_id", id).Count()
}

// UpdateTaxonomy applies term-level fields (name/slug) and taxonomy-level fields
// (description/parent_id) in one transaction; either map may be empty.
func (p *PG) UpdateTaxonomy(ctx context.Context, id, termID string, termFields, taxFields g.Map) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if len(termFields) > 0 {
			if _, err := tx.Model(tTerms).Ctx(ctx).Where("id", termID).Data(termFields).Update(); err != nil {
				return err
			}
		}
		if len(taxFields) > 0 {
			if _, err := tx.Model(tTaxonomies).Ctx(ctx).Where("id", id).Data(taxFields).Update(); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteTaxonomy removes a taxonomy (its object_taxonomies rows cascade via FK).
// The shared term row is left intact — it may back another kind and is reused by
// slug on the next CreateTaxonomy.
func (p *PG) DeleteTaxonomy(ctx context.Context, id string) error {
	_, err := p.db.Model(tTaxonomies).Ctx(ctx).Where("id", id).Delete()
	return err
}

// MergeTaxonomy re-points every resource tagged with source onto target (skipping
// resources that already carry target, to avoid the (object_id, taxonomy_id) PK
// clash), re-parents source's children onto target, then deletes source.
func (p *PG) MergeTaxonomy(ctx context.Context, sourceID, targetID string) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		srcVals, err := tx.Model(tObjTax).Ctx(ctx).Where("taxonomy_id", sourceID).Fields("object_id").Array()
		if err != nil {
			return err
		}
		tgtVals, err := tx.Model(tObjTax).Ctx(ctx).Where("taxonomy_id", targetID).Fields("object_id").Array()
		if err != nil {
			return err
		}
		have := make(map[string]bool, len(tgtVals))
		for _, v := range tgtVals {
			have[v.String()] = true
		}
		if _, err := tx.Model(tObjTax).Ctx(ctx).Where("taxonomy_id", sourceID).Delete(); err != nil {
			return err
		}
		for _, v := range srcVals {
			oid := v.String()
			if have[oid] {
				continue
			}
			if _, err := tx.Model(tObjTax).Ctx(ctx).Data(g.Map{
				"object_id": oid, "taxonomy_id": targetID, "sort_order": 0,
			}).Insert(); err != nil {
				return err
			}
		}
		if _, err := tx.Model(tTaxonomies).Ctx(ctx).Where("parent_id", sourceID).Data(g.Map{"parent_id": targetID}).Update(); err != nil {
			return err
		}
		if _, err := tx.Model(tTaxonomies).Ctx(ctx).Where("id", sourceID).Delete(); err != nil {
			return err
		}
		return nil
	})
}
