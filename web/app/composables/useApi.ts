// gokit envelope: code is a namespaced string; success sentinel is "ok".
interface Envelope<T> { code: string; data: T; message: string; traceId: string }

export function useApi() {
  async function call<T>(url: string, opts?: Parameters<typeof $fetch>[1]): Promise<T> {
    // On the server the dev proxy doesn't apply (internal SSR $fetch bypasses
    // it), so call the backend by absolute URL; on the client use a relative URL
    // so the request is same-origin and goes through the proxy.
    const base = import.meta.server ? useRuntimeConfig().apiBase : ''
    const res = await $fetch<Envelope<T>>(base + url, { ...opts })
    if (res.code !== 'ok') {
      throw createError({ statusCode: 400, statusMessage: res.message, data: res })
    }
    return res.data
  }
  return { call }
}
