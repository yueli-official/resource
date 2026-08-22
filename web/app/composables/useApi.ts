import { useApi as useFoundationApi } from "@yueli/nuxt-runtime/runtime";
import {
  toResourceApiRequest,
  type ResourceApiCallOptions,
} from "../utils/apiCompat";

export function useApi(target = "resource") {
  const request = useFoundationApi(target);

  async function call<T>(
    url: string,
    options?: ResourceApiCallOptions,
  ): Promise<T> {
    const prepared = toResourceApiRequest<T>(url, options);
    return request.request(prepared.path, prepared.options);
  }

  return { call, request };
}
