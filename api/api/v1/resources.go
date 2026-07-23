// Package v1 holds the resource-site typed-handler request/response contracts
// (g.Meta drives GoFrame's auto OpenAPI).
package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/foundation/go/discovery"
)

// ResourceView is the outward projection of a resource (fully free — no access
// gate / price / points).
type ResourceView struct {
	ID              string               `json:"id"`
	OwnerID         string               `json:"ownerId"`
	Title           string               `json:"title"`
	Slug            string               `json:"slug"`
	Summary         string               `json:"summary"`
	Description     string               `json:"description"`
	Type            string               `json:"type"`
	CoverAssetID    string               `json:"coverAssetId,omitempty"`
	CoverURL        string               `json:"coverUrl,omitempty"`
	DeliveryKind    string               `json:"deliveryKind"`
	DeliveryPayload *DeliveryPayloadView `json:"deliveryPayload,omitempty"`
	Status          string               `json:"status"`
	PublishedAt     string               `json:"publishedAt,omitempty"`
	ViewCount       int64                `json:"viewCount"`
	DownloadCount   int64                `json:"downloadCount"`
	IssueCount      int                  `json:"issueCount"`
	Tags            []string             `json:"tags"`
	CreatedAt       string               `json:"createdAt"`
	UpdatedAt       string               `json:"updatedAt"`
}

