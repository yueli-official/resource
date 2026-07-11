export function useSiteRuntime() {
  const config = useRuntimeConfig()
  return {
    slug: computed(() => config.public.siteSlug || 'resource-ae'),
    brand: computed(() => config.public.siteBrand || '资源库'),
    domain: computed(() => config.public.siteDomain || 'resource-ae.localhost'),
    assetSpace: computed(() => config.public.assetSpace || 'default'),
    assetNamespace: computed(() => config.public.assetNamespace || 'default'),
    assetProfile: computed(() => config.public.assetProfile || 'resource-default')
  }
}
