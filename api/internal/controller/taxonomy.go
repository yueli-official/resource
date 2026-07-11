package controller

import (
	"context"

	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/catalog"
)

// Taxonomy handles the operator (JWT) category/tag management endpoints.
type Taxonomy struct{ svc *catalog.Service }

func NewTaxonomy(svc *catalog.Service) *Taxonomy { return &Taxonomy{svc: svc} }

func (c *Taxonomy) CreateTaxonomy(ctx context.Context, req *v1.CreateTaxonomyReq) (*v1.CreateTaxonomyRes, error) {
	// categories are curated structure (superadmin only); tags are folksonomy any
	// logged-in operator may create freely.
	if req.Taxonomy == "category" {
		if err := requireAdmin(ctx); err != nil {
			return nil, err
		}
	} else if _, err := subject(ctx); err != nil {
		return nil, err
	}
	tx, err := c.svc.CreateTaxonomy(ctx, req.Name, req.Slug, req.Taxonomy, req.ParentID, req.Description)
	if err != nil {
		return nil, err
	}
	return &v1.CreateTaxonomyRes{Taxonomy: taxonomyView(tx)}, nil
}

// UpdateTaxonomy renames / re-slugs / re-parents a taxonomy (superadmin only).
func (c *Taxonomy) UpdateTaxonomy(ctx context.Context, req *v1.UpdateTaxonomyReq) (*v1.UpdateTaxonomyRes, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	tx, err := c.svc.UpdateTaxonomy(ctx, req.ID, req.Name, req.Slug, req.Description, req.ParentID)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateTaxonomyRes{Taxonomy: taxonomyView(tx)}, nil
}

// DeleteTaxonomy removes a taxonomy (superadmin only; refused if it has children).
func (c *Taxonomy) DeleteTaxonomy(ctx context.Context, req *v1.DeleteTaxonomyReq) (*v1.DeleteTaxonomyRes, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := c.svc.DeleteTaxonomy(ctx, req.ID); err != nil {
		return nil, err
	}
	return &v1.DeleteTaxonomyRes{Deleted: true}, nil
}

// MergeTaxonomy folds one taxonomy into another (superadmin only).
func (c *Taxonomy) MergeTaxonomy(ctx context.Context, req *v1.MergeTaxonomyReq) (*v1.MergeTaxonomyRes, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := c.svc.MergeTaxonomy(ctx, req.ID, req.TargetID); err != nil {
		return nil, err
	}
	return &v1.MergeTaxonomyRes{Merged: true}, nil
}

func (c *Taxonomy) AssignTaxonomies(ctx context.Context, req *v1.AssignTaxonomiesReq) (*v1.AssignTaxonomiesRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.svc.AssignTaxonomies(ctx, owner, req.ID, req.TaxonomyIDs); err != nil {
		return nil, err
	}
	return &v1.AssignTaxonomiesRes{Updated: true}, nil
}
