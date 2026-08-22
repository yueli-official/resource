// Package appconfig builds runtime objects from the GoFrame config
// (manifest/config/config.yaml + GF_* env overrides).
package appconfig

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	_ "github.com/lib/pq"

	"github.com/yueli-official/resource/api/internal/assetclient"
	"github.com/yueli-official/resource/api/internal/catalog"
	"github.com/yueli-official/resource/api/internal/resourcediscovery"
)

// OpenTrafficDB opens the standard-library PostgreSQL handle required by the
// Foundation Traffic Adapter against this consumer-owned database.
func OpenTrafficDB(ctx context.Context) (*sql.DB, error) {
	host := g.Cfg().MustGet(ctx, "database.default.host").String()
	port := g.Cfg().MustGet(ctx, "database.default.port", "5432").String()
	name := g.Cfg().MustGet(ctx, "database.default.name").String()
	user := g.Cfg().MustGet(ctx, "database.default.user").String()
	password := g.Cfg().MustGet(ctx, "database.default.pass").String()
	if host == "" || name == "" || user == "" {
		return nil, fmt.Errorf("database.default host, name, and user are required")
	}
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   name,
	}
	query := dsn.Query()
	query.Set("sslmode", g.Cfg().MustGet(ctx, "database.default.sslmode", "disable").String())
	dsn.RawQuery = query.Encode()
	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("open traffic database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping traffic database: %w", err)
	}
	return db, nil
}

// TrafficTimeZone is immutable after this siteSlug's traffic initialization.
func TrafficTimeZone(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.traffic.timeZone", "Asia/Shanghai").String()
}

// BuildAssetClient constructs the HTTP asset-service client from config. The
// resource site is fully free → only the base URL is needed (no service token).
func BuildAssetClient(ctx context.Context) assetclient.Client {
	return assetclient.NewHTTP(
		g.Cfg().MustGet(ctx, "resource.assetService.baseUrl").String(),
		AssetNamespace(ctx),
		AssetSpace(ctx),
	)
}

func SiteSlug(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.siteSlug", "resource").String()
}

func BootstrapAdministratorSubs(ctx context.Context) []string {
	return g.Cfg().MustGet(ctx, "resource.authorization.bootstrapAdministratorSubs").Strings()
}

func SiteURL(ctx context.Context) string {
	return strings.TrimRight(g.Cfg().MustGet(ctx, "resource.siteUrl", "http://localhost:3001").String(), "/")
}

func SiteBrand(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.brand", "资源库").String()
}

func DiscoveryConfig(ctx context.Context) resourcediscovery.Config {
	return resourcediscovery.Config{
		Origin:      SiteURL(ctx),
		Name:        SiteBrand(ctx),
		Description: g.Cfg().MustGet(ctx, "resource.siteDescription", "免费资源、工具与素材").String(),
		Locale:      g.Cfg().MustGet(ctx, "resource.locale", "zh-CN").String(),
		TTL:         g.Cfg().MustGet(ctx, "resource.discovery.ttl", 5*time.Minute).Duration(),
		Clock:       time.Now,
	}
}

func AssetSpace(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.assetSpace", "default").String()
}

func AssetNamespace(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "resource.assetNamespace", "resource").String()
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
	URL               string
	Issuer            string
	Audience          string
	AllowLoopbackHTTP bool
}

func LoadJWKS(ctx context.Context) JWKS {
	return JWKS{
		URL:               g.Cfg().MustGet(ctx, "resource.jwks.url").String(),
		Issuer:            g.Cfg().MustGet(ctx, "resource.jwks.issuer").String(),
		Audience:          g.Cfg().MustGet(ctx, "resource.jwks.audience").String(),
		AllowLoopbackHTTP: g.Cfg().MustGet(ctx, "resource.jwks.allowLoopbackHttp", false).Bool(),
	}
}
