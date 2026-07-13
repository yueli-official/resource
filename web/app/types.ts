// Outward contracts mirrored from products/resource/api/api/v1.
export interface ResourceView {
  id: string
  ownerId: string
  title: string
  slug: string
  summary: string
  description: string
  type: string
  coverAssetId?: string
  coverUrl?: string
  deliveryKind?: string
  deliveryPayload?: DeliveryPayloadView
  status: string
  publishedAt?: string
  viewCount: number
  downloadCount: number
  issueCount: number
  tags: string[]
  createdAt: string
  updatedAt: string
}

export interface AssetView {
  id: string
  visibility: string
  filename: string
  mime: string
  size: number
  width?: number
  height?: number
  altText: string
  title: string
  category: string
  siteKey: string
  profileKey: string
  deliveryPolicy: string
  cdnUrl?: string
  createdAt: string
}

export interface SettingsLinkView {
  label: string
  to: string
  icon: string
}

export interface FooterLinkGroupView {
  title: string
  links: SettingsLinkView[]
}

export interface HomeHighlightView {
  icon: string
  title: string
  text: string
}

export interface HomeSettingsView {
  heroTitle: string
  heroSubtitle: string
  introTitle: string
  introBody: string
  introHighlights: HomeHighlightView[]
  featuredResourceIds: string[]
  categorySlugs: string[]
  quickLinks: SettingsLinkView[]
}

export interface SiteSettingsView {
  site: {
    siteName: string
    tagline: string
    logoIcon: string
    announcement: string
    announcementEnabled: boolean
    supportEmail: string
  }
  footer: {
    tagline: string
    copyright: string
    compliance: {
      icpRecord: string
      icpUrl: string
      policeRecord: string
      policeUrl: string
      extraText: string
    }
    linkGroups: FooterLinkGroupView[]
    socialLinks: SettingsLinkView[]
  }
  resource: {
    resourcesPerPage: number
    downloadsEnabled: boolean
    largeFileThresholdMB: number
    largeFileHint: string
  }
}

export interface ResourceAssetView {
  assetId: string
  label: string
  mime: string
  filename: string
  size: number
  sort: number
}

export interface NetdiskDeliveryView {
  provider?: string
  url?: string
  accessCode?: string
  extractCode?: string
  note?: string
}

export interface DeliveryItemView {
  id?: string
  kind: 'asset_file' | 'netdisk' | string
  title?: string
  assetId?: string
  netdisk?: NetdiskDeliveryView
  sort?: number
  enabled: boolean
  required?: boolean
}

export interface DeliveryPayloadView {
  items?: DeliveryItemView[]
}

// TaxonomyView mirrors products/resource/api/api/v1.TaxonomyView (category | tag).
export interface TaxonomyView {
  id: string
  taxonomy: string // 'category' | 'tag'
  name: string
  slug: string
  description?: string
  parentId?: string
  count: number // published resources under this taxonomy
}

export interface ListTaxonomies {
  items: TaxonomyView[]
  total?: number
  page?: number
  size?: number
}

// SEOView mirrors products/resource/api/api/v1.SEOView.
export interface SEOView {
  metaTitle: string
  metaDesc: string
  ogTitle: string
  ogImage: string
  canonicalUrl: string
  robots: string
}

export interface ListResources {
  items: ResourceView[]
  total: number
  page: number
  size: number
}

export interface ResourceDetail {
  resource: ResourceView
  assets: ResourceAssetView[]
  taxonomies: TaxonomyView[]
  seo?: SEOView | null
}

// MyResources is the operator's own catalog (any status), distinct from the
// public ListResources (published only).
export interface MyResources {
  items: ResourceView[]
  total: number
  page: number
  size: number
  counts: ResourceLifecycleCounts
}

export interface ResourceLifecycleCounts {
  all: number
  published: number
  draft: number
  archived: number
  issues: number
}
