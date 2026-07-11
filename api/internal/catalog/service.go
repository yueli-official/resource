// Package catalog is the resource site's core logic: catalog CRUD, asset
// orchestration through the asset service, and public download delivery. The
// resource site is fully free — every resource is world-readable and its files
// are delivered straight from the asset service's public CDN URL (paid / gated
// delivery moved to the mall ⑤, per the content-kit-resource decision).
package catalog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"platform/products/resource/api/internal/assetclient"
	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/model"
	"platform/products/resource/api/internal/reserr"
)

// TypeRule is one configured resource type's upload policy.
type TypeRule struct {
	Label      string
	AllowedExt map[string]bool
	MaxSizeMB  int
}

// Service owns the catalog logic.
type Service struct {
	dao           *dao.PG
	asset         assetclient.Client
	types         map[string]TypeRule
	coverCategory string
	siteBrand     string
}

func New(d *dao.PG, asset assetclient.Client, types map[string]TypeRule, coverCategory, siteBrand string) *Service {
	if strings.TrimSpace(siteBrand) == "" {
		siteBrand = "资源库"
	}
	return &Service{dao: d, asset: asset, types: types, coverCategory: coverCategory, siteBrand: siteBrand}
}

// ── create / read ────────────────────────────────────────────────────────────

type CreateInput struct {
	Title, Summary, Description, Type, CoverAssetID string
	Tags                                            []string
}

// Create inserts a draft resource (validates type, generates a unique slug).
func (s *Service) Create(ctx context.Context, owner string, in CreateInput) (*model.Resource, error) {
	if _, ok := s.types[in.Type]; !ok {
		return nil, reserr.InvalidType(in.Type)
	}
	r := &model.Resource{
		OwnerID: owner, Title: in.Title, Summary: in.Summary, Description: in.Description,
		Type: in.Type, CoverAssetID: in.CoverAssetID,
		DeliveryKind: "asset_file", DeliveryPayload: "{}",
		Status: model.StatusDraft, Tags: in.Tags,
	}
	if err := s.insertWithSlug(ctx, r, slugify(in.Title)); err != nil {
		return nil, err
	}
	return s.dao.GetByID(ctx, r.ID)
}

// insertWithSlug tries base, base-2..4, then base-<rand> until the slug is free.
func (s *Service) insertWithSlug(ctx context.Context, r *model.Resource, base string) error {
	if base == "" {
		base = "r"
	}
	candidates := []string{base, base + "-2", base + "-3", base + "-4"}
	for i := 0; i < 8; i++ {
		var slug string
		if i < len(candidates) {
			slug = candidates[i]
		} else {
			slug = base + "-" + randHex(3)
		}
		r.Slug = slug
		err := s.dao.Insert(ctx, r)
		if err == nil {
			return nil
		}
		if err != dao.ErrSlugTaken {
			return err
		}
	}
	return reserr.SlugTaken(base)
}

// Get returns a resource visible to the viewer (published, or draft owned by viewer).
func (s *Service) Get(ctx context.Context, viewer, id string) (*model.Resource, error) {
	r, err := s.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil || (r.Status != model.StatusPublished && r.OwnerID != viewer) {
		return nil, reserr.NotFound(id)
	}
	return r, nil
}

// List returns published resources matching the filter, plus the total.
func (s *Service) List(ctx context.Context, f dao.ListFilter, page, size int) ([]*model.Resource, int, error) {
	page, size = norm(page, size)
	return s.dao.List(ctx, f, size, (page-1)*size)
}

// ListMine returns the owner's resources of any status (operator console),
// plus the normalized page/size and total.
func (s *Service) ListMine(ctx context.Context, owner string, f dao.OwnerListFilter, page, size int) ([]*model.Resource, int, int, int, error) {
	page, size = norm(page, size)
	items, total, err := s.dao.ListByOwner(ctx, owner, f, size, (page-1)*size)
	return items, total, page, size, err
}

