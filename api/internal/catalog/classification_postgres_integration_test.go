package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/lib/pq"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/resource/api/internal/assetclient"
	resourcebootstrap "github.com/yueli-official/resource/api/internal/bootstrap"
	"github.com/yueli-official/resource/api/internal/dao"
	"github.com/yueli-official/resource/api/internal/model"
)

func TestPostgreSQLResourceClassificationConsumer(t *testing.T) {
	host := strings.TrimSpace(os.Getenv("RESOURCE_CLASSIFICATION_PG_HOST"))
	if host == "" {
		t.Skip("set RESOURCE_CLASSIFICATION_PG_HOST to run the Resource classification integration test")
	}
	port := integrationEnv("RESOURCE_CLASSIFICATION_PG_PORT", "5432")
	user := integrationEnv("RESOURCE_CLASSIFICATION_PG_USER", "postgres")
	password := os.Getenv("RESOURCE_CLASSIFICATION_PG_PASS")
	admin, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password,
	))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })

	database := fmt.Sprintf("resource_classification_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(database)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1`, database)
		_, _ = admin.Exec(`DROP DATABASE IF EXISTS ` + pq.QuoteIdentifier(database))
	})

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database,
	)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, name := range []string{
		"0001_init.up.sql",
		"0002_cover_url.up.sql",
		"0003_free_taxonomy_seo.up.sql",
		"0004_site_settings.up.sql",
		"0005_delivery_payload.up.sql",
		"0006_resource_operational_fields.up.sql",
	} {
		applyResourceMigration(t, sqlDB, name)
	}
	applyResourceMigration(t, sqlDB, "0012_classification_foundation.up.sql")
	applyResourceMigration(t, sqlDB, "0012_classification_foundation.down.sql")
	applyResourceMigration(t, sqlDB, "0012_classification_foundation.up.sql")
	if err := resourcebootstrap.ReconcileClassification(context.Background(), sqlDB); err != nil {
		t.Fatal(err)
	}

	databaseHandle, err := gdb.New(gdb.ConfigNode{
		Type: "pgsql", Host: host, Port: port, User: user, Pass: password, Name: database,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = databaseHandle.Close(context.Background()) })
	store := dao.NewPG(databaseHandle)
	service := New(store, assetclient.NewFake(), map[string]TypeRule{
		"software": {Label: "Software", AllowedExt: map[string]bool{"zip": true}, MaxSizeMB: 10},
	}, "resource-cover", "Resource")

	root, err := service.CreateTaxonomy(context.Background(), "Tools", "tools", "category", "", "")
	if err != nil {
		t.Fatal(err)
	}
	child, err := service.CreateTaxonomy(context.Background(), "CLI", "cli", "category", root.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	tag, err := service.CreateTaxonomy(context.Background(), "Go", "go", "tag", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateTaxonomy(context.Background(), "Nested", "nested", "tag", root.ID, ""); err == nil {
		t.Fatal("nested Tag creation should be rejected")
	}

	snapshot, err := store.ClassificationSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	compiled := classification.Compile(snapshot)
	if compiled.Outcome != classification.OutcomeAccepted || compiled.Catalog == nil {
		t.Fatalf("Compile() = %#v", compiled)
	}

	resource := &model.Resource{
		OwnerID: "author-1", Title: "Command line", Slug: "command-line",
		Type: "software", Status: model.StatusDraft,
	}
	if err := store.Insert(context.Background(), resource); err != nil {
		t.Fatal(err)
	}
	if err := service.AssignTaxonomies(
		context.Background(), resource.OwnerID, resource.ID, []string{child.ID, tag.ID},
	); err != nil {
		t.Fatal(err)
	}
	categories, err := service.ListTaxonomies(context.Background(), "category")
	if err != nil || len(categories) != 2 {
		t.Fatalf("ListTaxonomies(category) = %#v, %v", categories, err)
	}
	assigned, err := store.GetResourceTaxonomies(context.Background(), resource.ID)
	if err != nil || len(assigned) != 2 {
		t.Fatalf("GetResourceTaxonomies() = %#v, %v", assigned, err)
	}
	ids, err := store.ResourceIDsByTaxonomySlug(context.Background(), root.Slug)
	if err != nil || len(ids) != 1 || ids[0] != resource.ID {
		t.Fatalf("parent Category filter = %#v, %v", ids, err)
	}
	if err := service.DeleteTaxonomy(context.Background(), child.ID); err == nil {
		t.Fatal("assigned Category deletion should be rejected")
	}
	if _, err := service.UpdateTaxonomy(
		context.Background(), root.ID, nil, nil, nil, stringPointer(child.ID),
	); err == nil {
		t.Fatal("Category cycle should be rejected")
	}
	renamed := "Golang"
	if _, err := service.UpdateTaxonomy(
		context.Background(), tag.ID, &renamed, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
	oldLookup, err := service.classificationTagLookup(context.Background(), "Go")
	if err != nil {
		t.Fatal(err)
	}
	oldMatches, _, err := store.ClassificationTagMatches(
		context.Background(), []classification.TagLookupRequest{oldLookup},
	)
	if err != nil || len(oldMatches) != 1 ||
		oldMatches[0].Kind != classification.TagMatchAlias ||
		oldMatches[0].TagID != tag.ID {
		t.Fatalf("renamed Tag alias = %#v, %v", oldMatches, err)
	}
	target, err := service.CreateTaxonomy(context.Background(), "Go Language", "go-language", "tag", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.MergeTaxonomy(context.Background(), tag.ID, target.ID); err != nil {
		t.Fatal(err)
	}
	assigned, err = store.GetResourceTaxonomies(context.Background(), resource.ID)
	if err != nil || !hasTaxonomy(assigned, target.ID) || hasTaxonomy(assigned, tag.ID) {
		t.Fatalf("merged Tag assignments = %#v, %v", assigned, err)
	}
}

func applyResourceMigration(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	body, err := os.ReadFile("../../manifest/sql/migrations/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(body)); err != nil {
		t.Fatalf("apply %s: %v", name, err)
	}
}

func integrationEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func stringPointer(value string) *string { return &value }

func hasTaxonomy(values []*model.Taxonomy, id string) bool {
	for _, value := range values {
		if value.ID == id {
			return true
		}
	}
	return false
}
