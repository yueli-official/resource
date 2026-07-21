package controller

import (
	"context"
	"strings"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/dao"
)

// PublicResources handles the public browse/download endpoints (optional login).
// It verifies a bearer token itself when present (not behind Foundation auth middleware).
type PublicResources struct {
	svc      *catalog.Service
	verifier *foundationauth.Verifier
}

func NewPublicResources(svc *catalog.Service, v *foundationauth.Verifier) *PublicResources {
	return &PublicResources{svc: svc, verifier: v}
}

func (c *PublicResources) ListResources(ctx context.Context, req *v1.ListResourcesReq) (*v1.ListResourcesRes, error) {
	items, total, err := c.svc.List(ctx, dao.ListFilter{
		Type: req.Type, Tags: req.Tags,
		TaxonomySlug: strings.TrimSpace(req.Taxonomy), Q: strings.TrimSpace(req.Q),
	}, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	views := make([]*v1.ResourceView, 0, len(items))
	for _, r := range items {
		views = append(views, resourceView(r))
	}
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return &v1.ListResourcesRes{Items: views, Total: total, Page: page, Size: size}, nil
}

func (c *PublicResources) GetResource(ctx context.Context, req *v1.GetResourceReq) (*v1.GetResourceRes, error) {
	viewer := optionalSubject(ctx, c.verifier)
	r, err := c.svc.Get(ctx, viewer, req.ID)
	if err != nil {
		return nil, err
	}
	assets, err := c.svc.Assets(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	views := make([]*v1.ResourceAssetView, 0, len(assets))
	for _, a := range assets {
		views = append(views, assetView(a))
	}
	taxes, err := c.svc.GetResourceTaxonomies(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	seo, err := c.svc.GetSEO(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	return &v1.GetResourceRes{
		Resource: resourceView(r), Assets: views,
		Taxonomies: taxonomyViews(taxes), SEO: seoView(seo),
	}, nil
}

// ListTaxonomies is the public category/tag browse (optional kind filter); each
// carries its published-resource count.
func (c *PublicResources) ListTaxonomies(ctx context.Context, req *v1.ListTaxonomiesReq) (*v1.ListTaxonomiesRes, error) {
	items, total, page, size, err := c.svc.ListTaxonomiesPage(ctx, req.Taxonomy, req.Q, req.Sort, req.Direction, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return &v1.ListTaxonomiesRes{Items: taxonomyViews(items), Total: total, Page: page, Size: size}, nil
}

func (c *PublicResources) Download(ctx context.Context, req *v1.DownloadReq) (*v1.DownloadRes, error) {
	viewer := optionalSubject(ctx, c.verifier)
	url, err := c.svc.Download(ctx, viewer, req.ID, req.AssetID)
	if err != nil {
		return nil, err
	}
	return &v1.DownloadRes{DeliveryURL: url}, nil
}
