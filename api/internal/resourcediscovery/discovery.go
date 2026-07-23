package resourcediscovery

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yueli-official/foundation/go/discovery"
	"github.com/yueli-official/foundation/go/siteprofile"

	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/model"
)

type Config struct {
	Origin      string
	Name        string
	Description string
	Locale      string
	TTL         time.Duration
	Clock       func() time.Time
}

type Manager struct {
	mu       sync.RWMutex
	store    *dao.PG
	config   Config
	module   *discovery.Module
	cache    *discovery.Cache
	revision siteprofile.Revision
	name     string
}

func NewManager(store *dao.PG, config Config, snapshot siteprofile.Snapshot) (*Manager, error) {
	manager := &Manager{store: store, config: config}
	if err := manager.Refresh(snapshot); err != nil {
		return nil, err
	}
	return manager, nil
}

func (manager *Manager) Refresh(snapshot siteprofile.Snapshot) error {
	manager.mu.RLock()
	if manager.module != nil && manager.revision == snapshot.Revision {
		manager.mu.RUnlock()
		return nil
	}
	manager.mu.RUnlock()
	config := manager.config
	config.Name = snapshot.Profile.Identity.Name
	config.Description = firstText(
		snapshot.Profile.Identity.Description,
		snapshot.Profile.Identity.Tagline,
	)
	module, cache, err := New(manager.store, config)
	if err != nil {
		return err
	}
	manager.mu.Lock()
	manager.module = module
	manager.cache = cache
	manager.revision = snapshot.Revision
	manager.name = config.Name
	manager.mu.Unlock()
	return nil
}

func (manager *Manager) ProjectResource(
	resource *model.Resource,
	seo *model.SEO,
) (discovery.PageProjection, error) {
	manager.mu.RLock()
	module := manager.module
	name := manager.name
	manager.mu.RUnlock()
	if module == nil {
		return discovery.PageProjection{}, fmt.Errorf("resource Discovery module is not initialized")
	}
	return ProjectResource(module, resource, seo, name)
}

func (manager *Manager) Snapshot(ctx context.Context) (discovery.MemoryPublication, error) {
	manager.mu.RLock()
	cache := manager.cache
	manager.mu.RUnlock()
	if cache == nil {
		return discovery.MemoryPublication{}, fmt.Errorf("resource Discovery cache is not initialized")
	}
	return cache.Snapshot(ctx)
}

func (manager *Manager) SiteName() string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.name
}

func (manager *Manager) ProfileRevision() siteprofile.Revision {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.revision
}

