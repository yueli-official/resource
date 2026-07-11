import type { HomeSettingsView, SiteSettingsView } from '~/types'

function defaultHomeSettings(siteBrand: string): HomeSettingsView {
  return {
    heroTitle: siteBrand,
    heroSubtitle: '浏览并下载软件、设计素材与脚本。',
    introTitle: '面向创作和开发的资源目录',
    introBody: '这里集中展示可下载工具、设计素材、脚本和模板，支持分类浏览、标签搜索和文件下载。',
    introHighlights: [],
    featuredResourceIds: [],
    categorySlugs: [],
    quickLinks: []
  }
}

function defaultSiteSettings(siteBrand: string): SiteSettingsView {
  return {
    site: {
      siteName: siteBrand,
      tagline: '软件、设计素材与脚本下载',
      logoIcon: 'i-tabler-package',
      announcement: '',
      announcementEnabled: false,
      supportEmail: ''
    },
    footer: {
      tagline: '可下载的软件、设计素材与脚本。',
      copyright: '© 2026 Yueli',
      compliance: { icpRecord: '', icpUrl: '', policeRecord: '', policeUrl: '', extraText: '' },
      linkGroups: [],
      socialLinks: []
    },
    resource: {
      resourcesPerPage: 12,
      downloadsEnabled: true,
      largeFileThresholdMB: 500,
      largeFileHint: '大文件建议使用 OSS/COS/S3 分片上传或网盘，不建议长期走本地存储。'
    }
  }
}

export function useResourceSettings() {
  const { call } = useApi()
  const { brand: siteBrand } = useSiteRuntime()
  const homeDefaults = defaultHomeSettings(siteBrand.value)
  const siteDefaults = defaultSiteSettings(siteBrand.value)
  const { data: siteData, pending: sitePending } = useAsyncData(
    'resource-site-settings',
    () => call<{ settings: SiteSettingsView }>('/api/v1/resource/settings'),
    { default: () => ({ settings: siteDefaults }) }
  )
  const { data: homeData, pending: homePending } = useAsyncData(
    'resource-home-settings',
    () => call<{ settings: HomeSettingsView }>('/api/v1/resource/home'),
    { default: () => ({ settings: homeDefaults }) }
  )

  const siteSettings = computed(() => siteData.value?.settings || siteDefaults)
  const homeSettings = computed(() => homeData.value?.settings || homeDefaults)
  const pending = computed(() => sitePending.value || homePending.value)

  return { siteSettings, homeSettings, pending }
}
