// Command resource is the resource site backend: catalog + browse + operator
// upload + download, layered on the asset service.
package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/yueli-official/foundation/go/traffic"
	trafficpostgres "github.com/yueli-official/foundation/go/traffic/postgres"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/authsetup"
	"platform/gokit/observability"
	"platform/gokit/openapiexport"
	"platform/products/resource/api/internal/appconfig"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/resourcetraffic"
	"platform/products/resource/api/internal/server"
)

func main() {
	ctx := gctx.New()
	shutdown, err := observability.StartFromEnvironment(ctx, "resource-api")
	if err != nil {
		panic(err)
	}
	defer observability.ShutdownWithTimeout(shutdown)

	// ── Catalog logic (DB + asset-service client) ────────────────────────────
	store := dao.NewPG(g.DB())
	legacyTraffic, err := resourcetraffic.SnapshotLegacy(ctx, store)
	if err != nil {
		panic(err)
	}
	trafficDB, err := appconfig.OpenTrafficDB(ctx)
	if err != nil {
		panic(err)
	}
	defer trafficDB.Close()
	trafficCatalog, err := traffic.Compile(resourcetraffic.Definition(appconfig.TrafficTimeZone(ctx)))
	if err != nil {
		panic(err)
	}
	trafficModule, err := trafficpostgres.New(ctx, trafficCatalog, trafficpostgres.Options{
		DB: trafficDB, InstanceKey: "resource:" + appconfig.SiteSlug(ctx),
		InitialBaselines: legacyTraffic.InitialBaselines,
	})
	if err != nil {
		panic(err)
	}
	if err := resourcetraffic.Reconcile(ctx, trafficModule, store, legacyTraffic.Resources); err != nil {
		panic(err)
	}
	cat := catalog.New(
		store,
		appconfig.BuildAssetClient(ctx),
		appconfig.LoadTypes(ctx),
		appconfig.CoverCategory(ctx),
		appconfig.SiteBrand(ctx),
	)
	cat.SetTraffic(trafficModule)

	// ── JWT verifier (IdP JWKS, lazy) ────────────────────────────────────────
	jw := appconfig.LoadJWKS(ctx)
	verifier, err := authsetup.NewRemoteVerifier(authsetup.RemoteVerifierConfig{
		JWKSURL: jw.URL, Issuer: jw.Issuer, Audience: jw.Audience,
		AllowLoopbackHTTP: jw.AllowLoopbackHTTP,
	})
	if err != nil {
		panic(err)
	}

	s := g.Server()
	server.Configure(s, server.Deps{Verifier: verifier, Catalog: cat})
	if handled, err := openapiexport.ExportIfRequested(s); handled {
		if err != nil {
			panic(err)
		}
		return
	}
	g.Log().Info(ctx, "resource-service starting")
	s.Run()
}
