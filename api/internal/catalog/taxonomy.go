package catalog

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/resource/api/internal/dao"
	"github.com/yueli-official/resource/api/internal/model"
	"github.com/yueli-official/resource/api/internal/rescause"
)

const resourceItemPolicyKey = "resource.item.default"

func (s *Service) CreateTaxonomy(ctx context.Context, name, explicitSlug, kind, parentID, description string) (*model.Taxonomy, error) {
	if kind != "category" && kind != "tag" {
		return nil, rescause.InvalidInput("taxonomy must be category or tag")
	}
	if kind == "tag" && strings.TrimSpace(parentID) != "" {
		return nil, rescause.InvalidInput("tags are flat and cannot have a parent")
	}
	slug := slugify(explicitSlug)
	if slug == "" {
		slug = slugify(name)
	}
	if slug == "" {
		return nil, rescause.InvalidInput("name produces an empty slug")
	}
	if existing, err := s.dao.GetTaxonomyBySlug(ctx, kind, slug); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}
	if kind == "category" && parentID != "" {
		parent, err := s.dao.GetTaxonomy(ctx, parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil || parent.Taxonomy != "category" || parent.Status != string(classification.StatusActive) {
			return nil, rescause.InvalidInput("category parent must be an active category")
		}
	}
	tagLookupKey := ""
	if kind == "tag" {
		lookup, err := s.classificationTagLookup(ctx, name)
		if err != nil {
			return nil, err
		}
		matches, _, err := s.dao.ClassificationTagMatches(ctx, []classification.TagLookupRequest{lookup})
		if err != nil {
			return nil, err
		}
		if len(matches) == 1 {
			switch matches[0].Kind {
			case classification.TagMatchCanonical, classification.TagMatchAlias, classification.TagMatchReplacement:
				return s.dao.GetTaxonomy(ctx, matches[0].TagID)
			case classification.TagMatchInactive:
				return nil, rescause.InvalidState("tag is inactive")
			}
		}
		tagLookupKey = lookup.LookupKey
	}
	id, err := s.dao.UpsertTermTaxonomyWithHook(
		ctx, strings.TrimSpace(name), slug, kind, description, parentID, tagLookupKey,
		s.taxonomyURLHook("resource taxonomy created"),
	)
	if err != nil {
		return nil, err
	}
	return s.dao.GetTaxonomy(ctx, id)
}

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

func (s *Service) GetTaxonomyBySlug(ctx context.Context, kind, slug string) (*model.Taxonomy, error) {
	return s.dao.GetTaxonomyBySlug(ctx, kind, slug)
}

func (s *Service) GetResourceTaxonomies(ctx context.Context, resourceID string) ([]*model.Taxonomy, error) {
	return s.dao.GetResourceTaxonomies(ctx, resourceID)
}

