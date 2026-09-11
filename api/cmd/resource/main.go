// Command resource is the resource site backend: catalog + browse + operator
// upload + download, layered on the asset service.
package main

import (
	"context"
	"github.com/yueli-official/resource/api/internal/assetreferences"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/yueli-official/foundation/go/authorization"
	authorizationpostgres "github.com/yueli-official/foundation/go/authorization/postgres"
	"github.com/yueli-official/foundation/go/siteprofile"
	"github.com/yueli-official/foundation/go/traffic"
	trafficpostgres "github.com/yueli-official/foundation/go/traffic/postgres"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"github.com/yueli-official/resource/api/internal/appconfig"
	"github.com/yueli-official/resource/api/internal/assetclient"
	"github.com/yueli-official/resource/api/internal/catalog"
	"github.com/yueli-official/resource/api/internal/dao"
	"github.com/yueli-official/resource/api/internal/resourceauthz"
	"github.com/yueli-official/resource/api/internal/resourcediscovery"
	"github.com/yueli-official/resource/api/internal/resourceprofile"
	"github.com/yueli-official/resource/api/internal/resourcesearch"
	"github.com/yueli-official/resource/api/internal/resourcetraffic"
	"github.com/yueli-official/resource/api/internal/resourceurls"
	"github.com/yueli-official/resource/api/internal/runtime"
	"github.com/yueli-official/resource/api/internal/server"
)

