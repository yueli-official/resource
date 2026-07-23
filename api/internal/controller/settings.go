package controller

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/catalog"
	"platform/products/resource/api/internal/reserr"
)

func (c *PublicResources) GetHomeSettings(ctx context.Context, req *v1.GetHomeSettingsReq) (*v1.GetHomeSettingsRes, error) {
	settings, err := c.svc.HomeSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetHomeSettingsRes{Settings: homeSettingsView(settings)}, nil
}

func (c *PublicResources) GetSiteSettings(ctx context.Context, req *v1.GetSiteSettingsReq) (*v1.GetSiteSettingsRes, error) {
	settings, err := c.svc.PublicSiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	setSiteProfileHeaders(ctx, settings.ETag, true)
	return &v1.GetSiteSettingsRes{Settings: siteSettingsView(settings)}, nil
}

func (c *Resources) AdminUpdateHomeSettings(ctx context.Context, req *v1.AdminUpdateHomeSettingsReq) (*v1.AdminUpdateHomeSettingsRes, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	settings, err := c.svc.SaveHomeSettings(ctx, homeSettingsInput(req.HomeSettingsView))
	if err != nil {
		return nil, err
	}
	return &v1.AdminUpdateHomeSettingsRes{Settings: homeSettingsView(settings)}, nil
}

func (c *Resources) AdminGetSiteSettings(ctx context.Context, req *v1.AdminGetSiteSettingsReq) (*v1.AdminGetSiteSettingsRes, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	settings, err := c.svc.AdminSiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	setSiteProfileHeaders(ctx, settings.ETag, false)
	return &v1.AdminGetSiteSettingsRes{Settings: adminSiteSettingsView(settings)}, nil
}

func (c *Resources) AdminUpdateSiteSettings(ctx context.Context, req *v1.AdminUpdateSiteSettingsReq) (*v1.AdminUpdateSiteSettingsRes, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	current, err := c.svc.AdminSiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	ifMatch := strings.TrimSpace(g.RequestFromCtx(ctx).Header.Get("If-Match"))
	if ifMatch == "" {
		return nil, reserr.PreconditionRequired()
	}
	if ifMatch != current.ETag {
		return nil, reserr.RevisionConflict()
	}
	settings, err := c.svc.SaveAdminSiteSettings(
		ctx,
		current.Snapshot.Revision,
		current.RuntimeRevision,
		req.Profile,
		resourceSettingsInput(req.Resource),
	)
	if err != nil {
		return nil, err
	}
	setSiteProfileHeaders(ctx, settings.ETag, false)
	return &v1.AdminUpdateSiteSettingsRes{Settings: adminSiteSettingsView(settings)}, nil
}

func setSiteProfileHeaders(ctx context.Context, etag string, public bool) {
	request := g.RequestFromCtx(ctx)
	request.Response.Header().Set("ETag", etag)
	if public {
		request.Response.Header().Set("Cache-Control", "public, no-cache")
		return
	}
	request.Response.Header().Set("Cache-Control", "private, no-store")
}

func homeSettingsView(settings catalog.HomeSettings) v1.HomeSettingsView {
	out := v1.HomeSettingsView{
		HeroTitle: settings.HeroTitle, HeroSubtitle: settings.HeroSubtitle, IntroTitle: settings.IntroTitle, IntroBody: settings.IntroBody,
		FeaturedResourceIDs: settings.FeaturedResourceIDs, CategorySlugs: settings.CategorySlugs,
	}
	for _, item := range settings.IntroHighlights {
		out.IntroHighlights = append(out.IntroHighlights, v1.HomeHighlightView{Icon: item.Icon, Title: item.Title, Text: item.Text})
	}
	out.QuickLinks = settingsLinkViews(settings.QuickLinks)
	return out
}

func homeSettingsInput(settings v1.HomeSettingsView) catalog.HomeSettings {
	out := catalog.HomeSettings{
		HeroTitle: settings.HeroTitle, HeroSubtitle: settings.HeroSubtitle, IntroTitle: settings.IntroTitle, IntroBody: settings.IntroBody,
		FeaturedResourceIDs: settings.FeaturedResourceIDs, CategorySlugs: settings.CategorySlugs,
	}
	for _, item := range settings.IntroHighlights {
		out.IntroHighlights = append(out.IntroHighlights, catalog.HomeHighlight{Icon: item.Icon, Title: item.Title, Text: item.Text})
	}
	out.QuickLinks = settingsLinksInput(settings.QuickLinks)
	return out
}

