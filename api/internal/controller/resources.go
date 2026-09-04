package controller

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/foundation/go/authorization"
	"github.com/yueli-official/foundation/go/identifier"

	v1 "github.com/yueli-official/resource/api/api/v1"
	"github.com/yueli-official/resource/api/internal/catalog"
	"github.com/yueli-official/resource/api/internal/dao"
	"github.com/yueli-official/resource/api/internal/reserr"
	"github.com/yueli-official/resource/api/internal/resourceauthz"
)

// Resources handles the operator (JWT) resource-management endpoints.
type Resources struct{ svc *catalog.Service }

func NewResources(svc *catalog.Service) *Resources { return &Resources{svc: svc} }

func (c *Resources) GetResourceMe(
	ctx context.Context,
	_ *v1.GetResourceMeReq,
) (*v1.GetResourceMeRes, error) {
	access, err := authorizationService(ctx).EffectiveAccess(ctx)
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	roles := make([]string, len(access.Grants))
	for index, grant := range access.Grants {
		roles[index] = string(grant.Role)
	}
	capabilities := make([]string, len(access.Capabilities))
	for index, capability := range access.Capabilities {
		capabilities[index] = string(capability)
	}
	return &v1.GetResourceMeRes{
		IsAdministrator: isAdmin(ctx), Roles: roles, Capabilities: capabilities,
	}, nil
}

func (c *Resources) ListMine(ctx context.Context, req *v1.ListMineReq) (*v1.ListMineRes, error) {
	owner, err := authorizationService(ctx).ManageOwner(ctx)
	if err != nil {
		return nil, mapAuthorizationError(err)
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
	owner, err := authorizationService(ctx).ManageOwner(ctx)
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	capability := resourceauthz.CapabilityItemUpdate
	switch req.Action {
	case "publish":
		capability = resourceauthz.CapabilityItemPublish
	case "archive":
		capability = resourceauthz.CapabilityItemArchive
	}
	for _, id := range req.IDs {
		resource, resourceErr := authorizationResource(ctx, id)
		if resourceErr != nil {
			return nil, resourceErr
		}
		if resourceErr = requireCapability(
			ctx, capability, resourceauthz.ResourceScopeID(id), resource,
		); resourceErr != nil {
			return nil, resourceErr
		}
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
	if err := requireCapability(
		ctx, resourceauthz.CapabilityItemCreate, resourceauthz.RootScopeID,
		authorization.ResourceFacts{},
	); err != nil {
		return nil, err
	}
	r, err := c.svc.Create(ctx, owner, catalog.CreateInput{
		Title: req.Title, Summary: req.Summary, Description: req.Description,
		Type: req.Type, CoverAssetID: req.CoverAssetID, Tags: req.Tags,
	})
	if err != nil {
		return nil, err
	}
	if err := authorizationService(ctx).EnsureResourceScope(ctx, r.ID); err != nil {
		return nil, reserr.AuthorizationUnavailable()
	}
	g.RequestFromCtx(ctx).Response.Header().Set("Location", "/api/v1/resources/"+url.PathEscape(r.ID))
	return &v1.CreateResourceRes{Resource: resourceView(r)}, nil
}

func (c *Resources) PatchResource(ctx context.Context, req *v1.PatchResourceReq) (*v1.PatchResourceRes, error) {
	if _, err := subject(ctx); err != nil {
		return nil, err
	}
	resource, err := authorizationResource(ctx, req.ID)
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
		raw, err := json.Marshal(normalizeDeliveryPayloadIDs(req.DeliveryPayload))
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
	if req.DownloadCount != nil {
		fields["download_count"] = maxInt64(0, *req.DownloadCount)
	}
	capability := resourceauthz.CapabilityItemUpdate
	if req.Status != nil {
		switch *req.Status {
		case "published":
			capability = resourceauthz.CapabilityItemPublish
		case "archived":
			capability = resourceauthz.CapabilityItemArchive
		}
	}
	if err := requireCapability(
		ctx, capability, resourceauthz.ResourceScopeID(req.ID), resource,
	); err != nil {
		return nil, err
	}
	r, err := c.svc.Patch(ctx, resourceauthz.ResourceOwner(resource), req.ID, fields)
	if err != nil {
		return nil, err
	}
	return &v1.PatchResourceRes{Resource: resourceView(r)}, nil
}

func normalizeDeliveryPayloadIDs(payload *v1.DeliveryPayloadView) *v1.DeliveryPayloadView {
	if payload == nil {
		return nil
	}
	out := &v1.DeliveryPayloadView{Items: make([]*v1.DeliveryItemView, 0, len(payload.Items))}
	for _, item := range payload.Items {
		if item == nil {
			continue
		}
		copy := *item
		value, err := identifier.Parse(strings.TrimSpace(copy.ID))
		if err != nil || value.Version() != 7 {
			copy.ID = identifier.MustNew().String()
		} else {
			copy.ID = value.String()
		}
		if item.Netdisk != nil {
			netdisk := *item.Netdisk
			copy.Netdisk = &netdisk
		}
		out.Items = append(out.Items, &copy)
	}
	return out
}

func maxInt64(min, value int64) int64 {
	if value < min {
		return min
	}
	return value
}

func (c *Resources) DeleteResource(ctx context.Context, req *v1.DeleteResourceReq) (*v1.DeleteResourceRes, error) {
	if _, err := subject(ctx); err != nil {
		return nil, err
	}
	resource, err := authorizationResource(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if err := requireCapability(
		ctx, resourceauthz.CapabilityItemDelete, resourceauthz.ResourceScopeID(req.ID), resource,
	); err != nil {
		return nil, err
	}
	if err := c.svc.Delete(
		ctx, resourceauthz.ResourceOwner(resource), bearerOf(ctx), req.ID,
	); err != nil {
		return nil, err
	}
	return &v1.DeleteResourceRes{}, nil
}
