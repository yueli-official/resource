// Command resource is the resource site backend: catalog + browse + operator
// upload + download, layered on the asset service.
package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/authjwt"
	"platform/products/resource/api/internal/appconfig"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/server"
)

func main() {
	ctx := gctx.New()

	// ── Catalog logic (DB + asset-service client) ────────────────────────────
	cat := catalog.New(
		dao.NewPG(g.DB()),
		appconfig.BuildAssetClient(ctx),
		appconfig.LoadTypes(ctx),
		appconfig.CoverCategory(ctx),
		appconfig.SiteBrand(ctx),
	)

	// ── JWT verifier (IdP JWKS, lazy) ────────────────────────────────────────
	jw := appconfig.LoadJWKS(ctx)
	verifier, err := authjwt.NewVerifier(authjwt.VerifierConfig{
		Keys: authjwt.NewRemoteKeySource(jw.URL), Issuer: jw.Issuer, Audience: jw.Audience,
	})
	if err != nil {
		panic(err)
	}

	s := g.Server()
	server.Configure(s, server.Deps{Verifier: verifier, Catalog: cat})
	g.Log().Info(ctx, "resource-service starting")
	s.Run()
}
