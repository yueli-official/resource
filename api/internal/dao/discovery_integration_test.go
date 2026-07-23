package dao_test

import (
	"context"
	"os"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"

	"platform/products/resource/api/internal/dao"
)

func TestPGDiscoveryQuery(t *testing.T) {
	host := os.Getenv("RESOURCE_PG_HOST")
	if host == "" {
		t.Skip("set RESOURCE_PG_HOST to run the resource discovery integration test")
	}
	db, err := gdb.New(gdb.ConfigNode{
		Type: "pgsql", Host: host, Port: resourceEnvOr("RESOURCE_PG_PORT", "5432"),
		User: resourceEnvOr("RESOURCE_PG_USER", "postgres"),
		Pass: os.Getenv("RESOURCE_PG_PASS"), Name: "resource",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dao.NewPG(db).ListDiscoveryPages(
		context.Background(), "https://resource.example.com", "", 1,
	); err != nil {
		t.Fatal(err)
	}
}

func resourceEnvOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
