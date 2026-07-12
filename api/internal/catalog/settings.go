package catalog

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
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
	settings := homeSettingsFromMap(payload, s.siteBrand)
	if err := validateHomeSettings(settings); err != nil {
		return HomeSettings{}, err
	}
	return settings, nil
}

func (s *Service) SaveHomeSettings(ctx context.Context, in HomeSettings) (HomeSettings, error) {
	settings := normalizeHomeSettings(in, s.siteBrand)
	if err := validateHomeSettings(settings); err != nil {
		return HomeSettings{}, err
	}
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
	settings := siteSettingsFromMap(payload, s.siteBrand)
	if err := validateSiteSettings(settings); err != nil {
		return SiteSettings{}, err
	}
	return settings, nil
}

func (s *Service) SaveSiteSettings(ctx context.Context, in SiteSettings) (SiteSettings, error) {
	settings := normalizeSiteSettings(in, s.siteBrand)
	if err := validateSiteSettings(settings); err != nil {
		return SiteSettings{}, err
	}
	payload := settingsToMap(settings)
	if err := s.dao.SaveSiteSettings(ctx, "site", payload); err != nil {
		return SiteSettings{}, err
	}
	return settings, nil
}

func validateHomeSettings(settings HomeSettings) error {
	if strings.TrimSpace(settings.HeroTitle) == "" || strings.TrimSpace(settings.HeroSubtitle) == "" || strings.TrimSpace(settings.IntroTitle) == "" || strings.TrimSpace(settings.IntroBody) == "" {
		return gerror.New("resource homepage content must be configured")
	}
	return nil
}

func validateSiteSettings(settings SiteSettings) error {
	if strings.TrimSpace(settings.Site.SiteName) == "" || strings.TrimSpace(settings.Site.Tagline) == "" || strings.TrimSpace(settings.Site.LogoIcon) == "" || strings.TrimSpace(settings.Footer.Tagline) == "" || strings.TrimSpace(settings.Footer.Copyright) == "" {
		return gerror.New("resource site and footer content must be configured")
	}
	if settings.Resource.ResourcesPerPage <= 0 || settings.Resource.LargeFileThresholdMB <= 0 || strings.TrimSpace(settings.Resource.LargeFileHint) == "" {
		return gerror.New("resource runtime settings are incomplete")
	}
	return nil
}

func homeSettingsFromMap(payload map[string]any, siteBrand string) HomeSettings {
	out := HomeSettings{}
	applyMap(payload, &out)
	return normalizeHomeSettings(out, siteBrand)
}

func siteSettingsFromMap(payload map[string]any, siteBrand string) SiteSettings {
	out := SiteSettings{}
	applyMap(payload, &out)
	return normalizeSiteSettings(out, siteBrand)
}

func normalizeHomeSettings(in HomeSettings, siteBrand string) HomeSettings {
	_ = siteBrand
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
	_ = siteBrand
	if in.Footer.LinkGroups == nil {
		in.Footer.LinkGroups = []FooterLinkGroup{}
	}
	if in.Footer.SocialLinks == nil {
		in.Footer.SocialLinks = []SettingsLink{}
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