func main() {
	if err := runtime.EnableEnvironmentConfig(); err != nil {
		panic(err)
	}
	ctx := gctx.New()
	if runtime.OpenAPIRequested() {
		exportOpenAPI(ctx)
		return
	}
	shutdown, err := runtime.StartTelemetry(ctx, "resource-api")
	if err != nil {
		panic(err)
	}
	defer runtime.ShutdownTelemetry(shutdown)

	// ── Catalog logic (DB + asset-service client) ────────────────────────────
	store := dao.NewPG(g.DB())
	trafficDB, err := appconfig.OpenTrafficDB(ctx)
	if err != nil {
		panic(err)
	}
	defer trafficDB.Close()
	stopReferences, err := assetreferences.Start(ctx, trafficDB)
	if err != nil {
		panic(err)
	}
	defer stopReferences()
	trafficCatalog, err := traffic.Compile(resourcetraffic.Definition(appconfig.TrafficTimeZone(ctx)))
	if err != nil {
		panic(err)
	}
	trafficModule, err := trafficpostgres.New(ctx, trafficCatalog, trafficpostgres.Options{
		DB: trafficDB, InstanceKey: "resource:" + appconfig.SiteSlug(ctx),
	})
	if err != nil {
		panic(err)
	}
	if err := resourcetraffic.ReconcileProjections(ctx, trafficModule, store); err != nil {
		panic(err)
	}
	urlLifecycle, err := resourceurls.NewPostgres(
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
	cat := catalog.New(
		store,
		appconfig.BuildAssetClient(ctx),
		appconfig.LoadTypes(ctx),
		appconfig.CoverCategory(ctx),
		appconfig.SiteBrand(ctx),
	)
	cat.SetTraffic(trafficModule)
	cat.SetURLLifecycle(urlLifecycle)
	searchIndex, err := resourcesearch.NewPostgres(ctx, trafficDB, appconfig.SiteSlug(ctx))
	if err != nil {
		panic(err)
	}
	if err := searchIndex.Reconcile(ctx, trafficDB); err != nil {
		panic(err)
	}
	cat.SetSearch(searchIndex)
	profiles, err := resourceprofile.NewPostgres(trafficDB)
	if err != nil {
		panic(err)
	}
	cat.SetSiteProfile(profiles)
	discoveryConfig := appconfig.DiscoveryConfig(ctx)
	discoverySnapshot := siteprofile.Snapshot{
		Profile: siteprofile.Profile{Identity: siteprofile.Identity{
			Name: discoveryConfig.Name, Description: discoveryConfig.Description,
		}},
	}
	settings, err := cat.AdminSiteSettings(ctx)
	if err != nil {
		panic(err)
	}
	discoverySnapshot = settings.Snapshot
	discoveryManager, err := resourcediscovery.NewManager(store, discoveryConfig, discoverySnapshot)
	if err != nil {
		panic(err)
	}
	cat.ObserveSiteProfile(discoveryManager)

	definition, err := authorization.Compile(resourceauthz.Definition())
	if err != nil {
		panic(err)
	}
	bootstrapSubs := appconfig.BootstrapAdministratorSubs(ctx)
	protected := make([]authorization.SubjectRef, 0, len(bootstrapSubs))
	for _, sub := range bootstrapSubs {
		if sub != "" {
			protected = append(protected, authorization.SubjectRef{
				Kind: authorization.SubjectUser, ID: sub,
			})
		}
	}
	authz, err := authorizationpostgres.New(ctx, definition, authorizationpostgres.Options{
		DB: trafficDB, InstanceKey: "resource:" + appconfig.SiteSlug(ctx),
		Memory: authorization.MemoryOptions{
			RootScopeID: resourceauthz.RootScopeID, ProtectedSubjects: protected,
			Constraints: resourceauthz.ConstraintEvaluators(),
			Predicates:  resourceauthz.PredicateEvaluators(),
		},
	})
	if err != nil {
		panic(err)
	}
	if authz.InstanceWasCreated() {
		if len(protected) == 0 {
			panic("resource authorization bootstrap requires at least one administrator subject")
		}
		if err := resourceauthz.SyncResourceScopes(ctx, trafficDB, authz); err != nil {
			panic(err)
		}
	}
	authorizationService := resourceauthz.New(authz, trafficDB)

	// ── JWT verifier (IdP JWKS, lazy) ────────────────────────────────────────
	jw := appconfig.LoadJWKS(ctx)
	verifier, err := runtime.NewRemoteVerifier(runtime.RemoteVerifierConfig{
		JWKSURL: jw.URL, Issuer: jw.Issuer, Audience: jw.Audience,
		AllowLoopbackHTTP: jw.AllowLoopbackHTTP,
	})
	if err != nil {
		panic(err)
	}

	s := g.Server()
	server.Configure(s, server.Deps{
		Verifier: verifier, Catalog: cat,
		Discovery:     discoveryManager,
		URLResolver:   urlLifecycle.Resolver(),
		Authorization: authorizationService,
	})
	g.Log().Info(ctx, "resource-service starting")
	s.Run()
}

func exportOpenAPI(ctx context.Context) {
	urlLifecycle, err := resourceurls.NewMemory("https://resource.example.test")
	if err != nil {
		panic(err)
	}
	trafficCatalog, err := traffic.Compile(resourcetraffic.Definition("UTC"))
	if err != nil {
		panic(err)
	}
	trafficModule, err := traffic.NewMemory(trafficCatalog, traffic.MemoryOptions{
		Clock: time.Now, Secret: []byte("resource-openapi-traffic-secret-32-bytes"),
	})
	if err != nil {
		panic(err)
	}
	cat := catalog.New(
		nil,
		assetclient.NewFake(),
		map[string]catalog.TypeRule{
			"software": {
				Label:      "Software",
				AllowedExt: map[string]bool{"zip": true},
				MaxSizeMB:  500,
			},
		},
		"resource-cover",
		"Resource",
	)
	cat.SetTraffic(trafficModule)
	cat.SetURLLifecycle(urlLifecycle)
	cat.SetSearch(resourcesearch.NewMemory())
	cat.SetSiteProfile(resourceprofile.NewMemory())

	discoveryManager, err := resourcediscovery.NewManager(nil, resourcediscovery.Config{
		Origin: "https://resource.example.test", Name: "Resource",
		Description: "OpenAPI export", Locale: "en", TTL: 5 * time.Minute, Clock: time.Now,
	}, siteprofile.Snapshot{
		Profile: siteprofile.Profile{Identity: siteprofile.Identity{
			Name: "Resource", Description: "OpenAPI export",
		}},
	})
	if err != nil {
		panic(err)
	}
	cat.ObserveSiteProfile(discoveryManager)

	definition, err := authorization.Compile(resourceauthz.Definition())
	if err != nil {
		panic(err)
	}
	authz, err := authorization.NewMemory(definition, authorization.MemoryOptions{
		RootScopeID: resourceauthz.RootScopeID,
		ProtectedSubjects: []authorization.SubjectRef{{
			Kind: authorization.SubjectUser, ID: "openapi-export-admin",
		}},
		Constraints: resourceauthz.ConstraintEvaluators(),
		Predicates:  resourceauthz.PredicateEvaluators(),
	})
	if err != nil {
		panic(err)
	}
	authorizationService := resourceauthz.New(authz, nil)

	s := g.Server()
	server.Configure(s, server.Deps{
		Catalog: cat, Discovery: discoveryManager,
		URLResolver: urlLifecycle.Resolver(), Authorization: authorizationService,
	})
	handled, err := runtime.ExportOpenAPIIfRequested(s)
	if err != nil {
		panic(err)
	}
	if !handled {
		panic("RESOURCE_OPENAPI_OUTPUT is required")
	}
}