func siteSettingsView(settings catalog.SiteSettings) v1.SiteSettingsView {
	return v1.SiteSettingsView{
		Revision:        settings.Revision,
		RuntimeRevision: settings.RuntimeRevision,
		ETag:            settings.ETag,
		Site: v1.SiteSettingsSiteView{
			SiteName: settings.Site.SiteName, Tagline: settings.Site.Tagline, LogoIcon: settings.Site.LogoIcon,
			Announcement: settings.Site.Announcement, AnnouncementEnabled: settings.Site.AnnouncementEnabled,
			SupportEmail: settings.Site.SupportEmail,
		},
		Footer: v1.SiteSettingsFooterView{
			Tagline: settings.Footer.Tagline, Copyright: settings.Footer.Copyright,
			Compliance: v1.ComplianceSettingsView{
				IcpRecord: settings.Footer.Compliance.IcpRecord, IcpURL: settings.Footer.Compliance.IcpURL,
				PoliceRecord: settings.Footer.Compliance.PoliceRecord, PoliceURL: settings.Footer.Compliance.PoliceURL,
				ExtraText: settings.Footer.Compliance.ExtraText,
			},
			LinkGroups: footerGroupViews(settings.Footer.LinkGroups), SocialLinks: settingsLinkViews(settings.Footer.SocialLinks),
		},
		Resource: v1.SiteSettingsResourceView{
			ResourcesPerPage: settings.Resource.ResourcesPerPage, DownloadsEnabled: settings.Resource.DownloadsEnabled,
			LargeFileThresholdMB: settings.Resource.LargeFileThresholdMB, LargeFileHint: settings.Resource.LargeFileHint,
		},
	}
}

func adminSiteSettingsView(settings catalog.AdminSiteSettings) v1.AdminSiteSettingsView {
	return v1.AdminSiteSettingsView{
		Snapshot:        settings.Snapshot,
		Schema:          settings.Schema,
		Resource:        resourceSettingsView(settings.Resource),
		RuntimeRevision: settings.RuntimeRevision,
		ETag:            settings.ETag,
	}
}

func resourceSettingsView(settings catalog.ResourceSection) v1.SiteSettingsResourceView {
	return v1.SiteSettingsResourceView{
		ResourcesPerPage:     settings.ResourcesPerPage,
		DownloadsEnabled:     settings.DownloadsEnabled,
		LargeFileThresholdMB: settings.LargeFileThresholdMB,
		LargeFileHint:        settings.LargeFileHint,
	}
}

func resourceSettingsInput(settings v1.SiteSettingsResourceView) catalog.ResourceSection {
	return catalog.ResourceSection{
		ResourcesPerPage:     settings.ResourcesPerPage,
		DownloadsEnabled:     settings.DownloadsEnabled,
		LargeFileThresholdMB: settings.LargeFileThresholdMB,
		LargeFileHint:        settings.LargeFileHint,
	}
}

func settingsLinkViews(links []catalog.SettingsLink) []v1.SettingsLinkView {
	out := make([]v1.SettingsLinkView, 0, len(links))
	for _, link := range links {
		out = append(out, v1.SettingsLinkView{ID: link.ID, Label: link.Label, To: link.To, Icon: link.Icon})
	}
	return out
}

func settingsLinksInput(links []v1.SettingsLinkView) []catalog.SettingsLink {
	out := make([]catalog.SettingsLink, 0, len(links))
	for _, link := range links {
		out = append(out, catalog.SettingsLink{ID: link.ID, Label: link.Label, To: link.To, Icon: link.Icon})
	}
	return out
}

func footerGroupViews(groups []catalog.FooterLinkGroup) []v1.FooterLinkGroupView {
	out := make([]v1.FooterLinkGroupView, 0, len(groups))
	for _, group := range groups {
		out = append(out, v1.FooterLinkGroupView{ID: group.ID, Title: group.Title, Links: settingsLinkViews(group.Links)})
	}
	return out
}

func footerGroupsInput(groups []v1.FooterLinkGroupView) []catalog.FooterLinkGroup {
	out := make([]catalog.FooterLinkGroup, 0, len(groups))
	for _, group := range groups {
		out = append(out, catalog.FooterLinkGroup{ID: group.ID, Title: group.Title, Links: settingsLinksInput(group.Links)})
	}
	return out
}
