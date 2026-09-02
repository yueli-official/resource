// Package dao is the PostgreSQL data-access layer for the resource catalog.
package dao

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"github.com/yueli-official/foundation/go/identifier"
	"github.com/yueli-official/resource/api/internal/model"
)

const (
	tResources = "resources"
	tAssets    = "resource_assets"
)

// ErrSlugTaken is returned by Insert when the slug unique index is violated.
var ErrSlugTaken = errors.New("dao: slug taken")

// PG wraps the GoFrame gdb handle.
type PG struct{ db gdb.DB }

func NewPG(db gdb.DB) *PG { return &PG{db: db} }

func tagsJSON(tags []string) string {
	if tags == nil {
		tags = []string{}
	}
	b, _ := json.Marshal(tags)
	return string(b)
}

func firstJSON(raw, fallback string) string {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	return raw
}

// ListFilter narrows a List query (empty fields ignored; Status defaults to
// published). Q is a site-search query (case-insensitive substring over
// title/summary/description).
type ListFilter struct {
	Type         string
	Tags         []string
	Status       string
	Q            string
	TaxonomySlug string // category/tag term slug filter (resolves via object_taxonomies)
}

// OwnerListFilter narrows the operator's own resource list.
type OwnerListFilter struct {
	Status    string
	Issues    bool
	Q         string
	SortBy    string
	SortOrder string
}

// Insert writes a new resource (generating id when unset). A slug unique
// violation maps to ErrSlugTaken so the service can retry with a suffix.
func (p *PG) Insert(ctx context.Context, r *model.Resource) error {
	if r.ID == "" {
		r.ID = identifier.MustNew().String()
	}
	data := g.Map{
		"id": r.ID, "owner_id": r.OwnerID, "title": r.Title, "slug": r.Slug,
		"summary": r.Summary, "description": r.Description, "type": r.Type,
		"cover_asset_id": r.CoverAssetID, "cover_url": r.CoverURL,
		"delivery_kind": r.DeliveryKind, "delivery_payload": firstJSON(r.DeliveryPayload, "{}"),
		"status": string(r.Status), "published_at": r.PublishedAt, "view_count": r.ViewCount,
		"download_count": r.DownloadCount, "tags": tagsJSON(r.Tags),
	}
	if _, err := p.db.Model(tResources).Ctx(ctx).Data(data).Insert(); err != nil {
		if isDupSlug(err) {
			return ErrSlugTaken
		}
		return err
	}
	return nil
}

// GetByID returns the resource, or (nil, nil) when absent.
func (p *PG) GetByID(ctx context.Context, id string) (*model.Resource, error) {
	return p.one(ctx, "id", id)
}

// GetBySlug returns the resource by slug, or (nil, nil).
func (p *PG) GetBySlug(ctx context.Context, slug string) (*model.Resource, error) {
	return p.one(ctx, "slug", slug)
}

// SlugExists reports whether a slug is taken.
func (p *PG) SlugExists(ctx context.Context, slug string) (bool, error) {
	n, err := p.db.Model(tResources).Ctx(ctx).Where("slug", slug).Count()
	return n > 0, err
}

