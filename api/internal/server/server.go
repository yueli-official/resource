// Package server wires the resource-site HTTP routes onto a GoFrame server.
// Shared by cmd/resource and integration tests so they exercise the same wiring.
package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/urllifecycle"
	"platform/gokit/authhttp"
	"platform/gokit/ghttpx"
	"platform/gokit/healthcheck"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/controller"
	"platform/products/resource/api/internal/resourcediscovery"
)

// Deps are the wiring dependencies. Catalog may be nil for a minimal health-only
// server.
type Deps struct {
	Verifier    *foundationauth.Verifier
	Catalog     *catalog.Service
	Discovery   *resourcediscovery.Manager
	URLResolver urllifecycle.Resolver
}

// Configure mounts: public health, public browse/download (optional auth in the
// handlers), and the JWT-protected operator API.
func Configure(s *ghttp.Server, d Deps) {
	apiMiddleware := ghttpx.NewMiddleware(ghttpx.MustRateLimiterFromEnvironment(), ghttpx.ForwardedClientIPKey)
	s.Use(ghttpx.TraceRouteMiddleware)
	s.Group("/", func(grp *ghttp.RouterGroup) {
		grp.Middleware(apiMiddleware)
		grp.GET("/healthz", controller.Healthz)
		grp.GET("/readyz", healthcheck.Handler(map[string]healthcheck.Check{"database": healthcheck.Database}))
	})

	if d.Catalog == nil {
		return
	}

	// Public browse/download: enveloped, no mandatory auth (handlers verify the
	// token themselves when present).
	s.Group("/", func(grp *ghttp.RouterGroup) {
		grp.Middleware(apiMiddleware)
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
		grp.Middleware(apiMiddleware, authhttp.Required(d.Verifier))
		grp.Bind(controller.Ping{})
		grp.Bind(controller.NewResources(d.Catalog))
		grp.Bind(controller.NewAssets(d.Catalog))
		grp.Bind(controller.NewTaxonomy(d.Catalog))
		grp.Bind(controller.NewSEO(d.Catalog))
	})
}
