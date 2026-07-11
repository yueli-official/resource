package server_test

// Full HTTP path against a loopback server + a fake asset client + live PG
// (database `resource` on the same instance as asset). Skipped unless
// RESOURCE_PG_HOST is set:
//
//	RESOURCE_PG_HOST=192.168.5.5 RESOURCE_PG_USER=postgres RESOURCE_PG_PASS=postgres \
//	  go test -p 1 -run TestResourceHTTPRoundTrip ./products/resource/api/internal/server/...
//
// The site is fully free: every resource is public and every file is delivered
// from the asset service's public CDN URL (no access gate / commerce / signing).

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/test/gtest"
	_ "github.com/lib/pq"

	"platform/products/resource/api/internal/assetclient"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/server"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func TestResourceHTTPRoundTrip(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		host := os.Getenv("RESOURCE_PG_HOST")
		if host == "" {
			t.Skip("set RESOURCE_PG_HOST to run the resource HTTP integration test")
		}
		port, user, pass := envOr("RESOURCE_PG_PORT", "5432"), envOr("RESOURCE_PG_USER", "postgres"), os.Getenv("RESOURCE_PG_PASS")
		ctx := context.Background()

		sdb, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=resource sslmode=disable", host, port, user, pass))
		t.AssertNil(err)
		// Rebuild the schema from the full migration chain so the test exercises
		// the current shape.
		_, err = sdb.Exec("DROP TABLE IF EXISTS resource_seo, object_taxonomies, taxonomies, terms, resource_assets, resources CASCADE")
		t.AssertNil(err)
		for _, f := range []string{"0001_init.up.sql", "0002_cover_url.up.sql", "0003_free_taxonomy_seo.up.sql", "0004_site_settings.up.sql", "0005_delivery_payload.up.sql"} {
			up, err := os.ReadFile("../../manifest/sql/migrations/" + f)
			t.AssertNil(err)
			_, err = sdb.Exec(string(up))
			t.AssertNil(err)
		}
		sdb.Close()

		db, err := gdb.New(gdb.ConfigNode{Type: "pgsql", Host: host, Port: port, User: user, Pass: pass, Name: "resource"})
		t.AssertNil(err)

		fake := assetclient.NewFake()
		types := map[string]catalog.TypeRule{
			"software": {Label: "软件", AllowedExt: map[string]bool{"zip": true}, MaxSizeMB: 500},
			"default":  {Label: "其它", AllowedExt: map[string]bool{"zip": true}, MaxSizeMB: 200},
		}
		cat := catalog.New(dao.NewPG(db), fake, types, "resource-cover", "Resource Test")

		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		t.AssertNil(err)
		s := g.Server(t.Name())
		s.SetAddr("127.0.0.1:0")
		server.Configure(s, server.Deps{Verifier: mustVerifier(t, priv), Catalog: cat})
		s.SetDumpRouterMap(false)
		s.Start()
		defer s.Shutdown()
		base := prefix(s)
		jwt := signToken(t, priv, testSub, time.Now().UTC().Add(10*time.Minute))

		op := func() *gclient.Client {
			c := g.Client()
			c.SetPrefix(base)
			c.ContentJson()
			c.SetHeader("Authorization", "Bearer "+jwt)
			return c
		}
		anon := func() *gclient.Client { c := g.Client(); c.SetPrefix(base); return c }

		// helper: full create→add→finalize→publish for a downloadable resource
		makePublished := func(title string) (id, assetID string) {
			rc, err := op().Post(ctx, "/api/v1/resources", g.Map{"title": title, "type": "software", "tags": []string{"cli", "tool"}})
			t.AssertNil(err)
			defer rc.Close()
			t.Assert(rc.StatusCode, 200)
			id = gjson.New(rc.ReadAllString()).Get("data.resource.id").String()
			t.AssertNE(id, "")

			ra, err := op().Post(ctx, "/api/v1/resources/"+id+"/assets", g.Map{"filename": "a.zip", "size": 1000})
			t.AssertNil(err)
			defer ra.Close()
			t.Assert(ra.StatusCode, 200)
			tok := gjson.New(ra.ReadAllString()).Get("data.uploadToken").String()
			t.AssertNE(tok, "")

			rf, err := op().Post(ctx, "/api/v1/resources/"+id+"/assets/finalize", g.Map{"uploadToken": tok, "label": "main"})
			t.AssertNil(err)
			defer rf.Close()
			t.Assert(rf.StatusCode, 200)
			assetID = gjson.New(rf.ReadAllString()).Get("data.asset.assetId").String()
			t.AssertNE(assetID, "")

			rp, err := op().Patch(ctx, "/api/v1/resources/"+id, g.Map{"status": "published"})
			t.AssertNil(err)
			defer rp.Close()
			t.Assert(rp.StatusCode, 200)
			return id, assetID
		}

		// ── public resource: full lifecycle + download + count ──────────────
		id, assetID := makePublished("My CLI Tool")

		// get (anon) → resource + 1 file; tags round-trip through jsonb
		rg, err := anon().Get(ctx, "/api/v1/resources/"+id)
		t.AssertNil(err)
		jg := gjson.New(rg.ReadAllString())
		rg.Close()
		t.Assert(jg.Get("data.resource.title").String(), "My CLI Tool")
		t.Assert(jg.Get("data.resource.slug").String(), "my-cli-tool")
		t.Assert(len(jg.Get("data.assets").Array()), 1)
		t.Assert(len(jg.Get("data.resource.tags").Strings()), 2)

		// list (anon, type filter)
		rl, err := anon().Get(ctx, "/api/v1/resources", g.Map{"type": "software"})
		t.AssertNil(err)
		t.Assert(gjson.New(rl.ReadAllString()).Get("data.total").Int() >= 1, true)
		rl.Close()

		// site search (ILIKE q): title token match → found; gibberish → none
		rsq, err := anon().Get(ctx, "/api/v1/resources", g.Map{"q": "CLI"})
		t.AssertNil(err)
		t.Assert(gjson.New(rsq.ReadAllString()).Get("data.total").Int() >= 1, true)
		rsq.Close()
		rsq0, err := anon().Get(ctx, "/api/v1/resources", g.Map{"q": "zzznomatchxyz"})
		t.AssertNil(err)
		t.Assert(gjson.New(rsq0.ReadAllString()).Get("data.total").Int(), 0)
		rsq0.Close()

		// download (anon) → public cdn url, no gating
		rd, err := anon().Get(ctx, "/api/v1/resources/"+id+"/download/"+assetID)
		t.AssertNil(err)
		jd := gjson.New(rd.ReadAllString())
		rd.Close()
		t.Assert(jd.Get("code").String(), "ok")
		t.AssertNE(jd.Get("data.deliveryUrl").String(), "")

		// count incremented
		rg2, err := anon().Get(ctx, "/api/v1/resources/"+id)
		t.AssertNil(err)
		t.Assert(gjson.New(rg2.ReadAllString()).Get("data.resource.downloadCount").Int(), 1)
		rg2.Close()

		// asset_not_found
		rbad, err := anon().Get(ctx, "/api/v1/resources/"+id+"/download/not-a-real-asset")
		t.AssertNil(err)
		t.Assert(rbad.StatusCode, 404)
		t.Assert(gjson.New(rbad.ReadAllString()).Get("code").String(), "resource.asset_not_found")
		rbad.Close()

		// ── taxonomy: tag create (any operator) + assign + filter ────────────
		rtag, err := op().Post(ctx, "/api/v1/taxonomies", g.Map{"name": "CLI 工具", "taxonomy": "tag"})
		t.AssertNil(err)
		t.Assert(rtag.StatusCode, 200)
		tagID := gjson.New(rtag.ReadAllString()).Get("data.taxonomy.id").String()
		t.AssertNE(tagID, "")
		rtag.Close()

		// category create requires admin → 403 (token carries no admin role)
		rcat, err := op().Post(ctx, "/api/v1/taxonomies", g.Map{"name": "命令行", "taxonomy": "category"})
		t.AssertNil(err)
		t.Assert(rcat.StatusCode, 403)
		rcat.Close()

		// assign the tag to the resource
		rasg, err := op().Put(ctx, "/api/v1/resources/"+id+"/taxonomies", g.Map{"taxonomyIds": []string{tagID}})
		t.AssertNil(err)
		t.Assert(rasg.StatusCode, 200)
		t.Assert(gjson.New(rasg.ReadAllString()).Get("data.updated").Bool(), true)
		rasg.Close()

		// detail now carries the taxonomy
		rgt, err := anon().Get(ctx, "/api/v1/resources/"+id)
		t.AssertNil(err)
		t.Assert(len(gjson.New(rgt.ReadAllString()).Get("data.taxonomies").Array()), 1)
		rgt.Close()

		// browse filtered by the tag slug → the resource
		rfil, err := anon().Get(ctx, "/api/v1/resources", g.Map{"taxonomy": "cli"})
		t.AssertNil(err)
		t.Assert(gjson.New(rfil.ReadAllString()).Get("data.total").Int(), 1)
		rfil.Close()

		// public taxonomy list carries the published-resource count
		rtl, err := anon().Get(ctx, "/api/v1/taxonomies", g.Map{"taxonomy": "tag"})
		t.AssertNil(err)
		t.Assert(gjson.New(rtl.ReadAllString()).Get("data.items.0.count").Int(), 1)
		rtl.Close()

		// ── SEO: owner upserts, detail reflects it ──────────────────────────
		rseo, err := op().Put(ctx, "/api/v1/resources/"+id+"/seo", g.Map{"metaTitle": "CLI Tool — Download", "robots": "index,follow"})
		t.AssertNil(err)
		t.Assert(rseo.StatusCode, 200)
		t.Assert(gjson.New(rseo.ReadAllString()).Get("data.seo.metaTitle").String(), "CLI Tool — Download")
		rseo.Close()
		rgs, err := anon().Get(ctx, "/api/v1/resources/"+id)
		t.AssertNil(err)
		t.Assert(gjson.New(rgs.ReadAllString()).Get("data.seo.metaTitle").String(), "CLI Tool — Download")
		rgs.Close()

		// ── publish constraint: no asset → invalid_state ────────────────────
		re, err := op().Post(ctx, "/api/v1/resources", g.Map{"title": "Empty", "type": "software"})
		t.AssertNil(err)
		emptyID := gjson.New(re.ReadAllString()).Get("data.resource.id").String()
		re.Close()
		rep, err := op().Patch(ctx, "/api/v1/resources/"+emptyID, g.Map{"status": "published"})
		t.AssertNil(err)
		t.Assert(rep.StatusCode, 400)
		t.Assert(gjson.New(rep.ReadAllString()).Get("code").String(), "resource.invalid_state")
		rep.Close()

		// ── unknown type ────────────────────────────────────────────────────
		rt, err := op().Post(ctx, "/api/v1/resources", g.Map{"title": "X", "type": "nope"})
		t.AssertNil(err)
		t.Assert(rt.StatusCode, 400)
		t.Assert(gjson.New(rt.ReadAllString()).Get("code").String(), "resource.invalid_type")
		rt.Close()

		// ── draft not public (the "Empty" draft) ────────────────────────────
		rdraft, err := anon().Get(ctx, "/api/v1/resources/"+emptyID)
		t.AssertNil(err)
		t.Assert(rdraft.StatusCode, 404)
		rdraft.Close()

		// ── owner isolation: another user can't patch ───────────────────────
		jwt2 := signToken(t, priv, testSub2, time.Now().UTC().Add(10*time.Minute))
		c2 := g.Client()
		c2.SetPrefix(base)
		c2.ContentJson()
		c2.SetHeader("Authorization", "Bearer "+jwt2)
		r2, err := c2.Patch(ctx, "/api/v1/resources/"+id, g.Map{"title": "hijack"})
		t.AssertNil(err)
		t.Assert(r2.StatusCode, 404)
		r2.Close()

		// ── resilience: asset client fails → upstream_failed (502) ──────────
		rdraw, err := op().Post(ctx, "/api/v1/resources", g.Map{"title": "Resil", "type": "software"})
		t.AssertNil(err)
		resilID := gjson.New(rdraw.ReadAllString()).Get("data.resource.id").String()
		rdraw.Close()
		fake.FailNext()
		rfail, err := op().Post(ctx, "/api/v1/resources/"+resilID+"/assets", g.Map{"filename": "a.zip", "size": 1000})
		t.AssertNil(err)
		t.Assert(rfail.StatusCode, 502)
		t.Assert(gjson.New(rfail.ReadAllString()).Get("code").String(), "resource.upstream_failed")
		rfail.Close()

		// ── delete cleans up assets via the (fake) asset client ─────────────
		rdel, err := op().Delete(ctx, "/api/v1/resources/"+id)
		t.AssertNil(err)
		t.Assert(rdel.StatusCode, 200)
		rdel.Close()
		t.Assert(len(fake.Deleted()) >= 1, true)

		_, _ = db.Exec(ctx, "TRUNCATE resource_seo, object_taxonomies, taxonomies, terms, resource_assets, resources CASCADE")
	})
}
