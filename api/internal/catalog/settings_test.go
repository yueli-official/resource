package catalog

import "testing"

func TestSettingsDefaultsUseInstanceBrand(t *testing.T) {
	home := homeSettingsFromMap(nil, "Ae Resource")
	if home.HeroTitle != "Ae Resource" {
		t.Fatalf("home hero title = %q, want Ae Resource", home.HeroTitle)
	}

	site := siteSettingsFromMap(nil, "Ae Resource")
	if site.Site.SiteName != "Ae Resource" {
		t.Fatalf("site name = %q, want Ae Resource", site.Site.SiteName)
	}

	explicit := siteSettingsFromMap(map[string]any{
		"site": map[string]any{"siteName": "Operator Brand"},
	}, "Ae Resource")
	if explicit.Site.SiteName != "Operator Brand" {
		t.Fatalf("explicit site name = %q, want Operator Brand", explicit.Site.SiteName)
	}
}
