package assetreferences

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/asset/referencesync"
	"os"
)

func Start(ctx context.Context, db *sql.DB) (context.CancelFunc, error) {
	cfg := referencesync.Config{
		BaseURL:      g.Cfg().MustGet(ctx, "resource.assetService.baseUrl").String(),
		TokenURL:     g.Cfg().MustGet(ctx, "resource.assetService.tokenUrl").String(),
		ClientID:     g.Cfg().MustGet(ctx, "resource.assetService.clientId", "resource-asset-svc").String(),
		ClientSecret: g.Cfg().MustGet(ctx, "resource.assetService.clientSecret").String(),
	}.WithEnvironment()
	return referencesync.Start(ctx, db, "resource:asset-references", Source(g.Cfg().MustGet(ctx, "resource.siteUrl").String(), os.Getenv("ASSET_PUBLIC_ORIGIN")), cfg, func(err error) { g.Log().Warning(ctx, "asset reference reconciliation:", err) })
}
