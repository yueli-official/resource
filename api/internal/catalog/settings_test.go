package catalog

import (
	"testing"
	"time"

	"github.com/yueli-official/foundation/go/siteprofile"
)

func TestSiteProfileProjectionCombinesProfileAndRuntimeSettings(t *testing.T) {
	profile := siteprofile.Profile{
		Identity: siteprofile.Identity{Name: "资源库", Tagline: "下载资源"},
		Branding: siteprofile.Branding{Logo: &siteprofile.Visual{
			Kind: siteprofile.VisualIcon, Ref: "i-tabler-package",
		}},
		Footer: siteprofile.Footer{
			Tagline: "页脚", Copyright: "Copyright",
			LinkGroups: []siteprofile.LinkGroup{{
				ID: "about", Title: "关于",
				Links: []siteprofile.Link{{ID: "intro", Label: "介绍", Href: "/about"}},
			}},
			Social: []siteprofile.SocialLink{{
				ID: "github", Platform: "github", Label: "GitHub", URL: "https://github.com/example",
			}},
		},
	}
	settings := siteSettingsFromProfile(siteprofile.Snapshot{
		Profile: profile, Revision: 3, ETag: `"etag"`, UpdatedAt: time.Now(),
	}, map[string]any{
		"revision": float64(2),
		"resource": map[string]any{
			"resourcesPerPage": 12, "downloadsEnabled": false,
			"largeFileThresholdMB": 64, "largeFileHint": "runtime",
		},
	})
	if settings.Revision != 3 || settings.Resource.ResourcesPerPage != 12 || settings.Site.SiteName != "资源库" {
		t.Fatalf("projected settings = %#v", settings)
	}
	if settings.RuntimeRevision != 2 || settings.Footer.LinkGroups[0].ID != "about" || settings.Footer.SocialLinks[0].ID != "github" {
		t.Fatalf("projected IDs = %#v %#v", settings.Footer.LinkGroups, settings.Footer.SocialLinks)
	}
}
