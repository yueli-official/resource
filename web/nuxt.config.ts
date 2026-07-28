// Nuxt 4 config for the resource-site app (consumer site of the asset service).
// Extends @yueli/identity-nuxt (the OIDC BFF layer): /auth/* + the /api/v1 proxy that
// injects the user's Bearer token come from the layer, so this app has no devProxy.
const siteBrand = process.env.NUXT_PUBLIC_SITE_BRAND || "资源库";

export default defineNuxtConfig({
  // Foundation content layer = the shared rich editor and prose renderer;
  // its nuxt.config contributes the tiptap optimizeDeps + katex css so this app
  // doesn't re-declare them (SP0 red ribbon).
  extends: [
    "@yueli/identity-nuxt",
    "@platform/site",
    "@platform/manage",
    "@yueli/asset-nuxt",
    "@yueli/content-nuxt",
  ],
  modules: ["@nuxt/ui", "@yueli/ui", "@yueli/discovery-nuxt"],
  css: ["~/assets/css/main.css"],
  buildDir: process.env.NUXT_BUILD_DIR || ".nuxt",
  devServer: { port: Number(process.env.NUXT_DEV_PORT || "3001") },
  runtimeConfig: {
    // SSR-only base for direct (anonymous) public-page reads to the resource
    // service. Authenticated client calls instead go through /api/v1 (the BFF
    // proxy from @yueli/identity-nuxt, which injects the Bearer token).
    apiBase: process.env.NUXT_API_BASE || "http://127.0.0.1:8083",
    identityBase:
      process.env.NUXT_IDENTITY_BASE ||
      process.env.NUXT_PUBLIC_OIDC_ISSUER ||
      "http://localhost:8081",
    assetBase: process.env.NUXT_ASSET_BASE || "http://127.0.0.1:8082",
    // OIDC BFF (layer) config — override in prod with NUXT_* env.
    sealSecret:
      process.env.NUXT_SEAL_SECRET ||
      "dev-resource-seal-secret-change-me-0123456789ab",
    downstreamBase: process.env.NUXT_DOWNSTREAM_BASE || "http://127.0.0.1:8083",
    public: {
      oidcIssuer:
        process.env.NUXT_PUBLIC_OIDC_ISSUER || "http://localhost:8081",
      oidcClientId:
        process.env.NUXT_PUBLIC_OIDC_CLIENT_ID || "resource-main-web",
      oidcRedirectUri:
        process.env.NUXT_PUBLIC_OIDC_REDIRECT_URI ||
        "http://localhost:3001/auth/callback",
      oidcScopes:
        process.env.NUXT_PUBLIC_OIDC_SCOPES ||
        "openid profile email roles offline_access",
      accountUrl:
        process.env.NUXT_PUBLIC_ACCOUNT_URL || "http://localhost:3000",
      siteSlug: process.env.NUXT_PUBLIC_SITE_SLUG || "resource-main",
      siteBrand,
      siteDomain:
        process.env.NUXT_PUBLIC_SITE_DOMAIN || "resource-main.localhost",
      assetSpace: process.env.NUXT_PUBLIC_ASSET_SPACE || "default",
      assetNamespace: process.env.NUXT_PUBLIC_ASSET_NAMESPACE || "default",
      assetProfile: process.env.NUXT_PUBLIC_ASSET_PROFILE || "resource-default",
    },
  },
  devtools: { enabled: true },
});
