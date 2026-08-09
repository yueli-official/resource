package resourceurls

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/yueli-official/foundation/go/identifier"
	"github.com/yueli-official/foundation/go/urllifecycle"
)

func TestPostgresResourceAndURLRollbackTogether(t *testing.T) {
	dsn := os.Getenv("URL_LIFECYCLE_CONSUMER_PG_DSN")
	if dsn == "" {
		t.Skip("URL_LIFECYCLE_CONSUMER_PG_DSN is not set")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	lifecycle, err := NewPostgres(ctx, db, "resource-test:"+identifier.MustNew().String(), "https://resource.test")
	if err != nil {
		t.Fatal(err)
	}
	id := identifier.MustNew().String()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO resources (
    id, owner_id, title, slug, type, status, published_at,
    delivery_kind, delivery_payload
) VALUES (
    $1::uuid, 'owner', 'Atomic', 'atomic', 'software', 'published', NOW(),
    'asset_file', '{}'::jsonb
)`, id); err != nil {
		t.Fatal(err)
	}
	if err := lifecycle.ReconcileResources(ctx, tx, []string{id}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	resolution, err := lifecycle.Resolver().Resolve(ctx, urllifecycle.Lookup{EscapedPath: "/resources/" + id})
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Kind != urllifecycle.ResolutionUnknown {
		t.Fatalf("URL state survived rollback: %#v", resolution)
	}
}
