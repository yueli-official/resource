export default defineEventHandler((event) => {
  const name = decodeURIComponent(
    getRequestURL(event).pathname.split("/").pop() ?? "",
  );
  return serveDiscoveryArtifact(event, name);
});
