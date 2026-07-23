package resourceurls

import (
	"context"
	"testing"

	"github.com/yueli-official/foundation/go/urllifecycle"
)

func TestTaxonomyRenameAndResourceDelete(t *testing.T) {
	ctx := context.Background()
	lifecycle, err := NewMemory("https://resources.example")
	if err != nil {
		t.Fatal(err)
	}
	tag := TaxonomyState{ID: "tag-1", Kind: TagKind, Slug: "old"}
	if err := ensure(ctx, lifecycle.module, taxonomyRoute(tag), taxonomyRef(tag), "claim"); err != nil {
		t.Fatal(err)
	}
	tag.Slug = "new"
	if err := ensure(ctx, lifecycle.module, taxonomyRoute(tag), taxonomyRef(tag), "rename"); err != nil {
		t.Fatal(err)
	}
	resolved, err := lifecycle.Resolver().Resolve(ctx, urllifecycle.Lookup{EscapedPath: "/tags/old"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Kind != urllifecycle.ResolutionRedirect || resolved.Location != "/tags/new" {
		t.Fatalf("old taxonomy resolution = %#v", resolved)
	}
	key := resourceRoute("resource-1")
	if err := ensure(ctx, lifecycle.module, key, urllifecycle.LocalRef{Path: "/resources/resource-1"}, "publish"); err != nil {
		t.Fatal(err)
	}
	if err := gone(ctx, lifecycle.module, key, "delete"); err != nil {
		t.Fatal(err)
	}
	resolved, err = lifecycle.Resolver().Resolve(ctx, urllifecycle.Lookup{EscapedPath: "/resources/resource-1"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Kind != urllifecycle.ResolutionGone {
		t.Fatalf("resource resolution = %#v", resolved)
	}
}