// List returns published resources matching the filter (newest first by default)
// plus the total count. Status defaults to published.
func (p *PG) List(ctx context.Context, f ListFilter, limit, offset int) ([]*model.Resource, int, error) {
	status := f.Status
	if status == "" {
		status = string(model.StatusPublished)
	}
	m := p.db.Model(tResources).Ctx(ctx).Where("status", status)
	if f.TaxonomySlug != "" {
		ids, err := p.ResourceIDsByTaxonomySlug(ctx, f.TaxonomySlug)
		if err != nil {
			return nil, 0, err
		}
		if len(ids) == 0 {
			return []*model.Resource{}, 0, nil
		}
		m = m.WhereIn("id", ids)
	}
	if f.Type != "" {
		m = m.Where("type", f.Type)
	}
	if len(f.Tags) > 0 {
		m = m.Where("tags @> ?::jsonb", tagsJSON(f.Tags)) // contains ALL (AND)
	}
	total, err := m.Clone().Count()
	if err != nil {
		return nil, 0, err
	}
	var out []*model.Resource
	if err := m.Order("COALESCE(published_at, created_at) DESC").Limit(limit).Offset(offset).Scan(&out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (p *PG) ListPublishedByIDs(ctx context.Context, ids []string) ([]*model.Resource, error) {
	if len(ids) == 0 {
		return []*model.Resource{}, nil
	}
	var out []*model.Resource
	err := p.db.Model(tResources).Ctx(ctx).WhereIn("id", ids).Where("status", model.StatusPublished).Scan(&out)
	return out, err
}

// ListByOwner returns the owner's resources of any status (for "my resources").
func (p *PG) ListByOwner(ctx context.Context, owner string, f OwnerListFilter, limit, offset int) ([]*model.Resource, int, error) {
	m := p.db.Model(tResources).Ctx(ctx)
	if strings.TrimSpace(owner) != "" {
		m = m.Where("owner_id", owner)
	}
	if f.Status != "" {
		m = m.Where("status", f.Status)
	}
	if f.Issues {
		m = m.Where(resourceIssuePredicate)
	}
	if f.Q != "" {
		like := "%" + f.Q + "%"
		m = m.Where("(title ILIKE ? OR summary ILIKE ? OR description ILIKE ?)", like, like, like)
	}
	total, err := m.Clone().Count()
	if err != nil {
		return nil, 0, err
	}
	sortCol := "updated_at"
	switch f.SortBy {
	case "title":
		sortCol = "title"
	case "createdAt":
		sortCol = "created_at"
	case "updatedAt":
		sortCol = "updated_at"
	case "publishedAt":
		sortCol = "published_at"
	case "viewCount":
		sortCol = "view_count"
	case "downloadCount":
		sortCol = "download_count"
	}
	var out []*model.Resource
	if strings.EqualFold(f.SortOrder, "asc") {
		m = m.OrderAsc(sortCol)
	} else {
		m = m.OrderDesc(sortCol)
	}
	m = m.Fields("resources.*", "(CASE WHEN "+resourceIssuePredicate+" THEN 1 ELSE 0 END) AS issue_count")
	if err := m.Limit(limit).Offset(offset).Scan(&out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

const resourceIssuePredicate = `(
  btrim(coalesce(summary, '')) = ''
  OR (btrim(coalesce(cover_asset_id, '')) = '' AND btrim(coalesce(cover_url, '')) = '')
  OR (
    NOT EXISTS (SELECT 1 FROM resource_assets ra WHERE ra.resource_id = resources.id)
    AND NOT EXISTS (
      SELECT 1
      FROM jsonb_array_elements(coalesce(delivery_payload->'items', '[]'::jsonb)) item
      WHERE item->>'kind' = 'netdisk'
        AND coalesce((item->>'enabled')::boolean, true)
        AND lower(btrim(coalesce(item->'netdisk'->>'url', ''))) LIKE 'http%'
    )
  )
)`

func (p *PG) OwnerResourceCounts(ctx context.Context, owner string) (model.ResourceLifecycleCounts, error) {
	var row struct {
		All       int `orm:"all_count"`
		Published int `orm:"published_count"`
		Draft     int `orm:"draft_count"`
		Archived  int `orm:"archived_count"`
		Issues    int `orm:"issues_count"`
	}
	err := p.db.Ctx(ctx).Raw(`
SELECT
  count(*) AS all_count,
  count(*) FILTER (WHERE status = 'published') AS published_count,
  count(*) FILTER (WHERE status = 'draft') AS draft_count,
  count(*) FILTER (WHERE status = 'archived') AS archived_count,
  count(*) FILTER (WHERE `+resourceIssuePredicate+`) AS issues_count
FROM resources
WHERE (? = '' OR owner_id = ?)`, owner, owner).Scan(&row)
	return model.ResourceLifecycleCounts{
		All: row.All, Published: row.Published, Draft: row.Draft,
		Archived: row.Archived, Issues: row.Issues,
	}, err
}

func (p *PG) OwnedByID(ctx context.Context, owner, id string) (*model.Resource, error) {
	return p.oneOwned(ctx, owner, id)
}

func (p *PG) ApplyOwnerResourceBatch(ctx context.Context, owner string, ids []string, status string) ([]string, error) {
	return p.ApplyOwnerResourceBatchWithHook(ctx, owner, ids, status, nil)
}

func (p *PG) ApplyOwnerResourceBatchWithHook(ctx context.Context, owner string, ids []string, status string, hook TransactionHook) ([]string, error) {
	query, args := ownerResourceBatchSQL(owner, ids, status)
	if query == "" {
		return []string{}, nil
	}
	var rows []struct {
		ID string `orm:"id"`
	}
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := tx.Ctx(ctx).Raw(query, args...).Scan(&rows); err != nil {
			return err
		}
		return runTransactionHook(ctx, tx, hook)
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ID)
	}
	return out, nil
}

func ownerResourceBatchSQL(owner string, ids []string, status string) (string, []any) {
	if len(ids) == 0 {
		return "", nil
	}
	placeholders := make([]string, 0, len(ids))
	args := []any{status, status}
	ownerClause := ""
	if strings.TrimSpace(owner) != "" {
		ownerClause = "  AND owner_id = ?\n"
		args = append(args, owner)
	}
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, status)
	return `UPDATE resources
SET status = ?,
    published_at = CASE WHEN ? = 'published' AND published_at IS NULL THEN now() ELSE published_at END,
    updated_at = now()
WHERE TRUE
` + ownerClause + `  AND id IN (` + strings.Join(placeholders, ", ") + `)
  AND (? <> 'published' OR (
    EXISTS (SELECT 1 FROM resource_assets ra WHERE ra.resource_id = resources.id)
    OR EXISTS (
      SELECT 1
      FROM jsonb_array_elements(coalesce(delivery_payload->'items', '[]'::jsonb)) item
      WHERE item->>'kind' = 'netdisk'
        AND coalesce((item->>'enabled')::boolean, true)
        AND lower(btrim(coalesce(item->'netdisk'->>'url', ''))) LIKE 'http%'
    )
  ))
RETURNING id`, args
}

// Patch updates mutable fields for the owner's resource (owner never changes);
// returns rows affected (0 = absent/not owner).
func (p *PG) Patch(ctx context.Context, owner, id string, data g.Map) (int64, error) {
	return p.PatchWithHook(ctx, owner, id, data, nil)
}

func (p *PG) PatchWithHook(ctx context.Context, owner, id string, data g.Map, hook TransactionHook) (int64, error) {
	var affected int64
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		data["updated_at"] = gtime.Now()
		if tags, ok := data["tags"].([]string); ok {
			data["tags"] = tagsJSON(tags)
		}
		result, err := tx.Model(tResources).Ctx(ctx).Where("owner_id", owner).Where("id", id).Data(data).Update()
		if err != nil {
			return err
		}
		affected, err = result.RowsAffected()
		if err != nil || affected == 0 {
			return err
		}
		return runTransactionHook(ctx, tx, hook)
	})
	return affected, err
}

