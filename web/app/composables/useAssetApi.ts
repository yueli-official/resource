export function useAssetApi() {
  async function call<T>(url: string, opts?: Parameters<typeof $fetch>[1]): Promise<T> {
    return await bffFetch<T>('/asset-api', url, opts)
  }
  return { call }
}
