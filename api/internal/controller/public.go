package controller

import (
	"context"
	"strings"
	"time"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/traffic"
	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/reserr"
	"platform/products/resource/api/internal/resourceauthz"
	"platform/products/resource/api/internal/resourcediscovery"
)

// PublicResources handles the public browse/download endpoints (optional login).
// It verifies a bearer token itself when present (not behind Foundation auth middleware).
type PublicResources struct {
	svc       *catalog.Service
	verifier  *foundationauth.Verifier
	discovery *resourcediscovery.Manager
}

func (c *PublicResources) RecordView(ctx context.Context, req *v1.RecordViewReq) (*v1.RecordViewRes, error) {
	occurredAt, err := time.Parse(time.RFC3339Nano, req.OccurredAt)
	if err != nil {
		return nil, reserr.InvalidInput("occurredAt must be an RFC3339 timestamp")
	}
	ip, ua := clientMeta(ctx)
	subject := optionalSubject(ctx, c.verifier)
	seed := anonymousVisitorSeed(ip, ua)
	if subject != "" {
		seed = []byte("subject\x00" + subject)
	}
	result, err := c.svc.RecordView(ctx, req.ID, catalog.ViewInput{
		EventID: req.EventID, OccurredAt: occurredAt,
		Class: classifyVisit(ua), VisitorSeed: seed,
	})
	if err != nil {
		if traffic.IsKind(err, traffic.ErrorInvalidInput) || traffic.IsKind(err, traffic.ErrorConflict) {
			return nil, reserr.InvalidInput(err.Error())
		}
		return nil, err
	}
	return &v1.RecordViewRes{
		Ok: true, Counted: result.Counted, Replay: result.Replay,
		ViewCount: result.ResourceTotals.Views,
	}, nil
}

func anonymousVisitorSeed(ip, userAgent string) []byte {
	ip = strings.TrimSpace(ip)
	userAgent = strings.TrimSpace(userAgent)
	if ip == "" && userAgent == "" {
		return nil
	}
	return []byte("network\x00" + ip + "\x00" + userAgent)
}

func classifyVisit(userAgent string) traffic.VisitClass {
	value := strings.ToLower(userAgent)
	for _, marker := range []string{
		"bot", "crawler", "spider", "slurp", "headless", "monitoring",
		"facebookexternalhit", "twitterbot", "preview",
	} {
		if strings.Contains(value, marker) {
			return traffic.VisitBot
		}
	}
	if strings.TrimSpace(value) == "" {
		return traffic.VisitUnknown
	}
	return traffic.VisitHuman
}

func NewPublicResources(svc *catalog.Service, v *foundationauth.Verifier, modules ...*resourcediscovery.Manager) *PublicResources {
	controller := &PublicResources{svc: svc, verifier: v}
	if len(modules) > 0 {
		controller.discovery = modules[0]
	}
	return controller
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
		if viewer == "" {
			return nil, err
		}
		manageContext := foundationauth.NewContext(ctx, &foundationauth.Principal{Subject: viewer})
		resource, resourceErr := authorizationResource(manageContext, req.ID)
		if resourceErr != nil {
			return nil, err
		}
		if resourceErr = requireCapability(
			manageContext, resourceauthz.CapabilityItemRead,
			resourceauthz.ResourceScopeID(req.ID), resource,
		); resourceErr != nil {
			return nil, err
		}
		r, err = c.svc.GetManage(ctx, req.ID)
		if err != nil {
			return nil, err
		}
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
	response := &v1.GetResourceRes{
		Resource: resourceView(r), Assets: views,
		Taxonomies: taxonomyViews(taxes), SEO: seoView(seo),
	}
	if c.discovery != nil && r.Status == "published" {
		projection, err := c.discovery.ProjectResource(r, seo)
		if err != nil {
			return nil, err
		}
		response.Discovery = &projection
	}
	return response, nil
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
