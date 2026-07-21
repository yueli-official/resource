// Package appconfig builds runtime objects from the GoFrame config
// (manifest/config/config.yaml + GF_* env overrides).
package appconfig

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"platform/products/resource/api/internal/assetclient"
	"platform/products/resource/api/internal/catalog"
)

// BuildAssetClient constructs the HTTP asset-service client from config. The
// resource site is fully free → only the base URL is needed (no service token).
func BuildAssetClient(ctx context.Context) assetclient.Client {
	return assetclient.NewHTTP(
		g.Cfg().MustGet(ctx, "resource.assetService.baseUrl").String(),
		SiteSlug(ctx),
		AssetSpace(ctx),
	)
}

func SiteSlug(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.siteSlug", "resource").String()
}

func SiteBrand(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.brand", "资源库").String()
}

func AssetSpace(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.assetSpace", "default").String()
}

// LoadTypes reads resource.types into the catalog's type rules.
func LoadTypes(ctx context.Context) map[string]catalog.TypeRule {
	out := map[string]catalog.TypeRule{}
	for name, raw := range g.Cfg().MustGet(ctx, "resource.types").Map() {
		m := g.NewVar(raw).Map()
		exts := map[string]bool{}
		for _, e := range g.NewVar(m["allowedExt"]).Strings() {
			exts[strings.ToLower(e)] = true
		}
		out[name] = catalog.TypeRule{
			Label:      g.NewVar(m["label"]).String(),
			AllowedExt: exts,
			MaxSizeMB:  g.NewVar(m["maxSizeMb"]).Int(),
		}
	}
	return out
}

// CoverCategory returns the asset-service category for cover images.
func CoverCategory(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.coverCategory", "resource-cover").String()
}

// JWKS is the IdP key/issuer config for the Foundation auth verifier.
type JWKS struct {
	URL      string
	Issuer   string
	Audience string
}

func LoadJWKS(ctx context.Context) JWKS {
	return JWKS{
		URL:      g.Cfg().MustGet(ctx, "resource.jwks.url").String(),
		Issuer:   g.Cfg().MustGet(ctx, "resource.jwks.issuer").String(),
		Audience: g.Cfg().MustGet(ctx, "resource.jwks.audience").String(),
	}
}
