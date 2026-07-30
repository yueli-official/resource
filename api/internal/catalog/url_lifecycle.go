package catalog

import (
	"context"
	"database/sql"

	"github.com/yueli-official/resource/api/internal/dao"
	"github.com/yueli-official/resource/api/internal/model"
	"github.com/yueli-official/resource/api/internal/resourceurls"
)

func (s *Service) SetURLLifecycle(lifecycle *resourceurls.Lifecycle) {
	s.urls = lifecycle
}

func (s *Service) resourceURLHook(ids []string) dao.TransactionHook {
	if s.urls == nil {
		return nil
	}
	return func(ctx context.Context, tx *sql.Tx) error {
		return s.urls.ReconcileResources(ctx, tx, ids)
	}
}

func (s *Service) resourceDeleteURLHook(id string) dao.TransactionHook {
	if s.urls == nil {
		return nil
	}
	return func(ctx context.Context, tx *sql.Tx) error {
		return s.urls.DeleteResource(ctx, tx, id)
	}
}

func (s *Service) resourceSearchHook(ids []string) dao.TransactionHook {
	if s.search == nil {
		return nil
	}
	return func(ctx context.Context, tx *sql.Tx) error {
		for _, id := range ids {
			if err := s.search.Hook(id)(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	}
}

func (s *Service) resourceSearchDeleteHook(id string, revision uint64) dao.TransactionHook {
	if s.search == nil {
		return nil
	}
	return s.search.DeleteHook(id, revision)
}

func (s *Service) taxonomyURLHook(reason string) dao.TransactionHook {
	if s.urls == nil {
		return nil
	}
	return func(ctx context.Context, tx *sql.Tx) error {
		return s.urls.ReconcileTaxonomies(ctx, tx, reason)
	}
}

func (s *Service) taxonomyMergeURLHook(source, target *model.Taxonomy) dao.TransactionHook {
	if s.urls == nil {
		return nil
	}
	return func(ctx context.Context, tx *sql.Tx) error {
		return s.urls.MergeTaxonomy(ctx, tx, taxonomyURLState(source), taxonomyURLState(target))
	}
}

func taxonomyURLState(value *model.Taxonomy) resourceurls.TaxonomyState {
	kind := resourceurls.CategoryKind
	if value != nil && value.Taxonomy == "tag" {
		kind = resourceurls.TagKind
	}
	if value == nil {
		return resourceurls.TaxonomyState{Kind: kind}
	}
	return resourceurls.TaxonomyState{ID: value.ID, Kind: kind, Slug: value.Slug}
}
