package model

// SEO is a resource's SEO metadata (1:1 with the resource).
type SEO struct {
	ResourceID   string `json:"resourceId" orm:"resource_id"`
	MetaTitle    string `json:"metaTitle" orm:"meta_title"`
	MetaDesc     string `json:"metaDesc" orm:"meta_desc"`
	OgTitle      string `json:"ogTitle" orm:"og_title"`
	OgImage      string `json:"ogImage" orm:"og_image"`
	CanonicalURL string `json:"canonicalUrl" orm:"canonical_url"`
	Robots       string `json:"robots" orm:"robots"`
}
