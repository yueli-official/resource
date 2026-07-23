export default defineEventHandler((event) =>
  serveDiscoveryArtifact(event, "robots.txt"),
);
