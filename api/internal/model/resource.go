// Package model holds the resource-catalog row structs. orm tags pin
// the column mapping for GoFrame gdb scan/insert.
package model

import "github.com/gogf/gf/v2/os/gtime"

// Status is the publish lifecycle of a resource.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

// Resource is one downloadable catalog item (= one or more assets + metadata).
type Resource struct {
	ID              string      `json:"id" orm:"id"`
	OwnerID         string      `json:"ownerId" orm:"owner_id"`
	Title           string      `json:"title" orm:"title"`
	Slug            string      `json:"slug" orm:"slug"`
	Summary         string      `json:"summary" orm:"summary"`
	Description     string      `json:"description" orm:"description"`
	Type            string      `json:"type" orm:"type"`
	CoverAssetID    string      `json:"coverAssetId" orm:"cover_asset_id"`
	CoverURL        string      `json:"coverUrl" orm:"cover_url"`
	DeliveryKind    string      `json:"deliveryKind" orm:"delivery_kind"`
	DeliveryPayload string      `json:"deliveryPayload" orm:"delivery_payload"`
	Status          Status      `json:"status" orm:"status"`
	PublishedAt     *gtime.Time `json:"publishedAt" orm:"published_at"`
	ViewCount       int64       `json:"viewCount" orm:"view_count"`
	DownloadCount   int64       `json:"downloadCount" orm:"download_count"`
	Tags            []string    `json:"tags" orm:"tags"`
	CreatedAt       *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt       *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

// ResourceAsset is one file within a resource.
type ResourceAsset struct {
	ID         string `json:"id" orm:"id"`
	ResourceID string `json:"resourceId" orm:"resource_id"`
	AssetID    string `json:"assetId" orm:"asset_id"`
	Label      string `json:"label" orm:"label"`
	CdnURL     string `json:"cdnUrl" orm:"cdn_url"`
	Size       int64  `json:"size" orm:"size"`
	Mime       string `json:"mime" orm:"mime"`
	Filename   string `json:"filename" orm:"filename"`
	Sort       int    `json:"sort" orm:"sort"`
}

type HomeSettings struct {
	ID      int    `json:"id" orm:"id"`
	Payload string `json:"payload" orm:"payload"`
}

type SiteSettings struct {
	Key     string `json:"key" orm:"key"`
	Payload string `json:"payload" orm:"payload"`
}
