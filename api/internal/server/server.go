// Package server wires the resource-site HTTP routes onto a GoFrame server.
// Shared by cmd/resource and integration tests so they exercise the same wiring.
package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"platform/gokit/authjwt"
	"platform/gokit/ghttpx"
	"platform/gokit/healthcheck"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/controller"
)

// Deps are the wiring dependencies. Catalog may be nil for a minimal health-only
// server.
type Deps struct {
	Verifier *authjwt.Verifier
	Catalog  *catalog.Service
}

// Configure mounts: public health, public browse/download (optional auth in the
// handlers), and the JWT-protected operator API.
func Configure(s *ghttp.Server, d Deps) {
	s.Use(ghttpx.TraceRouteMiddleware)
	s.Group("/", func(grp *ghttp.RouterGroup) {
		grp.Middleware(ghttpx.Middleware)
		grp.GET("/healthz", controller.Healthz)
		grp.GET("/readyz", healthcheck.Handler(map[string]healthcheck.Check{"database": healthcheck.Database}))
	})

	if d.Catalog == nil {
		return
	}

	// Public browse/download: enveloped, no mandatory auth (handlers verify the
	// token themselves when present).
	s.Group("/", func(grp *ghttp.RouterGroup) {
		grp.Middleware(ghttpx.Middleware)
		grp.Bind(controller.NewPublicResources(d.Catalog, d.Verifier))
	})

	// Operator API: envelope first, then mandatory JWT.
	s.Group("/", func(grp *ghttp.RouterGroup) {
		grp.Middleware(ghttpx.Middleware, authjwt.Middleware(d.Verifier))
		grp.Bind(controller.Ping{})
		grp.Bind(controller.NewResources(d.Catalog))
		grp.Bind(controller.NewAssets(d.Catalog))
		grp.Bind(controller.NewTaxonomy(d.Catalog))
		grp.Bind(controller.NewSEO(d.Catalog))
	})
}
