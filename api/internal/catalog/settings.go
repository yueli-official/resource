package catalog

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/yueli-official/foundation/go/siteprofile"

	"platform/products/resource/api/internal/dao"
	"platform/products/resource/api/internal/reserr"
)

type SettingsLink struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	To    string `json:"to"`
	Icon  string `json:"icon"`
}

type FooterLinkGroup struct {
	ID    string         `json:"id"`
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
	Revision        uint64          `json:"revision"`
	RuntimeRevision uint64          `json:"runtimeRevision"`
	ETag            string          `json:"etag"`
	Site            SiteSection     `json:"site"`
	Footer          FooterSection   `json:"footer"`
	Resource        ResourceSection `json:"resource"`
}

type AdminSiteSettings struct {
	Snapshot        siteprofile.Snapshot
	Schema          siteprofile.FormSchema
	Resource        ResourceSection
	RuntimeRevision uint64
	ETag            string
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
	if s.profiles == nil {
		return SiteSettings{}, errors.New("resource site profile module is not configured")
	}
	payload, err := s.dao.SiteSettings(ctx, "site")
	if err != nil {
		return SiteSettings{}, err
	}
	snapshot, err := s.profiles.Get(ctx)
	if err != nil {
		return SiteSettings{}, err
	}
	settings := siteSettingsFromProfile(snapshot, payload)
	if err := validateResourceSettings(settings.Resource); err != nil {
		return SiteSettings{}, err
	}
	return settings, nil
}

func (s *Service) PublicSiteSettings(ctx context.Context) (SiteSettings, error) {
	if s.profiles == nil {
		return SiteSettings{}, errors.New("resource site profile module is not configured")
	}
	payload, err := s.dao.SiteSettings(ctx, "site")
	if err != nil {
		return SiteSettings{}, err
	}
	projection, err := s.profiles.PublicAt(ctx)
	if err != nil {
		return SiteSettings{}, err
	}
	settings := siteSettingsFromProfile(projection.Snapshot, payload)
	if err := validateResourceSettings(settings.Resource); err != nil {
		return SiteSettings{}, err
	}
	return settings, nil
}

func (s *Service) AdminSiteSettings(ctx context.Context) (AdminSiteSettings, error) {
	if s.profiles == nil {
		return AdminSiteSettings{}, errors.New("resource site profile module is not configured")
	}
	payload, err := s.dao.SiteSettings(ctx, "site")
	if err != nil {
		return AdminSiteSettings{}, err
	}
	snapshot, err := s.profiles.Get(ctx)
	if err != nil {
		return AdminSiteSettings{}, err
	}
	var runtime struct {
		Revision uint64          `json:"revision"`
		Resource ResourceSection `json:"resource"`
	}
	applyMap(payload, &runtime)
	runtime.Revision = normalizedRuntimeRevision(runtime.Revision)
	if err := validateResourceSettings(runtime.Resource); err != nil {
		return AdminSiteSettings{}, err
	}
	return AdminSiteSettings{
		Snapshot: snapshot, Schema: s.profiles.Schema(), Resource: runtime.Resource,
		RuntimeRevision: runtime.Revision, ETag: consumerETag(snapshot.ETag, runtime.Revision, runtime.Resource),
	}, nil
}

func (s *Service) SaveAdminSiteSettings(
	ctx context.Context,
	expected siteprofile.Revision,
	expectedRuntimeRevision uint64,
	profile siteprofile.Profile,
	resource ResourceSection,
) (AdminSiteSettings, error) {
	if s.profiles == nil {
		return AdminSiteSettings{}, errors.New("resource site profile module is not configured")
	}
	if err := validateResourceSettings(resource); err != nil {
		return AdminSiteSettings{}, err
	}
	payload := settingsToMap(struct {
		Revision uint64          `json:"revision"`
		Resource ResourceSection `json:"resource"`
	}{Revision: expectedRuntimeRevision + 1, Resource: resource})
	err := s.dao.ReplaceSiteSettingsWithHook(ctx, "site", payload, expectedRuntimeRevision, func(ctx context.Context, tx *sql.Tx) error {
		_, replaceErr := s.profiles.ReplaceTx(ctx, tx, siteprofile.ReplaceCommand{
			ExpectedRevision: expected,
			Profile:          profile,
		})
		return replaceErr
	})
	if errors.Is(err, dao.ErrSiteSettingsRevisionConflict) {
		return AdminSiteSettings{}, reserr.RevisionConflict()
	}
	if err != nil {
		return AdminSiteSettings{}, mapSiteProfileError(err)
	}
	settings, err := s.AdminSiteSettings(ctx)
	if err != nil {
		return AdminSiteSettings{}, err
	}
	if s.profileObserver != nil {
		if err := s.profileObserver.Refresh(settings.Snapshot); err != nil {
			return AdminSiteSettings{}, fmt.Errorf("refresh resource Discovery profile: %w", err)
		}
	}
	return settings, nil
}