// AdvanceViewProjection moves the catalog read projection to a traffic-module
// total without allowing replay or out-of-order completion to move it backward.
func (p *PG) AdvanceViewProjection(ctx context.Context, id string, views int64) error {
	_, err := p.db.Exec(ctx,
		"UPDATE resources SET view_count = GREATEST(view_count, ?), updated_at = NOW() WHERE id = ?",
		views, id,
	)
	return err
}

// ReplaceViewProjection repairs catalog projection drift from Traffic truth.
func (p *PG) ReplaceViewProjection(ctx context.Context, id string, views int64) error {
	_, err := p.db.Exec(ctx,
		"UPDATE resources SET view_count = ?, updated_at = NOW() WHERE id = ?",
		views, id,
	)
	return err
}

func (p *PG) ListResourceIDs(ctx context.Context) ([]string, error) {
	var rows []struct {
		ID string `orm:"id"`
	}
	if err := p.db.Model(tResources).Ctx(ctx).Fields("id").Scan(&rows); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

// Delete removes the owner's resource, returning it (for cleanup) or nil if absent.
func (p *PG) Delete(ctx context.Context, owner, id string) (*model.Resource, error) {
	return p.DeleteWithHook(ctx, owner, id, nil)
}

func (p *PG) DeleteWithHook(ctx context.Context, owner, id string, hook TransactionHook) (*model.Resource, error) {
	r, err := p.oneOwned(ctx, owner, id)
	if err != nil || r == nil {
		return nil, err
	}
	err = p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(tResources).Ctx(ctx).Where("owner_id", owner).Where("id", id).Delete(); err != nil {
			return err
		}
		return runTransactionHook(ctx, tx, hook)
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// IncrementDownload bumps the counter (best-effort; caller ignores the error).
func (p *PG) IncrementDownload(ctx context.Context, id string) error {
	_, err := p.db.Model(tResources).Ctx(ctx).Where("id", id).Increment("download_count", 1)
	return err
}

// ── resource_assets ──────────────────────────────────────────────────────────

// InsertAsset links an asset to a resource.
func (p *PG) InsertAsset(ctx context.Context, a *model.ResourceAsset) error {
	if a.ID == "" {
		a.ID = identifier.MustNew().String()
	}
	_, err := p.db.Model(tAssets).Ctx(ctx).Data(g.Map{
		"id": a.ID, "resource_id": a.ResourceID, "asset_id": a.AssetID,
		"label": a.Label, "media_key": a.MediaKey, "size": a.Size,
		"mime": a.Mime, "filename": a.Filename, "sort": a.Sort,
	}).Insert()
	return err
}

// ListAssets returns a resource's files (by sort).
func (p *PG) ListAssets(ctx context.Context, resourceID string) ([]*model.ResourceAsset, error) {
	var out []*model.ResourceAsset
	err := p.db.Model(tAssets).Ctx(ctx).Where("resource_id", resourceID).OrderAsc("sort").Scan(&out)
	return out, err
}

// GetAsset returns a specific file within a resource, or (nil, nil).
func (p *PG) GetAsset(ctx context.Context, resourceID, assetID string) (*model.ResourceAsset, error) {
	var a *model.ResourceAsset
	if err := p.db.Model(tAssets).Ctx(ctx).
		Where("resource_id", resourceID).Where("asset_id", assetID).Limit(1).Scan(&a); err != nil {
		return nil, err
	}
	return a, nil
}

// CountAssets returns how many files a resource has.
func (p *PG) CountAssets(ctx context.Context, resourceID string) (int, error) {
	return p.db.Model(tAssets).Ctx(ctx).Where("resource_id", resourceID).Count()
}

// DeleteAsset removes one file link, returning it (for object cleanup) or nil.
func (p *PG) DeleteAsset(ctx context.Context, resourceID, assetID string) (*model.ResourceAsset, error) {
	a, err := p.GetAsset(ctx, resourceID, assetID)
	if err != nil || a == nil {
		return nil, err
	}
	if _, err := p.db.Model(tAssets).Ctx(ctx).Where("id", a.ID).Delete(); err != nil {
		return nil, err
	}
	return a, nil
}

func (p *PG) HomeSettings(ctx context.Context) (map[string]any, error) {
	var row *model.HomeSettings
	if err := p.db.Model("resource_home_settings").Ctx(ctx).Where("id", 1).Limit(1).Scan(&row); err != nil {
		return nil, err
	}
	if row == nil || strings.TrimSpace(row.Payload) == "" {
		return nil, gerror.New("resource homepage configuration is not seeded")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return nil, gerror.Wrap(err, "decode resource homepage configuration")
	}
	return payload, nil
}

func (p *PG) SaveHomeSettings(ctx context.Context, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `
INSERT INTO resource_home_settings (id, payload)
VALUES (1, ?::jsonb)
ON CONFLICT (id) DO UPDATE SET payload = EXCLUDED.payload
`, string(raw))
	return err
}

func (p *PG) SiteSettings(ctx context.Context, key string) (map[string]any, error) {
	var row *model.SiteSettings
	if err := p.db.Model("resource_site_settings").Ctx(ctx).Where("key", strings.TrimSpace(key)).Limit(1).Scan(&row); err != nil {
		return nil, err
	}
	if row == nil || strings.TrimSpace(row.Payload) == "" {
		return nil, gerror.New("resource site configuration is not seeded")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return nil, gerror.Wrap(err, "decode resource site configuration")
	}
	return payload, nil
}

func (p *PG) ReplaceSiteSettingsWithHook(
	ctx context.Context,
	key string,
	payload map[string]any,
	expectedRevision uint64,
	hook TransactionHook,
) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, execErr := tx.Ctx(ctx).Exec(`
UPDATE resource_site_settings
SET payload = ?::jsonb
WHERE key = ?
  AND COALESCE((payload ->> 'revision')::bigint, 1) = ?
`, string(raw), strings.TrimSpace(key), expectedRevision)
		if execErr != nil {
			return execErr
		}
		changed, execErr := result.RowsAffected()
		if execErr != nil {
			return execErr
		}
		if changed != 1 {
			return ErrSiteSettingsRevisionConflict
		}
		return runTransactionHook(ctx, tx, hook)
	})
}

func (p *PG) one(ctx context.Context, col, val string) (*model.Resource, error) {
	var r *model.Resource
	if err := p.db.Model(tResources).Ctx(ctx).Where(col, val).Limit(1).Scan(&r); err != nil {
		return nil, err
	}
	return r, nil
}

func (p *PG) oneOwned(ctx context.Context, owner, id string) (*model.Resource, error) {
	if strings.TrimSpace(owner) == "" {
		return p.one(ctx, "id", id)
	}
	var r *model.Resource
	if err := p.db.Model(tResources).Ctx(ctx).Where("owner_id", owner).Where("id", id).Limit(1).Scan(&r); err != nil {
		return nil, err
	}
	return r, nil
}

func isDupSlug(err error) bool {
	// lib/pq's error string is `pq: duplicate key value violates unique
	// constraint "uq_resources_slug"` — it does NOT embed the 23505 SQLSTATE, so
	// match the message text too (keep 23505 for drivers that surface the code).
	s := err.Error()
	dup := strings.Contains(s, "23505") || strings.Contains(s, "duplicate key")
	return dup && (strings.Contains(s, "slug") || strings.Contains(s, "uq_resources_slug"))
}