// Assets returns a resource's files.
func (s *Service) Assets(ctx context.Context, resourceID string) ([]*model.ResourceAsset, error) {
	return s.dao.ListAssets(ctx, resourceID)
}

// ── patch / delete ───────────────────────────────────────────────────────────

// Patch updates mutable fields (never owner). Promoting to published enforces the
// publish constraints. fields is the provided subset.
func (s *Service) Patch(ctx context.Context, owner, id string, fields g.Map) (*model.Resource, error) {
	cur, err := s.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil || cur.OwnerID != owner {
		return nil, reserr.NotFound(id)
	}
	if t, ok := fields["type"].(string); ok {
		if _, known := s.types[t]; !known {
			return nil, reserr.InvalidType(t)
		}
	}
	if rawSlug, ok := fields["slug"].(string); ok {
		sl := slugify(rawSlug)
		if sl == "" {
			return nil, reserr.InvalidInput("slug produces an empty value")
		}
		if sl != cur.Slug {
			existing, err := s.dao.GetBySlug(ctx, sl)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != id {
				return nil, reserr.SlugTaken(sl)
			}
		}
		fields["slug"] = sl
	}
	if st, ok := fields["status"].(string); ok && st == string(model.StatusPublished) {
		payload := cur.DeliveryPayload
		if next, ok := fields["delivery_payload"].(string); ok {
			payload = next
		}
		if err := s.checkPublishable(ctx, cur, payload); err != nil {
			return nil, err
		}
	}
	n, err := s.dao.Patch(ctx, owner, id, fields)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, reserr.NotFound(id)
	}
	return s.dao.GetByID(ctx, id)
}

// checkPublishable enforces that a published resource has at least one concrete
// delivery target: an uploaded file or an enabled netdisk link.
func (s *Service) checkPublishable(ctx context.Context, cur *model.Resource, payload string) error {
	n, err := s.dao.CountAssets(ctx, cur.ID)
	if err != nil {
		return err
	}
	if n == 0 && !hasEnabledNetdisk(payload) {
		return reserr.InvalidState("a published resource needs at least one file or netdisk delivery")
	}
	return nil
}

// Delete removes a resource: best-effort deletes each asset object (bearer = the
// owner's token), then the row (resource_assets cascade).
func (s *Service) Delete(ctx context.Context, owner, bearer, id string) error {
	assets, err := s.dao.ListAssets(ctx, id)
	if err != nil {
		return err
	}
	r, err := s.dao.Delete(ctx, owner, id) // returns nil if absent/not owner
	if err != nil {
		return err
	}
	if r == nil {
		return reserr.NotFound(id)
	}
	for _, a := range assets {
		_ = s.asset.UnregisterReference(ctx, bearer, assetclient.ReferenceInput{
			AssetID: a.AssetID, RefType: "resource-file", RefID: id,
		})
		_ = s.asset.Delete(ctx, bearer, a.AssetID) // best-effort
	}
	if r.CoverAssetID != "" {
		_ = s.asset.UnregisterReference(ctx, bearer, assetclient.ReferenceInput{
			AssetID: r.CoverAssetID, RefType: "resource-cover", RefID: id,
		})
		_ = s.asset.Delete(ctx, bearer, r.CoverAssetID)
	}
	return nil
}

// ── asset add / finalize / remove ────────────────────────────────────────────

// AddAsset validates the file against the resource's type policy and opens an
// upload on the asset service. Every file is a public asset (the site is free),
// delivered from the stable CDN URL. Returns the blob link + the upload token.
func (s *Service) AddAsset(ctx context.Context, owner, bearer, id, filename string, size int64, multipart bool) (assetclient.InitOutput, error) {
	r, err := s.ownedDraftable(ctx, owner, id)
	if err != nil {
		return assetclient.InitOutput{}, err
	}
	rule := s.types[r.Type]
	ext := extOf(filename)
	if !rule.AllowedExt[ext] {
		return assetclient.InitOutput{}, reserr.InvalidType(ext)
	}
	if rule.MaxSizeMB > 0 && size > int64(rule.MaxSizeMB)*1024*1024 {
		return assetclient.InitOutput{}, reserr.InvalidState("file exceeds the type size limit")
	}
	return s.asset.UploadInit(ctx, bearer, assetclient.InitInput{
		Filename: filename, Mime: "application/octet-stream", Category: "resource", Visibility: "public", Size: size, Multipart: multipart,
	})
}