func (s *Service) EnsureSiteProfile(ctx context.Context) error {
	if s.profiles == nil {
		return errors.New("resource site profile module is not configured")
	}
	if _, err := s.profiles.Get(ctx); err == nil {
		payload, payloadErr := s.dao.SiteSettings(ctx, "site")
		if payloadErr != nil {
			return payloadErr
		}
		if _, hasSite := payload["site"]; !hasSite {
			if _, hasFooter := payload["footer"]; !hasFooter {
				return nil
			}
		}
		current := siteSettingsFromMap(payload, s.siteBrand)
		revision := runtimeRevisionFromMap(payload)
		runtimePayload := settingsToMap(struct {
			Revision uint64          `json:"revision"`
			Resource ResourceSection `json:"resource"`
		}{Revision: revision, Resource: current.Resource})
		return s.dao.SaveSiteSettings(ctx, "site", runtimePayload)
	} else if !errors.Is(err, siteprofile.ErrNotInitialized) {
		return err
	}
	payload, err := s.dao.SiteSettings(ctx, "site")
	if err != nil {
		return err
	}
	legacy := siteSettingsFromMap(payload, s.siteBrand)
	if err := validateResourceSettings(legacy.Resource); err != nil {
		return fmt.Errorf("resource legacy site profile: %w", err)
	}
	runtimePayload := settingsToMap(struct {
		Revision uint64          `json:"revision"`
		Resource ResourceSection `json:"resource"`
	}{Revision: 1, Resource: legacy.Resource})
	_, err = s.dao.SaveSiteSettingsWithHook(ctx, "site", runtimePayload, func(ctx context.Context, tx *sql.Tx) error {
		_, replaceErr := s.profiles.ReplaceTx(ctx, tx, siteprofile.ReplaceCommand{
			Profile: profileFromSiteSettings(legacy),
		})
		return replaceErr
	})
	return err
}

func mapSiteProfileError(err error) error {
	var conflict *siteprofile.RevisionConflictError
	var validation *siteprofile.ValidationError
	switch {
	case errors.As(err, &conflict):
		return reserr.RevisionConflict()
	case errors.As(err, &validation):
		return reserr.InvalidInput(validation.Error())
	default:
		return err
	}
}

func validateHomeSettings(settings HomeSettings) error {
	if strings.TrimSpace(settings.HeroTitle) == "" || strings.TrimSpace(settings.HeroSubtitle) == "" || strings.TrimSpace(settings.IntroTitle) == "" || strings.TrimSpace(settings.IntroBody) == "" {
		return gerror.New("resource homepage content must be configured")
	}
	return nil
}

