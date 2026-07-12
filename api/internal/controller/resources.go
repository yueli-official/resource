package controller

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"

	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/dao"
)

// Resources handles the operator (JWT) resource-management endpoints.
type Resources struct{ svc *catalog.Service }

func NewResources(svc *catalog.Service) *Resources { return &Resources{svc: svc} }

func (c *Resources) ListMine(ctx context.Context, req *v1.ListMineReq) (*v1.ListMineRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	items, total, page, size, err := c.svc.ListMine(ctx, owner, dao.OwnerListFilter{
		Status:    req.Status,
		Q:         req.Q,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	counts, err := c.svc.OwnerResourceCounts(ctx, owner)
	if err != nil {
		return nil, err
	}
	views := make([]*v1.ResourceView, 0, len(items))
	for _, r := range items {
		views = append(views, resourceView(r))
	}
	return &v1.ListMineRes{
		Items: views, Total: total, Page: page, Size: size,
		Counts: v1.ResourceLifecycleCounts{
			All: counts.All, Published: counts.Published, Draft: counts.Draft,
			Archived: counts.Archived, Issues: counts.Issues,
		},
	}, nil
}

func (c *Resources) BatchMine(ctx context.Context, req *v1.BatchMineReq) (*v1.BatchMineRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	changed, failures, err := c.svc.BatchMine(ctx, owner, req.IDs, req.Action)
	if err != nil {
		return nil, err
	}
	views := make([]*v1.ResourceBatchFailure, 0, len(failures))
	for _, failure := range failures {
		views = append(views, &v1.ResourceBatchFailure{ID: failure.ID, Code: failure.Code, Message: failure.Message})
	}
	return &v1.BatchMineRes{Changed: changed, Failures: views}, nil
}

func (c *Resources) CreateResource(ctx context.Context, req *v1.CreateResourceReq) (*v1.CreateResourceRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	r, err := c.svc.Create(ctx, owner, catalog.CreateInput{
		Title: req.Title, Summary: req.Summary, Description: req.Description,
		Type: req.Type, CoverAssetID: req.CoverAssetID, Tags: req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateResourceRes{Resource: resourceView(r)}, nil
}

func (c *Resources) PatchResource(ctx context.Context, req *v1.PatchResourceReq) (*v1.PatchResourceRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	fields := g.Map{}
	if req.Slug != nil {
		fields["slug"] = *req.Slug
	}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Summary != nil {
		fields["summary"] = *req.Summary
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if req.CoverAssetID != nil {
		fields["cover_asset_id"] = *req.CoverAssetID
	}
	if req.CoverURL != nil {
		fields["cover_url"] = *req.CoverURL
	}
	if req.DeliveryKind != nil {
		fields["delivery_kind"] = *req.DeliveryKind
	}
	if req.DeliveryPayload != nil {
		raw, err := json.Marshal(req.DeliveryPayload)
		if err != nil {
			return nil, err
		}
		fields["delivery_payload"] = string(raw)
	}
	if req.Tags != nil {
		fields["tags"] = *req.Tags
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.PublishedAt != nil {
		if *req.PublishedAt == "" {
			fields["published_at"] = nil
		} else {
			fields["published_at"] = *req.PublishedAt
		}
	}
	if req.ViewCount != nil {
		fields["view_count"] = maxInt64(0, *req.ViewCount)
	}
	if req.DownloadCount != nil {
		fields["download_count"] = maxInt64(0, *req.DownloadCount)
	}
	r, err := c.svc.Patch(ctx, owner, req.ID, fields)
	if err != nil {
		return nil, err
	}
	return &v1.PatchResourceRes{Resource: resourceView(r)}, nil
}

func maxInt64(min, value int64) int64 {
	if value < min {
		return min
	}
	return value
}

func (c *Resources) DeleteResource(ctx context.Context, req *v1.DeleteResourceReq) (*v1.DeleteResourceRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.svc.Delete(ctx, owner, bearerOf(ctx), req.ID); err != nil {
		return nil, err
	}
	return &v1.DeleteResourceRes{Deleted: true}, nil
}
