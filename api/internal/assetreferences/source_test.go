package assetreferences

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/yueli-official/asset/referencesync"
	"os"
	"testing"
)

// Real PostgreSQL query tests use transaction-local tables only; no production data is touched.
func TestCommittedUsageLifecycle(t *testing.T) {
	dsn := os.Getenv("REFERENCE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set REFERENCE_TEST_DATABASE_URL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	exec := func(query string) {
		t.Helper()
		if _, err := tx.ExecContext(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	exec(`CREATE TEMP TABLE resource_seo(resource_id text,og_image text) ON COMMIT DROP; CREATE TEMP TABLE resources(id text,title text,cover_asset_id text,description text) ON COMMIT DROP; CREATE TEMP TABLE resource_assets(resource_id text,asset_id text) ON COMMIT DROP;INSERT INTO resources VALUES ('a','A','',''),('b','B','',''); INSERT INTO resource_assets VALUES ('a','KeyA'),('b','KeyA')`)
	read := func(want int) []referencesync.Reference {
		t.Helper()
		snapshots, err := Source()(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range snapshots {
			if s.RefType == "resource-file" {
				if len(s.References) != want {
					t.Fatalf("references=%d want=%d", len(s.References), want)
				}
				return s.References
			}
		}
		t.Fatal("snapshot missing")
		return nil
	}
	read(2)
	read(2) // Backfill and repeat keep two independent users of the same asset.
	exec(`DELETE FROM resources WHERE id='a'`)
	read(1)
	exec(`INSERT INTO resources VALUES ('a','A','','')`)
	read(2)
	exec(`UPDATE resource_assets SET asset_id='KeyB' WHERE resource_id='a'`)
	for _, ref := range read(2) {
		if ref.RefID == "a" && ref.AssetID != "KeyB" && ref.MediaKey != "KeyB" {
			t.Fatalf("replacement retained old asset: %+v", ref)
		}
	}
}
