package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	_ "github.com/lib/pq"

	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/resourceprofile"
)

func TestSiteProfileLegacyCutoverAndDowngradeProjection(t *testing.T) {
	host := os.Getenv("RESOURCE_SITE_PROFILE_PG_HOST")
	database := os.Getenv("RESOURCE_SITE_PROFILE_PG_DATABASE")
	if host == "" || database == "" {
		t.Skip("set RESOURCE_SITE_PROFILE_PG_HOST and RESOURCE_SITE_PROFILE_PG_DATABASE")
	}
	port := envOr("RESOURCE_SITE_PROFILE_PG_PORT", "5432")
	user := envOr("RESOURCE_SITE_PROFILE_PG_USER", "postgres")
	password := os.Getenv("RESOURCE_SITE_PROFILE_PG_PASS")
	gfdb, err := gdb.New(gdb.ConfigNode{
		Type: "pgsql", Host: host, Port: port, User: user, Pass: password, Name: database,
	})
	if err != nil {
		t.Fatal(err)
	}
	dsn := (&url.URL{
		Scheme: "postgres", User: url.UserPassword(user, password),
		Host: fmt.Sprintf("%s:%s", host, port), Path: database,
		RawQuery: "sslmode=disable",
	}).String()
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })

	legacy := map[string]any{
		"site": map[string]any{
			"siteName": "资源库", "tagline": "下载资源", "logoIcon": "i-tabler-package",
			"announcement": "公告", "announcementEnabled": true, "supportEmail": "support@example.com",
		},
		"footer": map[string]any{
			"tagline": "页脚", "copyright": "Copyright",
			"compliance": map[string]any{
				"icpRecord": "ICP 1", "icpUrl": "https://beian.miit.gov.cn/",
				"policeRecord": "", "policeUrl": "", "extraText": "",
			},
			"linkGroups": []any{map[string]any{
				"title": "关于", "links": []any{map[string]any{"label": "介绍", "to": "/about", "icon": "i-tabler-link"}},
			}},
			"socialLinks": []any{map[string]any{"label": "GitHub", "to": "https://github.com/example", "icon": "i-tabler-brand-github"}},
		},
		"resource": map[string]any{
			"resourcesPerPage": 20, "downloadsEnabled": true,
			"largeFileThresholdMB": 100, "largeFileHint": "提示",
		},
	}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqldb.ExecContext(context.Background(), `
INSERT INTO resource_site_settings (key, payload) VALUES ('site', $1::jsonb)
ON CONFLICT (key) DO UPDATE SET payload = EXCLUDED.payload
`, string(raw)); err != nil {
		t.Fatal(err)
	}
	if _, err := sqldb.ExecContext(context.Background(), `DELETE FROM site_profile_state`); err != nil {
		t.Fatal(err)
	}
	manager, err := resourceprofile.NewPostgres(sqldb)
	if err != nil {
		t.Fatal(err)
	}
	service := New(dao.NewPG(gfdb), nil, nil, "", "资源库")
	service.SetSiteProfile(manager)
	if err := service.EnsureSiteProfile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.EnsureSiteProfile(context.Background()); err != nil {
		t.Fatalf("idempotent EnsureSiteProfile: %v", err)
	}
	settings, err := service.SiteSettings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if settings.Revision != 1 || settings.Site.SiteName != "资源库" || settings.Resource.ResourcesPerPage != 20 {
		t.Fatalf("settings = %#v", settings)
	}
	admin, err := service.AdminSiteSettings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	updatedRuntime := admin.Resource
	updatedRuntime.ResourcesPerPage++
	if _, err := service.SaveAdminSiteSettings(
		context.Background(), admin.Snapshot.Revision, admin.RuntimeRevision,
		admin.Snapshot.Profile, updatedRuntime,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveAdminSiteSettings(
		context.Background(), admin.Snapshot.Revision, admin.RuntimeRevision,
		admin.Snapshot.Profile, updatedRuntime,
	); err == nil {
		t.Fatal("stale resource runtime revision did not conflict")
	}
	var hasSite, hasFooter, hasResource bool
	if err := sqldb.QueryRowContext(context.Background(), `
SELECT payload ? 'site', payload ? 'footer', payload ? 'resource'
FROM resource_site_settings WHERE key = 'site'
`).Scan(&hasSite, &hasFooter, &hasResource); err != nil {
		t.Fatal(err)
	}
	if hasSite || hasFooter || !hasResource {
		t.Fatalf("narrowed payload site=%v footer=%v resource=%v", hasSite, hasFooter, hasResource)
	}

	downPath := filepath.Join("..", "..", "manifest", "sql", "migrations", "0010_site_profile_cutover.down.sql")
	downSQL, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := sqldb.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(context.Background(), string(downSQL)); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRowContext(context.Background(), `
SELECT payload ? 'site', payload ? 'footer' FROM resource_site_settings WHERE key = 'site'
`).Scan(&hasSite, &hasFooter); err != nil {
		t.Fatal(err)
	}
	if !hasSite || !hasFooter {
		t.Fatalf("downgrade projection site=%v footer=%v", hasSite, hasFooter)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
