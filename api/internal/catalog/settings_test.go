package catalog

import (
	"testing"
	"time"

	"github.com/yueli-official/foundation/go/siteprofile"
)

func TestSettingsUseOnlyPersistedContent(t *testing.T) {
	explicit := siteSettingsFromMap(map[string]any{
		"site": map[string]any{"siteName": "Operator Brand"},
	}, "月离资源")
	if explicit.Site.SiteName != "Operator Brand" {
		t.Fatalf("explicit site name = %q, want Operator Brand", explicit.Site.SiteName)
	}
	if explicit.Site.Tagline != "" {
		t.Fatalf("missing persisted tagline was synthesized: %q", explicit.Site.Tagline)
	}
}

func TestSiteProfileLegacyProjectionPreservesRuntimeAndStableIDs(t *testing.T) {
	legacy := SiteSettings{
		Site: SiteSection{
			SiteName: "资源库", Tagline: "下载资源", LogoIcon: "i-tabler-package",
			Announcement: "公告", AnnouncementEnabled: true, SupportEmail: "support@example.com",
		},
		Footer: FooterSection{
			Tagline: "页脚", Copyright: "Copyright",
			LinkGroups:  []FooterLinkGroup{{Title: "关于", Links: []SettingsLink{{Label: "介绍", To: "/about"}}}},
			SocialLinks: []SettingsLink{{Label: "GitHub", To: "https://github.com/example"}},
			Compliance:  ComplianceSettings{IcpRecord: "ICP 1", IcpURL: "https://beian.miit.gov.cn/"},
		},
		Resource: ResourceSection{ResourcesPerPage: 20, DownloadsEnabled: true, LargeFileThresholdMB: 100, LargeFileHint: "提示"},
	}
	profile := profileFromSiteSettings(legacy)
	if profile.Footer.LinkGroups[0].ID != "footer-group-1" || profile.Footer.LinkGroups[0].Links[0].ID != "footer-group-1-link-1" {
		t.Fatalf("generated stable IDs = %#v", profile.Footer.LinkGroups[0])
	}
	settings := siteSettingsFromProfile(siteprofile.Snapshot{
		Profile: profile, Revision: 3, ETag: `"etag"`, UpdatedAt: time.Now(),
	}, map[string]any{"resource": map[string]any{
		"resourcesPerPage": 12, "downloadsEnabled": false,
		"largeFileThresholdMB": 64, "largeFileHint": "runtime",
	}})
	if settings.Revision != 3 || settings.Resource.ResourcesPerPage != 12 || settings.Site.SiteName != "资源库" {
		t.Fatalf("projected settings = %#v", settings)
	}
	if settings.Footer.LinkGroups[0].ID != "footer-group-1" || settings.Footer.SocialLinks[0].ID != "social-1" {
		t.Fatalf("projected IDs = %#v %#v", settings.Footer.LinkGroups, settings.Footer.SocialLinks)
	}
}