func (s *Service) MultipartPartURL(ctx context.Context, owner, bearer, id string, in assetclient.MultipartPartURLInput) (assetclient.MultipartPartURLOutput, error) {
	if _, err := s.ownedDraftable(ctx, owner, id); err != nil {
		return assetclient.MultipartPartURLOutput{}, err
	}
	return s.asset.MultipartPartURL(ctx, bearer, in)
}

func (s *Service) CompleteMultipart(ctx context.Context, owner, bearer, id string, in assetclient.MultipartCompleteInput) error {
	if _, err := s.ownedDraftable(ctx, owner, id); err != nil {
		return err
	}
	return s.asset.CompleteMultipart(ctx, bearer, in)
}

func (s *Service) AbortMultipart(ctx context.Context, owner, bearer, id string, in assetclient.MultipartAbortInput) error {
	if _, err := s.ownedDraftable(ctx, owner, id); err != nil {
		return err
	}
	return s.asset.AbortMultipart(ctx, bearer, in)
}

// FinalizeAsset finalizes an uploaded blob and links it to the resource.
func (s *Service) FinalizeAsset(ctx context.Context, owner, bearer, id, uploadToken, label string) (*model.ResourceAsset, error) {
	r, err := s.ownedDraftable(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	view, err := s.asset.Finalize(ctx, bearer, uploadToken)
	if err != nil {
		return nil, err
	}
	n, _ := s.dao.CountAssets(ctx, r.ID)
	ra := &model.ResourceAsset{
		ResourceID: r.ID, AssetID: view.ID, Label: label, CdnURL: view.CdnURL,
		Size: view.Size, Mime: view.Mime, Filename: view.Filename, Sort: n,
	}
	if err := s.dao.InsertAsset(ctx, ra); err != nil {
		return nil, err
	}
	if err := s.asset.RegisterReference(ctx, bearer, assetclient.ReferenceInput{
		AssetID: view.ID, RefType: "resource-file", RefID: r.ID,
		RefLabel: firstNonEmpty(label, r.Title), RefURL: "/resources/" + r.ID,
	}); err != nil {
		return nil, err
	}
	return ra, nil
}

// RemoveAsset removes a file from a resource (a published resource keeps ≥1 file).
func (s *Service) RemoveAsset(ctx context.Context, owner, bearer, id, assetID string) error {
	r, err := s.dao.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if r == nil || r.OwnerID != owner {
		return reserr.NotFound(id)
	}
	if r.Status == model.StatusPublished {
		if n, _ := s.dao.CountAssets(ctx, id); n <= 1 && !hasEnabledNetdisk(r.DeliveryPayload) {
			return reserr.InvalidState("a published resource needs at least one file or netdisk delivery")
		}
	}
	ra, err := s.dao.DeleteAsset(ctx, id, assetID)
	if err != nil {
		return err
	}
	if ra == nil {
		return reserr.AssetNotFound(assetID)
	}
	_ = s.asset.UnregisterReference(ctx, bearer, assetclient.ReferenceInput{
		AssetID: assetID, RefType: "resource-file", RefID: id,
	})
	_ = s.asset.Delete(ctx, bearer, assetID) // best-effort
	return nil
}

// ── cover image (public asset, set on the resource) ──────────────────────────

// AddCover opens an upload for a resource's cover image (always a public asset in
// the configured cover category). Returns the blob link + upload token.
func (s *Service) AddCover(ctx context.Context, owner, bearer, id, filename, mime string, size int64) (assetclient.InitOutput, error) {
	if _, err := s.ownedDraftable(ctx, owner, id); err != nil {
		return assetclient.InitOutput{}, err
	}
	if !isImageExt(extOf(filename)) {
		return assetclient.InitOutput{}, reserr.InvalidType(extOf(filename))
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	return s.asset.UploadInit(ctx, bearer, assetclient.InitInput{
		Filename: filename, Mime: mime, Category: s.coverCategory, Visibility: "public", Size: size,
	})
}

// FinalizeCover finalizes the uploaded cover, points the resource at it
// (cover_asset_id + cover_url snapshot), and best-effort removes the old cover.
func (s *Service) FinalizeCover(ctx context.Context, owner, bearer, id, uploadToken string) (assetID, coverURL string, err error) {
	r, err := s.ownedDraftable(ctx, owner, id)
	if err != nil {
		return "", "", err
	}
	old := r.CoverAssetID
	view, err := s.asset.Finalize(ctx, bearer, uploadToken)
	if err != nil {
		return "", "", err
	}
	if _, err := s.dao.Patch(ctx, owner, id, g.Map{"cover_asset_id": view.ID, "cover_url": view.CdnURL}); err != nil {
		return "", "", err
	}
	if err := s.asset.RegisterReference(ctx, bearer, assetclient.ReferenceInput{
		AssetID: view.ID, RefType: "resource-cover", RefID: r.ID,
		RefLabel: r.Title, RefURL: "/resources/" + r.ID,
	}); err != nil {
		return "", "", err
	}
	if old != "" && old != view.ID {
		_ = s.asset.UnregisterReference(ctx, bearer, assetclient.ReferenceInput{
			AssetID: old, RefType: "resource-cover", RefID: id,
		})
		_ = s.asset.Delete(ctx, bearer, old) // best-effort
	}
	return view.ID, view.CdnURL, nil
}

// ── download ──────────────────────────────────────────────────────────────────

// Download resolves the public delivery URL for one of a resource's files and
// counts the download (best-effort). Every resource is free → the stable public
// CDN URL (captured at finalize) is returned directly, no gating.
func (s *Service) Download(ctx context.Context, viewer, id, assetID string) (string, error) {
	if _, err := s.Get(ctx, viewer, id); err != nil {
		return "", err
	}
	ra, err := s.dao.GetAsset(ctx, id, assetID)
	if err != nil {
		return "", err
	}
	if ra == nil {
		return "", reserr.AssetNotFound(assetID)
	}
	_ = s.dao.IncrementDownload(ctx, id) // best-effort
	return ra.CdnURL, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *Service) ownedDraftable(ctx context.Context, owner, id string) (*model.Resource, error) {
	r, err := s.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil || r.OwnerID != owner {
		return nil, reserr.NotFound(id)
	}
	return r, nil
}

func norm(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return page, size
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
		case b.Len() > 0 && !prevHyphen:
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func extOf(filename string) string {
	i := strings.LastIndexByte(filename, '.')
	if i < 0 {
		return ""
	}
	return strings.ToLower(filename[i+1:])
}

func isImageExt(ext string) bool {
	switch ext {
	case "jpg", "jpeg", "png", "webp":
		return true
	}
	return false
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func hasEnabledNetdisk(raw string) bool {
	type netdisk struct {
		URL string `json:"url"`
	}
	type item struct {
		Kind    string   `json:"kind"`
		Enabled *bool    `json:"enabled"`
		Netdisk *netdisk `json:"netdisk"`
	}
	type payload struct {
		Items []item `json:"items"`
	}
	var p payload
	if strings.TrimSpace(raw) == "" || json.Unmarshal([]byte(raw), &p) != nil {
		return false
	}
	for _, item := range p.Items {
		if item.Kind != "netdisk" {
			continue
		}
		if item.Enabled != nil && !*item.Enabled {
			continue
		}
		if item.Netdisk != nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(item.Netdisk.URL)), "http") {
			return true
		}
	}
	return false
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
