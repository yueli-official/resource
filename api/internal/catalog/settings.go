package catalog

import (
	"context"
	"encoding/json"
)

type SettingsLink struct {
	Label string `json:"label"`
	To    string `json:"to"`
	Icon  string `json:"icon"`
}

type FooterLinkGroup struct {
	Title string         `json:"title"`
	Links []SettingsLink `json:"links"`
}

type ComplianceSettings struct {
	IcpRecord    string `json:"icpRecord"`
	IcpURL       string `json:"icpUrl"`
	PoliceRecord string `json:"policeRecord"`
	PoliceURL    string `json:"policeUrl"`
	ExtraText    string `json:"extraText"`
}

type HomeHighlight struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type HomeSettings struct {
	HeroTitle           string          `json:"heroTitle"`
	HeroSubtitle        string          `json:"heroSubtitle"`
	IntroTitle          string          `json:"introTitle"`
	IntroBody           string          `json:"introBody"`
	IntroHighlights     []HomeHighlight `json:"introHighlights"`
	FeaturedResourceIDs []string        `json:"featuredResourceIds"`
	CategorySlugs       []string        `json:"categorySlugs"`
	QuickLinks          []SettingsLink  `json:"quickLinks"`
}

type SiteSettings struct {
	Site     SiteSection     `json:"site"`
	Footer   FooterSection   `json:"footer"`
	Resource ResourceSection `json:"resource"`
}

type SiteSection struct {
	SiteName            string `json:"siteName"`
	Tagline             string `json:"tagline"`
	LogoIcon            string `json:"logoIcon"`
	Announcement        string `json:"announcement"`
	AnnouncementEnabled bool   `json:"announcementEnabled"`
	SupportEmail        string `json:"supportEmail"`
}

type FooterSection struct {
	Tagline     string             `json:"tagline"`
	Copyright   string             `json:"copyright"`
	Compliance  ComplianceSettings `json:"compliance"`
	LinkGroups  []FooterLinkGroup  `json:"linkGroups"`
	SocialLinks []SettingsLink     `json:"socialLinks"`
}

type ResourceSection struct {
	ResourcesPerPage     int    `json:"resourcesPerPage"`
	DownloadsEnabled     bool   `json:"downloadsEnabled"`
	LargeFileThresholdMB int    `json:"largeFileThresholdMB"`
	LargeFileHint        string `json:"largeFileHint"`
}

func (s *Service) HomeSettings(ctx context.Context) (HomeSettings, error) {
	payload, err := s.dao.HomeSettings(ctx)
	if err != nil {
		return HomeSettings{}, err
	}
	return homeSettingsFromMap(payload, s.siteBrand), nil
}

func (s *Service) SaveHomeSettings(ctx context.Context, in HomeSettings) (HomeSettings, error) {
	settings := normalizeHomeSettings(in, s.siteBrand)
	payload := settingsToMap(settings)
	if err := s.dao.SaveHomeSettings(ctx, payload); err != nil {
		return HomeSettings{}, err
	}
	return settings, nil
}

func (s *Service) SiteSettings(ctx context.Context) (SiteSettings, error) {
	payload, err := s.dao.SiteSettings(ctx, "site")
	if err != nil {
		return SiteSettings{}, err
	}
	return siteSettingsFromMap(payload, s.siteBrand), nil
}

func (s *Service) SaveSiteSettings(ctx context.Context, in SiteSettings) (SiteSettings, error) {
	settings := normalizeSiteSettings(in, s.siteBrand)
	payload := settingsToMap(settings)
	if err := s.dao.SaveSiteSettings(ctx, "site", payload); err != nil {
		return SiteSettings{}, err
	}
	return settings, nil
}

func defaultHomeSettings(siteBrand string) HomeSettings {
	return HomeSettings{
		HeroTitle:    siteBrand,
		HeroSubtitle: "浏览并下载软件、设计素材与脚本。",
		IntroTitle:   "面向创作和开发的资源目录",
		IntroBody:    "这里集中展示可下载工具、设计素材、脚本和模板，支持分类浏览、标签搜索和文件下载。",
		IntroHighlights: []HomeHighlight{
			{Icon: "i-tabler-package", Title: "可下载资源", Text: "集中管理软件、脚本、模板和素材。"},
			{Icon: "i-tabler-tags", Title: "分类标签", Text: "用目录和标签组织不同用途的资源。"},
			{Icon: "i-tabler-cloud-download", Title: "稳定交付", Text: "基于资源中心管理文件、封面和下载。"},
		},
		CategorySlugs: []string{},
		QuickLinks: []SettingsLink{
			{Label: "浏览资源", To: "/", Icon: "i-tabler-package"},
			{Label: "搜索", To: "/search", Icon: "i-tabler-search"},
		},
	}
}

