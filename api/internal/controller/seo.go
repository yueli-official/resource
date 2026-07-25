package controller

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/resourceauthz"
)

// SEO handles the operator (JWT) resource SEO endpoint.
type SEO struct{ svc *catalog.Service }

func NewSEO(svc *catalog.Service) *SEO { return &SEO{svc: svc} }

// PutSEO upserts the resource's SEO metadata (owner-gated; only provided fields
// are applied).
func (c *SEO) PutSEO(ctx context.Context, req *v1.PutSEOReq) (*v1.PutSEORes, error) {
	resource, err := authorizationResource(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	fields := g.Map{}
	if req.MetaTitle != nil {
		fields["meta_title"] = *req.MetaTitle
	}
	if req.MetaDesc != nil {
		fields["meta_desc"] = *req.MetaDesc
	}
	if req.OgTitle != nil {
		fields["og_title"] = *req.OgTitle
	}
	if req.OgImage != nil {
		fields["og_image"] = *req.OgImage
	}
	if req.CanonicalURL != nil {
		fields["canonical_url"] = *req.CanonicalURL
	}
	if req.Robots != nil {
		fields["robots"] = *req.Robots
	}
	if err := requireCapability(
		ctx, resourceauthz.CapabilityItemUpdate, resourceauthz.ResourceScopeID(req.ID), resource,
	); err != nil {
		return nil, err
	}
	s, err := c.svc.PutSEO(ctx, resourceauthz.ResourceOwner(resource), req.ID, fields)
	if err != nil {
		return nil, err
	}
	return &v1.PutSEORes{SEO: seoView(s)}, nil
}
