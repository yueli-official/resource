import type { HomeSettingsView, SiteSettingsView } from "~/types";

function emptyHomeSettings(): HomeSettingsView {
  return {
    heroTitle: "",
    heroSubtitle: "",
    introTitle: "",
    introBody: "",
    introHighlights: [],
    featuredResourceIds: [],
    categorySlugs: [],
    quickLinks: [],
  };
}

function emptySiteSettings(): SiteSettingsView {
  return {
    revision: 0,
    runtimeRevision: 0,
    etag: "",
    site: {
      siteName: "",
      tagline: "",
      logoIcon: "",
      announcement: "",
      announcementEnabled: false,
      supportEmail: "",
    },
    footer: {
      tagline: "",
      copyright: "",
      compliance: {
        icpRecord: "",
        icpUrl: "",
        policeRecord: "",
        policeUrl: "",
        extraText: "",
      },
      linkGroups: [],
      socialLinks: [],
    },
    resource: {
      resourcesPerPage: 0,
      downloadsEnabled: false,
      largeFileThresholdMB: 0,
      largeFileHint: "",
    },
  };
}

export function useResourceSettings() {
  const { call } = useApi();
  const {
    data: siteData,
    pending: sitePending,
    error: siteError,
  } = useAsyncData(
    "resource-site-settings",
    () => call<{ settings: SiteSettingsView }>("/api/v1/resource/settings"),
    { default: () => ({ settings: emptySiteSettings() }) },
  );
  const {
    data: homeData,
    pending: homePending,
    error: homeError,
  } = useAsyncData(
    "resource-home-settings",
    () => call<{ settings: HomeSettingsView }>("/api/v1/resource/home"),
    { default: () => ({ settings: emptyHomeSettings() }) },
  );

  const siteSettings = computed(() => siteData.value!.settings);
  const homeSettings = computed(() => homeData.value!.settings);
  const pending = computed(() => sitePending.value || homePending.value);

  const error = computed(() => siteError.value || homeError.value);
  return { siteSettings, homeSettings, pending, error };
}
