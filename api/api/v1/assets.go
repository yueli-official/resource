package v1

import "github.com/gogf/gf/v2/frame/g"

// AddAssetReq opens an upload for a file within a resource (operator).
type AddAssetReq struct {
	g.Meta    `path:"/api/v1/resources/{id}/assets" method:"post" tags:"resource" summary:"Begin a file upload for a resource"`
	ID        string `json:"id" in:"path" v:"required"`
	Filename  string `json:"filename" v:"required"`
	Size      int64  `json:"size" v:"required|min:1"`
	Multipart bool   `json:"multipart"`
}

type AddAssetRes struct {
	g.Meta        `status:"201"`
	UploadURL     string            `json:"uploadUrl"`
	UploadToken   string            `json:"uploadToken"` // = assetRef for finalize
	Method        string            `json:"method"`
	UploadHeaders map[string]string `json:"uploadHeaders,omitempty"`
	UploadID      string            `json:"uploadId,omitempty"`
	PartSize      int64             `json:"partSize,omitempty"`
	PartCount     int               `json:"partCount,omitempty"`
}

type MultipartPartURLReq struct {
	g.Meta      `path:"/api/v1/resources/{id}/assets/multipart/part-url" method:"post" tags:"resource" summary:"Create a multipart part upload URL"`
	ID          string `json:"id" in:"path" v:"required"`
	UploadToken string `json:"uploadToken" v:"required"`
	PartNumber  int    `json:"partNumber" v:"required|min:1"`
}

type MultipartPartURLRes struct {
	UploadURL     string            `json:"uploadUrl"`
	UploadHeaders map[string]string `json:"uploadHeaders,omitempty"`
}

type MultipartCompleteReq struct {
	g.Meta      `path:"/api/v1/resources/{id}/assets/multipart/complete" method:"post" tags:"resource" summary:"Complete a multipart upload"`
	ID          string               `json:"id" in:"path" v:"required"`
	UploadToken string               `json:"uploadToken" v:"required"`
	Parts       []MultipartPartInput `json:"parts" v:"required"`
}

type MultipartPartInput struct {
	PartNumber int    `json:"partNumber"`
	ETag       string `json:"etag"`
}

type MultipartCompleteRes struct {
	Completed bool `json:"completed"`
}

type MultipartAbortReq struct {
	g.Meta      `path:"/api/v1/resources/{id}/assets/multipart/abort" method:"post" tags:"resource" summary:"Abort a multipart upload"`
	ID          string `json:"id" in:"path" v:"required"`
	UploadToken string `json:"uploadToken" v:"required"`
}

type MultipartAbortRes struct {
	g.Meta `status:"204"`
}

// FinalizeAssetReq finalizes an uploaded blob and links it to the resource.
type FinalizeAssetReq struct {
	g.Meta      `path:"/api/v1/resources/{id}/assets/finalize" method:"post" tags:"resource" summary:"Finalize + link an uploaded file"`
	ID          string `json:"id" in:"path" v:"required"`
	UploadToken string `json:"uploadToken" v:"required"`
	Label       string `json:"label"`
}

type FinalizeAssetRes struct {
	g.Meta `status:"201"`
	Asset  *ResourceAssetView `json:"asset"`
}

// RemoveAssetReq removes one file from a resource (operator).
type RemoveAssetReq struct {
	g.Meta  `path:"/api/v1/resources/{id}/assets/{assetId}" method:"delete" tags:"resource" summary:"Remove a file from a resource"`
	ID      string `json:"id" in:"path" v:"required"`
	AssetID string `json:"assetId" in:"path" v:"required"`
}

type RemoveAssetRes struct {
	g.Meta `status:"204"`
}

// PingReq is the authenticated liveness probe — proof the JWT chain is wired.
type PingReq struct {
	g.Meta `path:"/api/v1/resources/_ping" method:"get" tags:"resource" summary:"Authenticated ping; echoes the caller subject"`
}

type PingRes struct {
	Subject string `json:"subject"`
}