func (s *Service) UpdateTaxonomy(ctx context.Context, id string, name, slug, description, parentID *string) (*model.Taxonomy, error) {
	current, err := s.dao.GetTaxonomy(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil || current.Status == string(classification.StatusReplaced) {
		return nil, rescause.NotFound(id)
	}
	termFields := g.Map{}
	taxonomyFields := g.Map{}
	tagLookupKey := ""
	if name != nil && strings.TrimSpace(*name) != "" {
		normalizedName := strings.TrimSpace(*name)
		termFields["name"] = normalizedName
		if current.Taxonomy == "tag" {
			lookup, err := s.classificationTagLookup(ctx, normalizedName)
			if err != nil {
				return nil, err
			}
			matches, _, err := s.dao.ClassificationTagMatches(ctx, []classification.TagLookupRequest{lookup})
			if err != nil {
				return nil, err
			}
			if len(matches) == 1 && matches[0].Kind != classification.TagMatchNotFound && matches[0].TagID != current.ID {
				return nil, rescause.InvalidInput("tag name already resolves to another canonical tag")
			}
			tagLookupKey = lookup.LookupKey
		}
	}
	if slug != nil {
		normalized := slugify(*slug)
		if normalized == "" {
			return nil, rescause.InvalidInput("slug produces an empty value")
		}
		termFields["slug"] = normalized
	}
	if description != nil {
		taxonomyFields["description"] = *description
	}
	if parentID != nil {
		if current.Taxonomy == "tag" && *parentID != "" {
			return nil, rescause.InvalidInput("tags are flat and cannot have a parent")
		}
		switch *parentID {
		case id:
			return nil, rescause.InvalidInput("a category cannot be its own parent")
		case "":
			taxonomyFields["parent_id"] = nil
		default:
			parent, err := s.dao.GetTaxonomy(ctx, *parentID)
			if err != nil {
				return nil, err
			}
			if parent == nil || parent.Taxonomy != "category" || parent.Status != string(classification.StatusActive) {
				return nil, rescause.InvalidInput("category parent must be an active category")
			}
			descendant, err := s.categoryDescendsFrom(ctx, *parentID, id)
			if err != nil {
				return nil, err
			}
			if descendant {
				return nil, rescause.InvalidInput("category parent would create a cycle")
			}
			taxonomyFields["parent_id"] = *parentID
		}
	}
	if len(termFields) == 0 && len(taxonomyFields) == 0 {
		return current, nil
	}
	if err := s.dao.UpdateTaxonomyWithHook(
		ctx, current, termFields, taxonomyFields, tagLookupKey,
		s.taxonomyURLHook("resource taxonomy updated"),
	); err != nil {
		return nil, err
	}
	return s.dao.GetTaxonomy(ctx, id)
}

// DeleteTaxonomy is intentionally conservative. It refuses implicit subtree or
// assignment cascades; callers must first reparent, merge or unassign.
func (s *Service) DeleteTaxonomy(ctx context.Context, id string) error {
	current, err := s.dao.GetTaxonomy(ctx, id)
	if err != nil {
		return err
	}
	if current == nil || current.Status == string(classification.StatusReplaced) {
		return rescause.NotFound(id)
	}
	children, err := s.dao.TaxonomyChildCount(ctx, id)
	if err != nil {
		return err
	}
	if children > 0 {
		return rescause.InvalidState("taxonomy has children; reparent or merge them first")
	}
	assignments, err := s.dao.TaxonomyAssignmentCount(ctx, id)
	if err != nil {
		return err
	}
	if assignments > 0 {
		return rescause.InvalidState("taxonomy has assignments; unassign or merge it first")
	}
	return s.dao.DeleteTaxonomyWithHook(
		ctx, current, s.taxonomyURLHook("resource taxonomy deleted"),
	)
}

func (s *Service) MergeTaxonomy(ctx context.Context, sourceID, targetID string) error {
	if sourceID == targetID {
		return rescause.InvalidInput("cannot merge a taxonomy into itself")
	}
	source, err := s.dao.GetTaxonomy(ctx, sourceID)
	if err != nil {
		return err
	}
	target, err := s.dao.GetTaxonomy(ctx, targetID)
	if err != nil {
		return err
	}
	if source == nil || target == nil {
		return rescause.NotFound(sourceID)
	}
	if source.Status == string(classification.StatusReplaced) {
		return rescause.InvalidState("merge source is already replaced")
	}
	if source.Taxonomy != target.Taxonomy {
		return rescause.InvalidInput("category and tag cannot be merged across kinds")
	}
	if target.Status != string(classification.StatusActive) {
		return rescause.InvalidState("merge target must be active")
	}
	if source.Taxonomy == "category" {
		descendant, err := s.categoryDescendsFrom(ctx, targetID, sourceID)
		if err != nil {
			return err
		}
		if descendant {
			return rescause.InvalidInput("category cannot merge into its own subtree")
		}
	}
	return s.dao.MergeTaxonomyWithHook(
		ctx, source, target, s.taxonomyMergeURLHook(source, target),
	)
}

func (s *Service) AssignTaxonomies(ctx context.Context, owner, resourceID string, taxonomyIDs []string) error {
	resource, err := s.dao.GetByID(ctx, resourceID)
	if err != nil {
		return err
	}
	if resource == nil || resource.OwnerID != owner {
		return rescause.NotFound(resourceID)
	}
	ids := uniqueTaxonomyIDs(taxonomyIDs)
	values, err := s.dao.TaxonomiesByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(values) != len(ids) {
		return rescause.InvalidInput("one or more taxonomy ids do not exist")
	}
	categoryIDs := make([]string, 0, len(values))
	tags := make([]string, 0, len(values))
	for _, value := range values {
		switch value.Taxonomy {
		case "category":
			categoryIDs = append(categoryIDs, value.ID)
		case "tag":
			tags = append(tags, value.Name)
		default:
			return rescause.InvalidInput("unsupported taxonomy identity")
		}
	}
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return err
	}
	preparation := catalog.Classify(classification.ClassifyRequest{
		PolicyKey: resourceItemPolicyKey, CategoryIDs: categoryIDs, Tags: tags,
	})
	request := preparation.FactRequest()
	matches, freshnessToken, err := s.dao.ClassificationTagMatches(ctx, request.TagLookups)
	if err != nil {
		return err
	}
	result := preparation.Complete(classification.ClassifyFacts{
		CatalogRevision: request.CatalogRevision,
		RequestToken:    request.RequestToken,
		FreshnessToken:  freshnessToken,
		TagMatches:      matches,
	})
	if result.Outcome != classification.OutcomeAccepted {
		return resourceClassificationError(result.Diagnostics)
	}
	if len(result.TagCreations) != 0 || len(result.TagProposals) != 0 {
		return rescause.InvalidState("selected tags must resolve before assignment")
	}
	return s.dao.SetResourceClassification(ctx, resourceID, result.Assignments)
}

func (s *Service) classificationCatalog(ctx context.Context) (*classification.Catalog, error) {
	snapshot, err := s.dao.ClassificationSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	result := classification.Compile(snapshot)
	if result.Outcome != classification.OutcomeAccepted || result.Catalog == nil {
		return nil, resourceClassificationError(result.Diagnostics)
	}
	return result.Catalog, nil
}

func (s *Service) classificationTagLookup(ctx context.Context, name string) (classification.TagLookupRequest, error) {
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return classification.TagLookupRequest{}, err
	}
	request := catalog.Classify(classification.ClassifyRequest{
		PolicyKey: resourceItemPolicyKey, Tags: []string{name},
	}).FactRequest()
	if len(request.TagLookups) != 1 {
		return classification.TagLookupRequest{}, rescause.InvalidInput("tag name is empty")
	}
	return request.TagLookups[0], nil
}

func (s *Service) categoryDescendsFrom(ctx context.Context, categoryID, ancestorID string) (bool, error) {
	seen := map[string]bool{}
	current := categoryID
	for current != "" {
		if current == ancestorID || seen[current] {
			return true, nil
		}
		seen[current] = true
		value, err := s.dao.GetTaxonomy(ctx, current)
		if err != nil {
			return false, err
		}
		if value == nil || value.Taxonomy != "category" {
			return false, nil
		}
		current = value.ParentID
	}
	return false, nil
}

func uniqueTaxonomyIDs(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, raw := range values {
		if value := strings.TrimSpace(raw); value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func resourceClassificationError(diagnostics []classification.Diagnostic) error {
	if len(diagnostics) == 0 {
		return rescause.InvalidState("classification rejected without diagnostics")
	}
	return rescause.InvalidInput("classification rejected: " + string(diagnostics[0].Code))
}
