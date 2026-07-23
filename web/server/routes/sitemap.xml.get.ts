export default defineEventHandler((event) =>
  serveDiscoveryArtifact(event, "sitemap.xml"),
);
