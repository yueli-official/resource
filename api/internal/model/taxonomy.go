package model

// Term is a reusable taxonomy label (shared across category/tag).
type Term struct {
	ID   string `json:"id" orm:"id"`
	Name string `json:"name" orm:"name"`
	Slug string `json:"slug" orm:"slug"`
}

// Taxonomy is a category or tag (a term applied under a taxonomy kind). Name/Slug
// are joined from terms for views and are not stored on this table. Count is the
// number of published resources under the taxonomy (computed on read).
type Taxonomy struct {
	ID                string `json:"id" orm:"id"`
	CatalogID         string `json:"catalogId" orm:"catalog_id"`
	TermID            string `json:"termId" orm:"term_id"`
	Taxonomy          string `json:"taxonomy" orm:"taxonomy"`
	Description       string `json:"description" orm:"description"`
	ParentID          string `json:"parentId" orm:"parent_id"`
	Status            string `json:"status" orm:"status"`
	EditorialPosition *int   `json:"editorialPosition,omitempty" orm:"editorial_position"`
	ReplacementID     string `json:"replacementId,omitempty" orm:"replacement_id"`
	Count             int    `json:"count" orm:"post_count"`
	Name              string `json:"name"` // joined from terms (LeftJoin in ListTaxonomies)
	Slug              string `json:"slug"` // joined from terms
}
