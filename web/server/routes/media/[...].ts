import { createBffHandler } from "@yueli/nuxt-runtime/server";

// Temporary local copy of the @yueli/asset-nuxt public media route. Remove it
// after the package release containing the shared `/media` route is adopted.
export default createBffHandler({
  mountPath: "/media",
  profile: "asset",
  timeoutMs: 60_000,
  resolveTarget({ event }) {
    const target = assetBffTarget(String(useRuntimeConfig(event).assetBase));
    return {
      origin: target.origin,
      pathPrefix: `${target.pathPrefix || ""}/api/v1/media`,
    };
  },
});
