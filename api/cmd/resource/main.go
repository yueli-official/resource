// Command resource is the resource site backend: catalog + browse + operator
// upload + download, layered on the asset service.
package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/yueli-official/foundation/go/siteprofile"
	"github.com/yueli-official/foundation/go/traffic"
	trafficpostgres "github.com/yueli-official/foundation/go/traffic/postgres"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/authsetup"
	"platform/gokit/observability"
	"platform/gokit/openapiexport"
	"platform/products/resource/api/internal/appconfig"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/resourcediscovery"
	"platform/products/resource/api/internal/resourceprofile"
	"platform/products/resource/api/internal/resourcetraffic"
	"platform/products/resource/api/internal/resourceurls"
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
	var urlLifecycle *resourceurls.Lifecycle
	if openapiexport.Requested() {
		urlLifecycle, err = resourceurls.NewMemory(appconfig.SiteURL(ctx))
		if err != nil {
			panic(err)
		}
	} else {
		urlLifecycle, err = resourceurls.NewPostgres(
			ctx,
			trafficDB,
			"resource:"+appconfig.SiteSlug(ctx),
			appconfig.SiteURL(ctx),
		)
		if err != nil {
			panic(err)
		}
		if err := urlLifecycle.ReconcileAll(ctx, trafficDB); err != nil {
			panic(err)
		}
	}
	cat := catalog.New(
		store,
		appconfig.BuildAssetClient(ctx),
		appconfig.LoadTypes(ctx),
		appconfig.CoverCategory(ctx),
		appconfig.SiteBrand(ctx),
	)
	cat.SetTraffic(trafficModule)
	cat.SetURLLifecycle(urlLifecycle)
	var profiles *resourceprofile.Manager
	if openapiexport.Requested() {
		profiles = resourceprofile.NewMemory()
	} else {
		profiles, err = resourceprofile.NewPostgres(trafficDB)
		if err != nil {
			panic(err)
		}
	}
	cat.SetSiteProfile(profiles)
	if !openapiexport.Requested() {
		if err := cat.EnsureSiteProfile(ctx); err != nil {
			panic(err)
		}
	}
	discoveryConfig := appconfig.DiscoveryConfig(ctx)
	discoverySnapshot := siteprofile.Snapshot{
		Profile: siteprofile.Profile{Identity: siteprofile.Identity{
			Name: discoveryConfig.Name, Description: discoveryConfig.Description,
		}},
	}
	if !openapiexport.Requested() {
		settings, err := cat.AdminSiteSettings(ctx)
		if err != nil {
			panic(err)
		}
		discoverySnapshot = settings.Snapshot
	}
	discoveryManager, err := resourcediscovery.NewManager(store, discoveryConfig, discoverySnapshot)
	if err != nil {
		panic(err)
	}
	cat.ObserveSiteProfile(discoveryManager)

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
	server.Configure(s, server.Deps{
		Verifier: verifier, Catalog: cat,
		Discovery:   discoveryManager,
		URLResolver: urlLifecycle.Resolver(),
	})
	if handled, err := openapiexport.ExportIfRequested(s); handled {
		if err != nil {
			panic(err)
		}
		return
	}
	g.Log().Info(ctx, "resource-service starting")
	s.Run()
}
