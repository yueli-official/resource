interface Envelope<T> { code: string, data: T, message: string, traceId: string }

export async function bffFetch<T>(base: string, url: string, opts?: Parameters<typeof $fetch>[1]): Promise<T> {
  const nextOpts = { ...opts }
  if (import.meta.server) {
    nextOpts.headers = {
      ...useRequestHeaders(['cookie']),
      ...(opts?.headers as Record<string, string> | undefined)
    }
  }

  try {
    const res = await $fetch<Envelope<T>>(base + url, nextOpts)
    if (res.code !== 'ok') {
      throw createError({ statusCode: 400, statusMessage: res.message, data: res })
    }
    return res.data
  } catch (error: any) {
    const status = error?.statusCode || error?.response?.status
    const data = error?.data || error?.response?._data
    const message = data?.message || error?.statusMessage || error?.message
    if (status === 401) {
      if (import.meta.client) {
        const { login } = useAuth()
        await login()
      }
      throw createError({ statusCode: 401, statusMessage: 'Unauthorized', data: error?.data })
    }
    if (status) {
      throw createError({ statusCode: status, statusMessage: message, data })
    }
    throw error
  }
}
