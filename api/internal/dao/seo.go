package dao

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"platform/products/resource/api/internal/model"
)

const tSEO = "resource_seo"

// UpsertSEO ensures the resource's SEO row exists, then applies the given fields.
func (p *PG) UpsertSEO(ctx context.Context, resourceID string, fields g.Map) error {
	if _, err := p.db.Exec(ctx, "INSERT INTO resource_seo (resource_id) VALUES (?) ON CONFLICT (resource_id) DO NOTHING", resourceID); err != nil {
		return err
	}
	if len(fields) == 0 {
		return nil
	}
	_, err := p.db.Model(tSEO).Ctx(ctx).Where("resource_id", resourceID).Data(fields).Update()
	return err
}

// GetSEO returns the resource's SEO metadata, or (nil, nil) when absent.
func (p *PG) GetSEO(ctx context.Context, resourceID string) (*model.SEO, error) {
	var s *model.SEO
	if err := p.db.Model(tSEO).Ctx(ctx).Where("resource_id", resourceID).Limit(1).Scan(&s); err != nil {
		return nil, err
	}
	return s, nil
}