func firstText(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func New(store *dao.PG, config Config) (*discovery.Module, *discovery.Cache, error) {
	module, err := discovery.Compile(discovery.Definition{
		ContractVersion: discovery.ContractVersion,
		Site: discovery.SiteProfile{
			Origin: config.Origin, Name: config.Name,
			Description: config.Description, DefaultLocale: config.Locale,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	cache, err := discovery.NewCache(module, discovery.CacheOptions{
		TTL: config.TTL, Clock: config.Clock,
		Build: func(context.Context) (discovery.PublicationPlan, discovery.Sources, error) {
			return discovery.PublicationPlan{
					Sitemap: &discovery.SitemapPlan{Source: "pages"},
					Robots:  &discovery.RobotsPlan{},
				}, discovery.Sources{
					"pages": &source{
						module: module, store: store, origin: config.Origin,
						name: config.Name, description: config.Description,
						locale: config.Locale,
					},
				}, nil
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return module, cache, nil
}

func ProjectResource(
	module *discovery.Module,
	resource *model.Resource,
	seo *model.SEO,
	brand string,
) (discovery.PageProjection, error) {
	if module == nil || resource == nil || resource.Status != model.StatusPublished {
		return discovery.PageProjection{}, fmt.Errorf("published resource and Discovery module are required")
	}
	title, description, imageURL := resource.Title, resource.Summary, resource.CoverURL
	pagePath, robots := "/resources/"+resource.ID, ""
	if seo != nil {
		if seo.MetaTitle != "" {
			title = seo.MetaTitle
		}
		if seo.MetaDesc != "" {
			description = seo.MetaDesc
		}
		if seo.OgImage != "" {
			imageURL = seo.OgImage
		}
		if seo.CanonicalURL != "" {
			pagePath = seo.CanonicalURL
		}
		robots = seo.Robots
	}
	if description == "" {
		description = excerpt(resource.Description, 300)
	}
	var image *discovery.Image
	if imageURL != "" {
		image = &discovery.Image{URL: imageURL, Alt: title}
	}
	var modifiedAt *time.Time
	if resource.UpdatedAt != nil {
		value := resource.UpdatedAt.Time.UTC()
		modifiedAt = &value
	}
	visibility, follow := robotsPolicy(robots)
	projection, _, err := module.Project(discovery.PageDescriptor{
		Key: "resource:" + resource.ID, Path: pagePath,
		Locale: "zh-CN", Visibility: visibility, Follow: follow,
		ModifiedAt: modifiedAt,
		Subject: discovery.ProductSubject(discovery.Product{
			Name: title, Description: description, Image: image,
			SKU: resource.ID, Brand: brand,
		}),
	})
	return projection, err
}

type source struct {
	module      *discovery.Module
	store       *dao.PG
	origin      string
	name        string
	description string
	locale      string
}

func (value *source) Next(ctx context.Context, cursor discovery.Cursor, limit int) (discovery.Batch, error) {
	after := string(cursor)
	rowsLimit := limit + 1
	records := make([]discovery.Record, 0, limit)
	if after == "" {
		page := discovery.PageDescriptor{
			Key: "site:home", Path: "/", Locale: value.locale,
			Subject: discovery.WebPageSubject(discovery.WebPage{
				Title: value.name, Description: value.description,
			}),
		}
		projection, _, err := value.module.Project(page)
		if err != nil {
			return discovery.Batch{}, err
		}
		records = append(records, discovery.Record{SortKey: projection.CanonicalURL, Page: page})
		after = projection.CanonicalURL
		rowsLimit = limit
	}
	rows, err := value.store.ListDiscoveryPages(ctx, value.origin, after, rowsLimit)
	if err != nil {
		return discovery.Batch{}, err
	}
	hasMore := len(rows) >= rowsLimit
	if hasMore {
		rows = rows[:rowsLimit-1]
	}
	for _, row := range rows {
		record, err := value.record(row)
		if err != nil {
			return discovery.Batch{}, err
		}
		records = append(records, record)
		if len(records) == limit {
			hasMore = true
			break
		}
	}
	var next discovery.Cursor
	if len(records) > 0 {
		next = discovery.Cursor(records[len(records)-1].SortKey)
	}
	return discovery.Batch{Records: records, NextCursor: next, Done: !hasMore}, nil
}

func (value *source) record(row dao.DiscoveryRow) (discovery.Record, error) {
	pagePath := row.Path
	if row.CanonicalURL != "" {
		pagePath = row.CanonicalURL
	}
	visibility, follow := robotsPolicy(row.Robots)
	var modifiedAt *time.Time
	if row.UpdatedAt != nil {
		updated := row.UpdatedAt.Time.UTC()
		modifiedAt = &updated
	}
	var image *discovery.Image
	if row.ImageURL != "" {
		image = &discovery.Image{URL: row.ImageURL, Alt: row.Title}
	}
	var subject discovery.Subject
	if row.Kind == "resource" {
		subject = discovery.ProductSubject(discovery.Product{
			Name: row.Title, Description: row.Description,
			Image: image, SKU: strings.TrimPrefix(row.Key, "resource:"),
			Brand: value.name,
		})
	} else {
		subject = discovery.CollectionSubject(discovery.Collection{
			Title: row.Title, Description: row.Description,
		})
	}
	page := discovery.PageDescriptor{
		Key: row.Key, Path: pagePath, Locale: value.locale,
		Visibility: visibility, Follow: follow, ModifiedAt: modifiedAt,
		Subject: subject,
	}
	projection, _, err := value.module.Project(page)
	if err != nil {
		return discovery.Record{}, err
	}
	return discovery.Record{SortKey: projection.CanonicalURL, Page: page}, nil
}

func robotsPolicy(value string) (discovery.Visibility, discovery.FollowPolicy) {
	lower := strings.ToLower(value)
	visibility := discovery.Discoverable
	if strings.Contains(lower, "noindex") {
		visibility = discovery.Unlisted
	}
	follow := discovery.Follow
	if strings.Contains(lower, "nofollow") {
		follow = discovery.NoFollow
	}
	return visibility, follow
}

func excerpt(value string, max int) string {
	runes := []rune(strings.Join(strings.Fields(value), " "))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max]) + "…"
}
