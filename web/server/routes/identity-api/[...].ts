export default defineEventHandler(async (event) => {
  const path = getRouterParam(event, '_') || ''
  const search = getRequestURL(event).search
  const headers = await sessionAuthHeaders(event)
  const cookie = getHeader(event, 'cookie')
  if (cookie) headers.cookie = cookie
  return await proxyRequest(event, `${useRuntimeConfig(event).identityBase}/${path}${search}`, { headers })
})
