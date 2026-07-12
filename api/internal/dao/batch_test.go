package dao

import (
	"strings"
	"testing"
)

func TestOwnerResourceBatchSQLGuardsOwnerAndPublishability(t *testing.T) {
	query, args := ownerResourceBatchSQL("owner-1", []string{"resource-1", "resource-2"}, "published")
	for _, fragment := range []string{"owner_id = ?", "resource_assets", "delivery_payload", "RETURNING id"} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("query missing %q: %s", fragment, query)
		}
	}
	if len(args) != 6 {
		t.Fatalf("args length = %d, want 6", len(args))
	}
	if args[2] != "owner-1" || args[5] != "published" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestOwnerResourceBatchSQLEmptyIDsIsNoop(t *testing.T) {
	query, args := ownerResourceBatchSQL("owner-1", nil, "draft")
	if query != "" || args != nil {
		t.Fatalf("empty batch = %q, %#v", query, args)
	}
}
