package v1

import "github.com/gogf/gf/v2/frame/g"

// AddCoverReq opens an upload for a resource's cover image (operator). The cover
// is always stored as a public asset, so it renders on browse/detail regardless
// of the resource's access gate.
type AddCoverReq struct {
	g.Meta   `path:"/api/v1/resources/{id}/cover" method:"post" tags:"resource" summary:"Begin a cover-image upload"`
	ID       string `json:"id" in:"path" v:"required"`
	Filename string `json:"filename" v:"required"`
	Mime     string `json:"mime"`
	Size     int64  `json:"size" v:"required|min:1"`
}

type AddCoverRes struct {
	UploadURL     string            `json:"uploadUrl"`
	UploadToken   string            `json:"uploadToken"`
	Method        string            `json:"method"`
	UploadHeaders map[string]string `json:"uploadHeaders,omitempty"`
}

// FinalizeCoverReq finalizes the uploaded cover and points the resource at it.
type FinalizeCoverReq struct {
	g.Meta      `path:"/api/v1/resources/{id}/cover/finalize" method:"post" tags:"resource" summary:"Finalize + set a resource cover"`
	ID          string `json:"id" in:"path" v:"required"`
	UploadToken string `json:"uploadToken" v:"required"`
}

type FinalizeCoverRes struct {
	CoverAssetID string `json:"coverAssetId"`
	CoverURL     string `json:"coverUrl"`
}
