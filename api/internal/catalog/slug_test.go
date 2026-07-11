package catalog

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello World":   "hello-world",
		"  Foo__Bar!! ": "foo-bar",
		"App v2.0":      "app-v2-0",
		"---a---":       "a",
		"软件工具":          "", // all non-ascii → empty (caller falls back to "r")
		"MyTool 3":      "mytool-3",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtOf(t *testing.T) {
	for in, want := range map[string]string{"a.ZIP": "zip", "x.tar.gz": "gz", "noext": "", "a.PnG": "png"} {
		if got := extOf(in); got != want {
			t.Errorf("extOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCoverImageExtensionsMatchAssetProfile(t *testing.T) {
	for _, ext := range []string{"jpg", "jpeg", "png", "webp"} {
		if !isImageExt(ext) {
			t.Fatalf("cover extension %q was rejected", ext)
		}
	}
	if isImageExt("gif") {
		t.Fatal("gif cover was accepted outside the resource-cover profile")
	}
}
