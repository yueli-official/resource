package v1

import "github.com/gogf/gf/v2/frame/g"

type SettingsLinkView struct {
	Label string `json:"label"`
	To    string `json:"to"`
	Icon  string `json:"icon"`
}

type FooterLinkGroupView struct {
	Title string             `json:"title"`
	Links []SettingsLinkView `json:"links"`
}

type ComplianceSettingsView struct {
	IcpRecord    string `json:"icpRecord"`
	IcpURL       string `json:"icpUrl"`
	PoliceRecord string `json:"policeRecord"`
	PoliceURL    string `json:"policeUrl"`
	ExtraText    string `json:"extraText"`
}

type HomeHighlightView struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type HomeSettingsView struct {
	HeroTitle           string              `json:"heroTitle"`
	HeroSubtitle        string              `json:"heroSubtitle"`
	IntroTitle          string              `json:"introTitle"`
	IntroBody           string              `json:"introBody"`
	IntroHighlights     []HomeHighlightView `json:"introHighlights"`
	FeaturedResourceIDs []string            `json:"featuredResourceIds"`
	CategorySlugs       []string            `json:"categorySlugs"`
	QuickLinks          []SettingsLinkView  `json:"quickLinks"`
}

type SiteSettingsView struct {
	Site     SiteSettingsSiteView     `json:"site"`
	Footer   SiteSettingsFooterView   `json:"footer"`
	Resource SiteSettingsResourceView `json:"resource"`
}

type SiteSettingsSiteView struct {
	SiteName            string `json:"siteName"`
	Tagline             string `json:"tagline"`
	LogoIcon            string `json:"logoIcon"`
	Announcement        string `json:"announcement"`
	AnnouncementEnabled bool   `json:"announcementEnabled"`
	SupportEmail        string `json:"supportEmail"`
}

type SiteSettingsFooterView struct {
	Tagline     string                 `json:"tagline"`
	Copyright   string                 `json:"copyright"`
	Compliance  ComplianceSettingsView `json:"compliance"`
	LinkGroups  []FooterLinkGroupView  `json:"linkGroups"`
	SocialLinks []SettingsLinkView     `json:"socialLinks"`
}

type SiteSettingsResourceView struct {
	ResourcesPerPage     int    `json:"resourcesPerPage"`
	DownloadsEnabled     bool   `json:"downloadsEnabled"`
	LargeFileThresholdMB int    `json:"largeFileThresholdMB"`
	LargeFileHint        string `json:"largeFileHint"`
}

type GetHomeSettingsReq struct {
	g.Meta `path:"/api/v1/resource/home" method:"get" tags:"resource" summary:"Get resource home settings"`
}

type GetHomeSettingsRes struct {
	Settings HomeSettingsView `json:"settings"`
}

type GetSiteSettingsReq struct {
	g.Meta `path:"/api/v1/resource/settings" method:"get" tags:"resource" summary:"Get resource site settings"`
}

type GetSiteSettingsRes struct {
	Settings SiteSettingsView `json:"settings"`
}

type AdminUpdateHomeSettingsReq struct {
	g.Meta `path:"/api/v1/admin/resource/home" method:"put" tags:"resource-admin" summary:"Update resource home settings"`
	HomeSettingsView
}

type AdminUpdateHomeSettingsRes struct {
	Settings HomeSettingsView `json:"settings"`
}

type AdminGetSiteSettingsReq struct {
	g.Meta `path:"/api/v1/admin/resource/settings" method:"get" tags:"resource-admin" summary:"Get admin resource settings"`
}

type AdminGetSiteSettingsRes struct {
	Settings SiteSettingsView `json:"settings"`
}

type AdminUpdateSiteSettingsReq struct {
	g.Meta `path:"/api/v1/admin/resource/settings" method:"put" tags:"resource-admin" summary:"Update resource settings"`
	SiteSettingsView
}

type AdminUpdateSiteSettingsRes struct {
	Settings SiteSettingsView `json:"settings"`
}
