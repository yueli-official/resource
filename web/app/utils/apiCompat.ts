import type {
  HttpMethod,
  QueryValue,
  RequestOptions,
} from "@yueli/http-runtime";

export interface ResourceApiCallOptions {
  readonly method?: string;
  readonly query?: Readonly<Record<string, QueryValue>>;
  readonly body?: unknown;
  readonly headers?: HeadersInit;
  readonly signal?: AbortSignal;
  readonly timeout?: number;
}

function queryFromUrl(search: string) {
  const query: Record<string, string | string[]> = {};
  for (const [name, value] of new URLSearchParams(search)) {
    const current = query[name];
    if (current === undefined) query[name] = value;
    else if (Array.isArray(current)) query[name] = [...current, value];
    else query[name] = [current, value];
  }
  return query;
}

export function splitResourceApiUrl(url: string) {
  if (url.includes("#")) throw new Error("foundation.request.invalid_path");
  const queryIndex = url.indexOf("?");
  const path = queryIndex === -1 ? url : url.slice(0, queryIndex);
  const search = queryIndex === -1 ? "" : url.slice(queryIndex + 1);
  return {
    path: path as `/${string}`,
    query: queryFromUrl(search),
  };
}

export function toResourceApiRequest<T>(
  url: string,
  options: ResourceApiCallOptions = {},
): { path: `/${string}`; options: RequestOptions<T> } {
  const parsed = splitResourceApiUrl(url);
  const method = String(options.method ?? "GET").toUpperCase() as HttpMethod;
  const headers = options.headers
    ? (Object.fromEntries(
        new Headers(options.headers),
      ) as RequestOptions<T>["headers"])
    : undefined;

  return {
    path: parsed.path,
    options: {
      method,
      query: { ...parsed.query, ...options.query },
      body: options.body as RequestOptions<T>["body"],
      headers,
      signal: options.signal,
      timeoutMs: options.timeout,
    },
  };
}
