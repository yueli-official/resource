// Package server wires the resource-site HTTP routes onto a GoFrame server.
// Shared by cmd/resource and integration tests so they exercise the same wiring.
package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/urllifecycle"
	"github.com/yueli-official/resource/api/internal/catalog"
	"github.com/yueli-official/resource/api/internal/controller"
	"github.com/yueli-official/resource/api/internal/resourceauthz"
	"github.com/yueli-official/resource/api/internal/resourcediscovery"
	"github.com/yueli-official/resource/api/internal/runtime"
)

// Deps are the wiring dependencies. Catalog may be nil for a minimal health-only
// server.
type Deps struct {
	Verifier      *foundationauth.Verifier
	Catalog       *catalog.Service
	Discovery     *resourcediscovery.Manager
	URLResolver   urllifecycle.Resolver
	Authorization *resourceauthz.Service
}

// Configure mounts: public health, public browse/download (optional auth in the
// handlers), and the JWT-protected operator API.
func Configure(s *ghttp.Server, d Deps) {
	apiMiddleware := runtime.MustAPIMiddleware(runtime.MustRateLimiterFromEnvironment())
	s.Use(runtime.TraceRouteMiddleware)
	s.Group("/", func(grp *ghttp.RouterGroup) {
		grp.Middleware(apiMiddleware.Handle, controller.CauseMappingMiddleware)
		grp.GET("/healthz", controller.Healthz)
		grp.GET("/readyz", runtime.ReadinessHandler(map[string]runtime.ReadinessCheck{"database": runtime.DatabaseReadiness}))
	})

	if d.Catalog == nil {
		return
	}

	// Public browse/download: enveloped, no mandatory auth (handlers verify the
	// token themselves when present).
	s.Group("/", func(grp *ghttp.RouterGroup) {
		grp.Middleware(apiMiddleware.Handle, controller.CauseMappingMiddleware, controller.AuthorizationMiddleware(d.Authorization))
		grp.Bind(controller.NewPublicResources(d.Catalog, d.Verifier, d.Discovery))
		if d.Discovery != nil {
			grp.Bind(controller.NewPublicDiscovery(d.Discovery))
		}
		if d.URLResolver != nil {
			grp.Bind(controller.NewPublicURLLifecycle(d.URLResolver))
		}
	})

	// Operator API: envelope first, then mandatory JWT.
	s.Group("/", func(grp *ghttp.RouterGroup) {
		if d.Verifier != nil {
			grp.Middleware(
				apiMiddleware.Handle,
				controller.CauseMappingMiddleware,
				runtime.RequiredAuth(d.Verifier),
				controller.AuthorizationMiddleware(d.Authorization),
			)
		} else {
			// OpenAPI export has no runtime verifier, but protected route shapes
			// still belong in the generated contract.
			grp.Middleware(apiMiddleware.Handle, controller.CauseMappingMiddleware, controller.AuthorizationMiddleware(d.Authorization))
		}
		grp.Bind(controller.Ping{})
		grp.Bind(controller.NewResources(d.Catalog))
		grp.Bind(controller.NewAssets(d.Catalog))
		grp.Bind(controller.NewTaxonomy(d.Catalog))
		grp.Bind(controller.NewSEO(d.Catalog))
		grp.Bind(controller.NewAuthorization())
	})
}
