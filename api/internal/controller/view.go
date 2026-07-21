// Package controller holds the resource-site HTTP handlers.
package controller

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/model"
	"platform/products/resource/api/internal/reserr"
)

// subject extracts the authenticated subject (JWT group), or a forbidden error.
func subject(ctx context.Context) (string, error) {
	p, ok := foundationauth.FromContext(ctx)
	if !ok {
		return "", reserr.Forbidden()
	}
	return p.Subject, nil
}

// isAdmin is site-scoped: an IdP global admin does not automatically operate
// every Resource deployment.
func isAdmin(ctx context.Context) bool {
	p, ok := foundationauth.FromContext(ctx)
	operators, err := g.Cfg().Get(ctx, "resource.operatorSubs")
	return ok && p != nil && err == nil && slices.Contains(operators.Strings(), p.Subject)
}

// requireAdmin returns a 403 unless the caller is a resource superadmin.
func requireAdmin(ctx context.Context) error {
	if !isAdmin(ctx) {
		return reserr.Forbidden()
	}
	return nil
}

// bearerOf returns the raw bearer token on the request (to forward to the asset
// service; operator == asset owner).
func bearerOf(ctx context.Context) string {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	return stripBearer(r.Request.Header.Get("Authorization"))
}

// optionalSubject verifies the bearer token if present, returning the subject or
// "" (anonymous). Used by the public browse/download endpoints (optional login).
func optionalSubject(ctx context.Context, v *foundationauth.Verifier) string {
	raw := bearerOf(ctx)
	if raw == "" || v == nil {
		return ""
	}
	p, err := v.Verify(ctx, raw)
	if err != nil {
		return ""
	}
	return p.Subject
}

func stripBearer(h string) string {
	const p = "bearer "
	if len(h) < len(p) || !strings.EqualFold(h[:len(p)], p) {
		return ""
	}
	return strings.TrimSpace(h[len(p):])
}

func resourceView(r *model.Resource) *v1.ResourceView {
	v := &v1.ResourceView{
		ID: r.ID, OwnerID: r.OwnerID, Title: r.Title, Slug: r.Slug,
		Summary: r.Summary, Description: r.Description, Type: r.Type,
		CoverAssetID: r.CoverAssetID, CoverURL: r.CoverURL,
		DeliveryKind:    deliveryKindView(r.DeliveryKind),
		DeliveryPayload: deliveryPayloadView(r.DeliveryPayload),
		Status:          string(r.Status),
		ViewCount:       r.ViewCount,
		DownloadCount:   r.DownloadCount, IssueCount: r.IssueCount, Tags: r.Tags,
	}
	if r.PublishedAt != nil {
		v.PublishedAt = r.PublishedAt.Time.UTC().Format(time.RFC3339)
	}
	if v.Tags == nil {
		v.Tags = []string{}
	}
	if r.CreatedAt != nil {
		v.CreatedAt = r.CreatedAt.Time.UTC().Format(time.RFC3339)
	}
	if r.UpdatedAt != nil {
		v.UpdatedAt = r.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}
	return v
}

func deliveryKindView(kind string) string {
	if strings.TrimSpace(kind) == "" {
		return "asset_file"
	}
	return kind
}

func deliveryPayloadView(raw string) *v1.DeliveryPayloadView {
	var payload v1.DeliveryPayloadView
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &payload)
	}
	if payload.Items == nil {
		payload.Items = []*v1.DeliveryItemView{}
	}
	for i, item := range payload.Items {
		if item == nil {
			continue
		}
		if item.Kind == "" {
			item.Kind = "netdisk"
		}
		if item.Title == "" {
			item.Title = "网盘交付"
		}
		if item.Sort == 0 {
			item.Sort = i
		}
	}
	return &payload
}

func assetView(a *model.ResourceAsset) *v1.ResourceAssetView {
	return &v1.ResourceAssetView{
		AssetID: a.AssetID, Label: a.Label, Mime: a.Mime,
		Filename: a.Filename, Size: a.Size, Sort: a.Sort,
	}
}

func taxonomyView(t *model.Taxonomy) *v1.TaxonomyView {
	return &v1.TaxonomyView{
		ID: t.ID, Taxonomy: t.Taxonomy, Name: t.Name, Slug: t.Slug,
		Description: t.Description, ParentID: t.ParentID, Count: t.Count,
	}
}

func taxonomyViews(ts []*model.Taxonomy) []*v1.TaxonomyView {
	out := make([]*v1.TaxonomyView, 0, len(ts))
	for _, t := range ts {
		out = append(out, taxonomyView(t))
	}
	return out
}

func seoView(s *model.SEO) *v1.SEOView {
	if s == nil {
		return nil
	}
	return &v1.SEOView{
		MetaTitle: s.MetaTitle, MetaDesc: s.MetaDesc, OgTitle: s.OgTitle,
		OgImage: s.OgImage, CanonicalURL: s.CanonicalURL, Robots: s.Robots,
	}
}
