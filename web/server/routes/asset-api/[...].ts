export default defineEventHandler(async (event) => {
  const path = getRouterParam(event, '_') || ''
  const search = getRequestURL(event).search
  const headers = await sessionAuthHeaders(event)
  return await proxyRequest(event, `${useRuntimeConfig(event).assetBase}/${path}${search}`, { headers })
})
