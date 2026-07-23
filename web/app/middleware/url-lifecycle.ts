import type { URLResolutionResponse } from "~/types";

export default defineNuxtRouteMiddleware(async (to) => {
  const { call } = useApi();
  const result = await call<URLResolutionResponse>(
    "/api/v1/url-lifecycle/resolve",
    {
      query: {
        path: to.path,
        query: to.fullPath.split("?")[1]?.split("#")[0] || "",
      },
    },
  ).catch(() => null);
  const resolution = result?.resolution;

  if (
    resolution?.location &&
    (resolution.kind === "redirect" || resolution.kind === "alias")
  ) {
    return navigateTo(resolution.location, {
      redirectCode: resolution.statusCode || 308,
      replace: true,
    });
  }
  if (resolution?.kind === "gone") {
    throw createError({
      statusCode: 410,
      statusMessage: "页面已删除",
      fatal: true,
    });
  }
});
