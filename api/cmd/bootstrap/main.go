package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/yueli-official/foundation/go/siteprofile"

	"github.com/yueli-official/resource/api/internal/resourceprofile"
)

const initialHomeConfig = `
INSERT INTO resource_home_settings (id, payload)
VALUES (
	1,
	'{"heroTitle":"资源","heroSubtitle":"浏览并下载软件、设计素材与脚本。","introTitle":"面向创作和开发的资源目录","introBody":"这里集中展示可下载工具、设计素材、脚本和模板。","introHighlights":[],"featuredResourceIds":[],"categorySlugs":[],"quickLinks":[]}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO resource_site_settings (key, payload)
VALUES (
	'site',
	'{"revision":1,"resource":{"resourcesPerPage":12,"downloadsEnabled":true,"largeFileThresholdMB":500,"largeFileHint":"大文件建议使用 OSS/COS/S3 分片上传或网盘。"}}'::jsonb
)
ON CONFLICT (key) DO NOTHING`

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("RESOURCE_DATABASE_URL"))
	if databaseURL == "" {
		fail("RESOURCE_DATABASE_URL is required")
	}
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		fail("open database: %v", err)
	}
	defer database.Close()
	ctx := context.Background()
	if err := database.PingContext(ctx); err != nil {
		fail("connect database: %v", err)
	}
	if _, err := database.ExecContext(ctx, initialHomeConfig); err != nil {
		fail("install initial Resource configuration: %v", err)
	}
	profiles, err := resourceprofile.NewPostgres(database)
	if err != nil {
		fail("open Resource site profile: %v", err)
	}
	if _, err := profiles.Get(ctx); errors.Is(err, siteprofile.ErrNotInitialized) {
		if _, err := profiles.Replace(ctx, siteprofile.ReplaceCommand{
			Profile: resourceprofile.InitialProfile(
				os.Getenv("RESOURCE_SITE_BRAND"),
				os.Getenv("RESOURCE_SITE_DESCRIPTION"),
			),
		}); err != nil {
			fail("install initial Resource site profile: %v", err)
		}
	} else if err != nil {
		fail("read Resource site profile: %v", err)
	}
	fmt.Println("Resource initial records are ready")
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "bootstrap: "+format+"\n", arguments...)
	os.Exit(1)
}
