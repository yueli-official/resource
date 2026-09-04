package catalog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/yueli-official/resource/api/internal/model"
	"github.com/yueli-official/resource/api/internal/rescause"
)

// GetSEO returns a resource's SEO metadata (nil when unset).
func (s *Service) GetSEO(ctx context.Context, resourceID string) (*model.SEO, error) {
	return s.dao.GetSEO(ctx, resourceID)
}

// PutSEO upserts a resource's SEO metadata (owner-gated). fields is the provided
// subset built by the caller; an empty map just ensures the row exists.
func (s *Service) PutSEO(ctx context.Context, owner, resourceID string, fields g.Map) (*model.SEO, error) {
	r, err := s.dao.GetByID(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	if r == nil || r.OwnerID != owner {
		return nil, rescause.NotFound(resourceID)
	}
	if err := s.dao.UpsertSEO(ctx, resourceID, fields); err != nil {
		return nil, err
	}
	return s.dao.GetSEO(ctx, resourceID)
}
