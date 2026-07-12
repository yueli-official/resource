package catalog

import "testing"

func TestSettingsUseOnlyPersistedContent(t *testing.T) {
	explicit := siteSettingsFromMap(map[string]any{
		"site": map[string]any{"siteName": "Operator Brand"},
	}, "Ae Resource")
	if explicit.Site.SiteName != "Operator Brand" {
		t.Fatalf("explicit site name = %q, want Operator Brand", explicit.Site.SiteName)
	}
	if explicit.Site.Tagline != "" {
		t.Fatalf("missing persisted tagline was synthesized: %q", explicit.Site.Tagline)
	}
}