func defaultSiteSettings(siteBrand string) SiteSettings {
	return SiteSettings{
		Site: SiteSection{
			SiteName:     siteBrand,
			Tagline:      "软件、设计素材与脚本下载",
			LogoIcon:     "i-tabler-package",
			SupportEmail: "",
		},
		Footer: FooterSection{
			Tagline:   "可下载的软件、设计素材与脚本。",
			Copyright: "© 2026 Yueli",
			LinkGroups: []FooterLinkGroup{
				{Title: "资源", Links: []SettingsLink{
					{Label: "全部资源", To: "/", Icon: "i-tabler-package"},
					{Label: "搜索资源", To: "/search", Icon: "i-tabler-search"},
				}},
			},
		},
		Resource: ResourceSection{
			ResourcesPerPage:     12,
			DownloadsEnabled:     true,
			LargeFileThresholdMB: 500,
			LargeFileHint:        "大文件建议使用 OSS/COS/S3 分片上传或网盘，不建议长期走本地存储。",
		},
	}
}

func homeSettingsFromMap(payload map[string]any, siteBrand string) HomeSettings {
	out := defaultHomeSettings(siteBrand)
	applyMap(payload, &out)
	return normalizeHomeSettings(out, siteBrand)
}

func siteSettingsFromMap(payload map[string]any, siteBrand string) SiteSettings {
	out := defaultSiteSettings(siteBrand)
	applyMap(payload, &out)
	return normalizeSiteSettings(out, siteBrand)
}

func normalizeHomeSettings(in HomeSettings, siteBrand string) HomeSettings {
	defaults := defaultHomeSettings(siteBrand)
	if in.HeroTitle == "" {
		in.HeroTitle = defaults.HeroTitle
	}
	if in.HeroSubtitle == "" {
		in.HeroSubtitle = defaults.HeroSubtitle
	}
	if in.IntroTitle == "" {
		in.IntroTitle = defaults.IntroTitle
	}
	if in.IntroBody == "" {
		in.IntroBody = defaults.IntroBody
	}
	if in.IntroHighlights == nil {
		in.IntroHighlights = []HomeHighlight{}
	}
	if in.FeaturedResourceIDs == nil {
		in.FeaturedResourceIDs = []string{}
	}
	if in.CategorySlugs == nil {
		in.CategorySlugs = []string{}
	}
	if in.QuickLinks == nil {
		in.QuickLinks = []SettingsLink{}
	}
	return in
}

func normalizeSiteSettings(in SiteSettings, siteBrand string) SiteSettings {
	defaults := defaultSiteSettings(siteBrand)
	if in.Site.SiteName == "" {
		in.Site.SiteName = defaults.Site.SiteName
	}
	if in.Site.Tagline == "" {
		in.Site.Tagline = defaults.Site.Tagline
	}
	if in.Site.LogoIcon == "" {
		in.Site.LogoIcon = defaults.Site.LogoIcon
	}
	if in.Footer.Tagline == "" {
		in.Footer.Tagline = defaults.Footer.Tagline
	}
	if in.Footer.Copyright == "" {
		in.Footer.Copyright = defaults.Footer.Copyright
	}
	if in.Footer.LinkGroups == nil {
		in.Footer.LinkGroups = []FooterLinkGroup{}
	}
	if in.Footer.SocialLinks == nil {
		in.Footer.SocialLinks = []SettingsLink{}
	}
	if in.Resource.ResourcesPerPage <= 0 || in.Resource.ResourcesPerPage > 96 {
		in.Resource.ResourcesPerPage = defaults.Resource.ResourcesPerPage
	}
	if in.Resource.LargeFileThresholdMB <= 0 {
		in.Resource.LargeFileThresholdMB = defaults.Resource.LargeFileThresholdMB
	}
	if in.Resource.LargeFileHint == "" {
		in.Resource.LargeFileHint = defaults.Resource.LargeFileHint
	}
	return in
}

func settingsToMap(v any) map[string]any {
	raw, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func applyMap(payload map[string]any, target any) {
	if len(payload) == 0 {
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, target)
}