// ResourceAssetView is one file within a resource (delivery via /download).
type ResourceAssetView struct {
	AssetID  string `json:"assetId"`
	Label    string `json:"label"`
	Mime     string `json:"mime"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Sort     int    `json:"sort"`
}

type NetdiskDeliveryView struct {
	Provider    string `json:"provider"`
	URL         string `json:"url"`
	AccessCode  string `json:"accessCode"`
	ExtractCode string `json:"extractCode"`
	Note        string `json:"note"`
}

type DeliveryItemView struct {
	ID       string               `json:"id"`
	Kind     string               `json:"kind"`
	Title    string               `json:"title"`
	AssetID  string               `json:"assetId,omitempty"`
	Netdisk  *NetdiskDeliveryView `json:"netdisk,omitempty"`
	Sort     int                  `json:"sort"`
	Enabled  bool                 `json:"enabled"`
	Required bool                 `json:"required"`
}

type DeliveryPayloadView struct {
	Items []*DeliveryItemView `json:"items"`
}

// ── browse (public) ──────────────────────────────────────────────────────────

type ListResourcesReq struct {
	g.Meta   `path:"/api/v1/resources" method:"get" tags:"resource" summary:"Browse / search published resources"`
	Type     string   `json:"type"`
	Tags     []string `json:"tags"`
	Taxonomy string   `json:"taxonomy"` // optional category/tag term slug filter
	Q        string   `json:"q"`        // optional site-search query (title/summary/description)
	Page     int      `json:"page"`
	Size     int      `json:"size"`
}

type ListResourcesRes struct {
	Items []*ResourceView `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

type GetResourceReq struct {
	g.Meta `path:"/api/v1/resources/{id}" method:"get" tags:"resource" summary:"Get a resource (with its files, taxonomies, SEO)"`
	ID     string `json:"id" in:"path" v:"required"`
}

type GetResourceRes struct {
	Resource   *ResourceView             `json:"resource"`
	Assets     []*ResourceAssetView      `json:"assets"`
	Taxonomies []*TaxonomyView           `json:"taxonomies"`
	SEO        *SEOView                  `json:"seo,omitempty"`
	Discovery  *discovery.PageProjection `json:"discovery,omitempty"`
}

type RecordViewReq struct {
	g.Meta     `path:"/api/v1/resources/{id}/view" method:"post" tags:"resource" summary:"Record an idempotent resource view"`
	ID         string `json:"id" in:"path" v:"required"`
	EventID    string `json:"eventId" v:"required|length:16,200"`
	OccurredAt string `json:"occurredAt" v:"required"`
}

type RecordViewRes struct {
	Ok        bool  `json:"ok"`
	Counted   bool  `json:"counted"`
	Replay    bool  `json:"replay"`
	ViewCount int64 `json:"viewCount"`
}

type DownloadReq struct {
	g.Meta  `path:"/api/v1/resources/{id}/download/{assetId}" method:"get" tags:"resource" summary:"Resolve the public delivery URL"`
	ID      string `json:"id" in:"path" v:"required"`
	AssetID string `json:"assetId" in:"path" v:"required"`
}

type DownloadRes struct {
	DeliveryURL string `json:"deliveryUrl"`
}

// ── manage (operator JWT) ────────────────────────────────────────────────────

// ListMineReq lists the caller's own resources of any status (draft/published/
// archived) for the operator console — distinct from the public browse list,
// which only returns published resources.
type ListMineReq struct {
	g.Meta    `path:"/api/v1/resources/mine" method:"get" tags:"resource" summary:"List my resources (any status)"`
	Q         string `json:"q"`
	Status    string `json:"status"`
	SortBy    string `json:"sortBy"`
	SortOrder string `json:"sortOrder"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

type ListMineRes struct {
	Items  []*ResourceView         `json:"items"`
	Total  int                     `json:"total"`
	Page   int                     `json:"page"`
	Size   int                     `json:"size"`
	Counts ResourceLifecycleCounts `json:"counts"`
}

type ResourceLifecycleCounts struct {
	All       int `json:"all"`
	Published int `json:"published"`
	Draft     int `json:"draft"`
	Archived  int `json:"archived"`
	Issues    int `json:"issues"`
}

type BatchMineReq struct {
	g.Meta `path:"/api/v1/resources/mine/batch" method:"post" tags:"resource" summary:"Batch update my resource lifecycle"`
	IDs    []string `json:"ids" v:"required"`
	Action string   `json:"action" v:"required|in:publish,draft,archive"`
}

type BatchMineRes struct {
	Changed  int                     `json:"changed"`
	Failures []*ResourceBatchFailure `json:"failures"`
}

type ResourceBatchFailure struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CreateResourceReq struct {
	g.Meta       `path:"/api/v1/resources" method:"post" tags:"resource" summary:"Create a draft resource"`
	Title        string   `json:"title" v:"required"`
	Summary      string   `json:"summary"`
	Description  string   `json:"description"`
	Type         string   `json:"type" v:"required"`
	CoverAssetID string   `json:"coverAssetId"`
	Tags         []string `json:"tags"`
}

type CreateResourceRes struct {
	Resource *ResourceView `json:"resource"`
}

type PatchResourceReq struct {
	g.Meta          `path:"/api/v1/resources/{id}" method:"patch" tags:"resource" summary:"Update a resource"`
	ID              string               `json:"id" in:"path" v:"required"`
	Slug            *string              `json:"slug"`
	Title           *string              `json:"title"`
	Summary         *string              `json:"summary"`
	Description     *string              `json:"description"`
	Type            *string              `json:"type"`
	CoverAssetID    *string              `json:"coverAssetId"`
	CoverURL        *string              `json:"coverUrl"`
	DeliveryKind    *string              `json:"deliveryKind"`
	DeliveryPayload *DeliveryPayloadView `json:"deliveryPayload"`
	Tags            *[]string            `json:"tags"`
	Status          *string              `json:"status"`
	PublishedAt     *string              `json:"publishedAt"`
	DownloadCount   *int64               `json:"downloadCount"`
}

type PatchResourceRes struct {
	Resource *ResourceView `json:"resource"`
}

type DeleteResourceReq struct {
	g.Meta `path:"/api/v1/resources/{id}" method:"delete" tags:"resource" summary:"Delete a resource"`
	ID     string `json:"id" in:"path" v:"required"`
}

type DeleteResourceRes struct {
	Deleted bool `json:"deleted"`
}
