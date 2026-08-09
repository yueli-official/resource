package resourcesearch

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/yueli-official/foundation/go/identifier"
)

func TestPostgresResourceAndProjectionCommitOrRollbackTogether(t *testing.T) {
	dsn := os.Getenv("SEARCH_CONSUMER_PG_DSN")
	if dsn == "" {
		t.Skip("SEARCH_CONSUMER_PG_DSN is not set")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	site := "test-" + identifier.MustNew().String()
	index, err := NewPostgres(ctx, db, site)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM search_instances WHERE instance_key=$1`, "resource."+site)
	})

	insert := func(tx *sql.Tx, id, title string) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO resources (
				id,owner_id,title,slug,description,type,tags,status,published_at,
				delivery_kind,delivery_payload
			) VALUES (
				$1::uuid,'owner',$2,$1::text,'body','software','["postgres","search"]'::jsonb,
				'published',NOW(),'asset_file','{}'::jsonb
			)
		`, id, title)
		return err
	}

	rolledBackID := identifier.MustNew().String()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := insert(tx, rolledBackID, "Rollback Resource Token"); err != nil {
		t.Fatal(err)
	}
	if err := index.Hook(rolledBackID)(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	page, err := index.Search(ctx, "Rollback Resource Token", "", nil, nil, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Hits) != 0 {
		t.Fatalf("rolled-back resource remained searchable: %#v", page.Hits)
	}

	committedID := identifier.MustNew().String()
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := insert(tx, committedID, "Committed Resource Token"); err != nil {
		t.Fatal(err)
	}
	if err := index.Hook(committedID)(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM resources WHERE id=$1::uuid`, committedID) })

	page, err = index.Search(ctx, "Committed Resource Token", "software", []string{"postgres", "search"}, nil, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Hits) != 1 || string(page.Hits[0].Key.ID) != committedID {
		t.Fatalf("committed search hits = %#v", page.Hits)
	}
}
