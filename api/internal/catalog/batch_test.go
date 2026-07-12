package catalog

import "testing"

func TestNormalizeResourceBatchIDs(t *testing.T) {
	got := normalizeResourceBatchIDs([]string{" resource-1 ", "", "resource-1", "resource-2"})
	if len(got) != 2 || got[0] != "resource-1" || got[1] != "resource-2" {
		t.Fatalf("normalizeResourceBatchIDs() = %#v", got)
	}
}

func TestResourceBatchStatus(t *testing.T) {
	tests := map[string]string{
		"publish": "published",
		"draft":   "draft",
		"archive": "archived",
	}
	for action, want := range tests {
		got, err := resourceBatchStatus(action)
		if err != nil || got != want {
			t.Fatalf("resourceBatchStatus(%q) = %q, %v", action, got, err)
		}
	}
	if _, err := resourceBatchStatus("delete"); err == nil {
		t.Fatal("expected unsupported delete action to fail")
	}
}
