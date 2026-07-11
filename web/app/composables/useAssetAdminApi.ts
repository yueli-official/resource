const assetAdminPrefix = '/api/v1/admin/assets'

function toAssetAdminProxyPath(url: string) {
  const [rawPath, query] = url.split('?')
  const path = rawPath || ''
  if (!path.startsWith(assetAdminPrefix)) {
    throw createError({ statusCode: 400, statusMessage: `Unsupported asset admin path: ${path}` })
  }
  const suffix = path.slice(assetAdminPrefix.length) || '/stats'
  return `/api/v1/admin/assets-proxy${suffix}${query ? `?${query}` : ''}`
}

export function useAssetAdminApi() {
  async function call<T>(url: string, opts?: Parameters<typeof $fetch>[1]): Promise<T> {
    return await bffFetch<T>('/identity-api', toAssetAdminProxyPath(url), opts)
  }
  return { call }
}
