package catalog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/model"
	"platform/products/resource/api/internal/reserr"
)

// CreateTaxonomy upserts a term and its taxonomy (category/tag).
func (s *Service) CreateTaxonomy(ctx context.Context, name, explicitSlug, kind, parentID, description string) (*model.Taxonomy, error) {
	slug := slugify(explicitSlug)
	if slug == "" {
		slug = slugify(name)
	}
	if slug == "" {
		return nil, reserr.InvalidInput("name produces an empty slug")
	}
	termID, err := s.dao.UpsertTerm(ctx, name, slug)
	if err != nil {
		return nil, err
	}
	id, err := s.dao.UpsertTaxonomy(ctx, termID, kind, description, parentID)
	if err != nil {
		return nil, err
	}
	return s.dao.GetTaxonomy(ctx, id)
}

// ListTaxonomies returns taxonomies (optionally filtered by kind), each carrying
// its published-resource count (for tag-cloud sizing + governance display).
func (s *Service) ListTaxonomies(ctx context.Context, kind string) ([]*model.Taxonomy, error) {
	items, _, _, _, err := s.ListTaxonomiesPage(ctx, kind, "", "name", "asc", 0, 0)
	return items, err
}

func (s *Service) ListTaxonomiesPage(ctx context.Context, kind, q, sortBy, direction string, page, size int) ([]*model.Taxonomy, int, int, int, error) {
	limit, offset := 0, 0
	if size > 0 {
		page, size = norm(page, size)
		limit, offset = size, (page-1)*size
	}
	items, total, err := s.dao.ListTaxonomiesPage(ctx, dao.TaxonomyListFilter{
		Kind: kind, Q: q, Sort: sortBy, Direction: direction,
	}, limit, offset)
	return items, total, page, size, err
}

// GetTaxonomyBySlug resolves a (kind, slug) to a taxonomy (filter page header).
func (s *Service) GetTaxonomyBySlug(ctx context.Context, kind, slug string) (*model.Taxonomy, error) {
	return s.dao.GetTaxonomyBySlug(ctx, kind, slug)
}

// GetResourceTaxonomies returns a resource's assigned taxonomies (detail tags +
// editor selection state).
func (s *Service) GetResourceTaxonomies(ctx context.Context, resourceID string) ([]*model.Taxonomy, error) {
	return s.dao.GetResourceTaxonomies(ctx, resourceID)
}

// UpdateTaxonomy renames / re-slugs / re-describes / re-parents a taxonomy (admin
// governance). Nil pointers are left unchanged; a non-nil empty parentId clears
// the parent. Self-parenting is rejected (a deeper cycle is admin-only, low risk).
func (s *Service) UpdateTaxonomy(ctx context.Context, id string, name, slug, description, parentID *string) (*model.Taxonomy, error) {
	cur, err := s.dao.GetTaxonomy(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, reserr.NotFound(id)
	}
	termFields := g.Map{}
	if name != nil && *name != "" {
		termFields["name"] = *name
	}
	if slug != nil {
		sl := slugify(*slug)
		if sl == "" {
			return nil, reserr.InvalidInput("slug produces an empty value")
		}
		termFields["slug"] = sl
	}
	taxFields := g.Map{}
	if description != nil {
		taxFields["description"] = *description
	}
	if parentID != nil {
		switch *parentID {
		case id:
			return nil, reserr.InvalidInput("a taxonomy cannot be its own parent")
		case "":
			taxFields["parent_id"] = nil
		default:
			taxFields["parent_id"] = *parentID
		}
	}
	if len(termFields) == 0 && len(taxFields) == 0 {
		return cur, nil
	}
	if err := s.dao.UpdateTaxonomy(ctx, id, cur.TermID, termFields, taxFields); err != nil {
		return nil, err
	}
	return s.dao.GetTaxonomy(ctx, id)
}

// DeleteTaxonomy removes a taxonomy; refused while it still has child taxonomies
// (reparent or remove them first) to avoid orphaning a subtree.
func (s *Service) DeleteTaxonomy(ctx context.Context, id string) error {
	n, err := s.dao.TaxonomyChildCount(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return reserr.InvalidState("taxonomy has child taxonomies; reparent or remove them first")
	}
	return s.dao.DeleteTaxonomy(ctx, id)
}

// MergeTaxonomy folds source into target: resources move to target, source's
// children re-parent to target, then source is deleted. Both must exist and differ.
func (s *Service) MergeTaxonomy(ctx context.Context, sourceID, targetID string) error {
	if sourceID == targetID {
		return reserr.InvalidInput("cannot merge a taxonomy into itself")
	}
	n, err := s.dao.CountTaxonomiesByIDs(ctx, []string{sourceID, targetID})
	if err != nil {
		return err
	}
	if n != 2 {
		return reserr.NotFound(sourceID)
	}
	return s.dao.MergeTaxonomy(ctx, sourceID, targetID)
}

// AssignTaxonomies replaces a resource's taxonomy assignments (owner-gated).
// Unknown taxonomy ids are rejected up front (otherwise the FK violation
// surfaces as 500).
func (s *Service) AssignTaxonomies(ctx context.Context, owner, resourceID string, taxIDs []string) error {
	r, err := s.dao.GetByID(ctx, resourceID)
	if err != nil {
		return err
	}
	if r == nil || r.OwnerID != owner {
		return reserr.NotFound(resourceID)
	}
	if len(taxIDs) > 0 {
		n, err := s.dao.CountTaxonomiesByIDs(ctx, taxIDs)
		if err != nil {
			return err
		}
		if n != len(taxIDs) {
			return reserr.InvalidInput("one or more taxonomy ids do not exist")
		}
	}
	return s.dao.SetResourceTaxonomies(ctx, resourceID, taxIDs)
}