func validateResourceSettings(settings ResourceSection) error {
	if settings.ResourcesPerPage <= 0 || settings.LargeFileThresholdMB <= 0 || strings.TrimSpace(settings.LargeFileHint) == "" {
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

func siteSettingsFromProfile(snapshot siteprofile.Snapshot, payload map[string]any) SiteSettings {
	out := SiteSettings{
		Revision: uint64(snapshot.Revision), RuntimeRevision: runtimeRevisionFromMap(payload),
	}
	var runtime struct {
		Resource ResourceSection `json:"resource"`
	}
	applyMap(payload, &runtime)
	out.Resource = runtime.Resource
	out.ETag = consumerETag(snapshot.ETag, out.RuntimeRevision, out.Resource)
	profile := snapshot.Profile
	out.Site.SiteName = profile.Identity.Name
	out.Site.Tagline = profile.Identity.Tagline
	if profile.Branding.Logo != nil && profile.Branding.Logo.Kind == siteprofile.VisualIcon {
		out.Site.LogoIcon = profile.Branding.Logo.Ref
	}
	out.Site.Announcement = profile.Announcement.Text
	out.Site.AnnouncementEnabled = profile.Announcement.Enabled
	for _, contact := range profile.Support.Contacts {
		if contact.Kind == siteprofile.ContactEmail {
			out.Site.SupportEmail = contact.Value
			break
		}
	}
	out.Footer.Tagline = profile.Footer.Tagline
	out.Footer.Copyright = profile.Footer.Copyright
	out.Footer.Compliance.ExtraText = profile.Footer.Compliance.ExtraText
	for _, record := range profile.Footer.Compliance.Records {
		switch record.Kind {
		case "icp":
			out.Footer.Compliance.IcpRecord, out.Footer.Compliance.IcpURL = record.Number, record.URL
		case "police":
			out.Footer.Compliance.PoliceRecord, out.Footer.Compliance.PoliceURL = record.Number, record.URL
		}
	}
	for _, group := range profile.Footer.LinkGroups {
		item := FooterLinkGroup{ID: group.ID, Title: group.Title}
		for _, link := range group.Links {
			item.Links = append(item.Links, SettingsLink{ID: link.ID, Label: link.Label, To: link.Href, Icon: link.Icon})
		}
		out.Footer.LinkGroups = append(out.Footer.LinkGroups, item)
	}
	for _, social := range profile.Footer.Social {
		out.Footer.SocialLinks = append(out.Footer.SocialLinks, SettingsLink{
			ID: social.ID, Label: firstNonEmpty(social.Label, social.Platform), To: social.URL, Icon: social.Icon,
		})
	}
	return normalizeSiteSettings(out, "")
}

func normalizedRuntimeRevision(value uint64) uint64 {
	if value == 0 {
		return 1
	}
	return value
}

func runtimeRevisionFromMap(payload map[string]any) uint64 {
	var value struct {
		Revision uint64 `json:"revision"`
	}
	applyMap(payload, &value)
	return normalizedRuntimeRevision(value.Revision)
}

func consumerETag(profileETag string, runtimeRevision uint64, runtime any) string {
	raw, _ := json.Marshal(runtime)
	sum := sha256.Sum256(append([]byte(fmt.Sprintf("%s:%d:", profileETag, runtimeRevision)), raw...))
	return fmt.Sprintf(`"resource-settings-r%d-%s"`, runtimeRevision, hex.EncodeToString(sum[:8]))
}

func profileFromSiteSettings(settings SiteSettings) siteprofile.Profile {
	profile := siteprofile.Profile{
		Identity: siteprofile.Identity{Name: settings.Site.SiteName, Tagline: settings.Site.Tagline},
		Branding: siteprofile.Branding{Logo: &siteprofile.Visual{
			Kind: siteprofile.VisualIcon, Ref: settings.Site.LogoIcon, Alt: settings.Site.SiteName,
		}},
		Announcement: siteprofile.Announcement{
			Enabled: settings.Site.AnnouncementEnabled, Text: settings.Site.Announcement,
			Tone: siteprofile.AnnouncementInfo, Dismissible: true,
		},
		Footer: siteprofile.Footer{
			Tagline: settings.Footer.Tagline, Copyright: settings.Footer.Copyright,
			LinkGroups: []siteprofile.LinkGroup{}, Social: []siteprofile.SocialLink{}, Legal: []siteprofile.Link{},
			Compliance: siteprofile.Compliance{Records: []siteprofile.ComplianceRecord{}, ExtraText: settings.Footer.Compliance.ExtraText},
		},
		Support: siteprofile.Support{Contacts: []siteprofile.Contact{}},
	}
	if email := strings.TrimSpace(settings.Site.SupportEmail); email != "" {
		profile.Support.Contacts = append(profile.Support.Contacts, siteprofile.Contact{
			ID: "support-email", Kind: siteprofile.ContactEmail, Label: "支持邮箱", Value: email,
		})
	}
	for groupIndex, group := range settings.Footer.LinkGroups {
		groupID := stableSettingsID(group.ID, "footer-group", groupIndex)
		item := siteprofile.LinkGroup{ID: groupID, Title: group.Title, Links: []siteprofile.Link{}}
		for linkIndex, link := range group.Links {
			item.Links = append(item.Links, siteprofile.Link{
				ID:    stableSettingsID(link.ID, groupID+"-link", linkIndex),
				Label: link.Label, Href: link.To, Icon: link.Icon,
			})
		}
		profile.Footer.LinkGroups = append(profile.Footer.LinkGroups, item)
	}
	for index, link := range settings.Footer.SocialLinks {
		id := stableSettingsID(link.ID, "social", index)
		profile.Footer.Social = append(profile.Footer.Social, siteprofile.SocialLink{
			ID: id, Platform: socialPlatform(link.Label, id), Label: link.Label, URL: link.To, Icon: link.Icon,
		})
	}
	compliance := settings.Footer.Compliance
	if strings.TrimSpace(compliance.IcpRecord) != "" {
		profile.Footer.Compliance.Records = append(profile.Footer.Compliance.Records, siteprofile.ComplianceRecord{
			ID: "icp", Kind: "icp", Label: "ICP备案", Number: compliance.IcpRecord, URL: compliance.IcpURL,
		})
	}
	if strings.TrimSpace(compliance.PoliceRecord) != "" {
		profile.Footer.Compliance.Records = append(profile.Footer.Compliance.Records, siteprofile.ComplianceRecord{
			ID: "police", Kind: "police", Label: "公安备案", Number: compliance.PoliceRecord, URL: compliance.PoliceURL,
		})
	}
	return profile
}

func stableSettingsID(value, prefix string, index int) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fmt.Sprintf("%s-%d", prefix, index+1)
}

func socialPlatform(label, fallback string) string {
	value := strings.ToLower(strings.TrimSpace(label))
	value = strings.NewReplacer(" ", "-", "/", "-", "\\", "-").Replace(value)
	if value == "" {
		return fallback
	}
	return value
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
