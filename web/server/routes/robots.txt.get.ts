// robots.txt — allow all + point crawlers at the sitemap. Served from the
// resource domain so the Sitemap line uses the live origin.
export default defineEventHandler((event) => {
  const origin = getRequestURL(event).origin
  setHeader(event, 'content-type', 'text/plain; charset=utf-8')
  return `User-agent: *
Allow: /

Sitemap: ${origin}/sitemap.xml
`
})
