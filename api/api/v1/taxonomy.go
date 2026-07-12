package v1

import "github.com/gogf/gf/v2/frame/g"

// TaxonomyView is the outward projection of a category/tag.
type TaxonomyView struct {
	ID          string `json:"id"`
	Taxonomy    string `json:"taxonomy"` // "category" | "tag"
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parentId,omitempty"`
	Count       int    `json:"count"` // published resources under this taxonomy
}

// ── browse (public) ──────────────────────────────────────────────────────────

type ListTaxonomiesReq struct {
	g.Meta    `path:"/api/v1/taxonomies" method:"get" tags:"resource" summary:"List taxonomies (category/tag)"`
	Taxonomy  string `json:"taxonomy"` // optional kind filter
	Q         string `json:"q"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

type ListTaxonomiesRes struct {
	Items []*TaxonomyView `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// ── manage (operator JWT) ────────────────────────────────────────────────────

type CreateTaxonomyReq struct {
	g.Meta      `path:"/api/v1/taxonomies" method:"post" tags:"resource" summary:"Create a category/tag"`
	Name        string `json:"name" v:"required"`
	Slug        string `json:"slug"`
	Taxonomy    string `json:"taxonomy" v:"required|in:category,tag"`
	ParentID    string `json:"parentId"`
	Description string `json:"description"`
}

type CreateTaxonomyRes struct {
	Taxonomy *TaxonomyView `json:"taxonomy"`
}

type AssignTaxonomiesReq struct {
	g.Meta      `path:"/api/v1/resources/{id}/taxonomies" method:"put" tags:"resource" summary:"Set a resource's taxonomies"`
	ID          string   `json:"id" in:"path" v:"required"`
	TaxonomyIDs []string `json:"taxonomyIds"`
}

type AssignTaxonomiesRes struct {
	Updated bool `json:"updated"`
}

// ── governance (superadmin JWT) ──────────────────────────────────────────────

// UpdateTaxonomyReq renames / re-slugs / re-describes / re-parents a taxonomy.
// Pointer fields distinguish "omitted" from "set"; a non-nil empty parentId
// clears the parent (promotes to a root).
type UpdateTaxonomyReq struct {
	g.Meta      `path:"/api/v1/taxonomies/{id}" method:"patch" tags:"resource" summary:"Update a taxonomy (admin)"`
	ID          string  `json:"id" in:"path" v:"required"`
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	ParentID    *string `json:"parentId"`
}

type UpdateTaxonomyRes struct {
	Taxonomy *TaxonomyView `json:"taxonomy"`
}

type DeleteTaxonomyReq struct {
	g.Meta `path:"/api/v1/taxonomies/{id}" method:"delete" tags:"resource" summary:"Delete a taxonomy (admin; refused if it has children)"`
	ID     string `json:"id" in:"path" v:"required"`
}

type DeleteTaxonomyRes struct {
	Deleted bool `json:"deleted"`
}

// MergeTaxonomyReq folds the source taxonomy ({id}) into the target: resources
// and child taxonomies move to target, then source is deleted.
type MergeTaxonomyReq struct {
	g.Meta   `path:"/api/v1/taxonomies/{id}/merge" method:"post" tags:"resource" summary:"Merge a taxonomy into another (admin)"`
	ID       string `json:"id" in:"path" v:"required"`
	TargetID string `json:"targetId" v:"required"`
}

type MergeTaxonomyRes struct {
	Merged bool `json:"merged"`
}
