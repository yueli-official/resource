// XML sitemap for crawlers: home + every published resource detail page. Served
// from the resource domain (Nitro), pulling public resources from the resource
// service. SEO infra absorbed from the old monolith (mirrors products/blog/web) — a
// built-in feature with sane defaults, not a runtime-config knob.
function esc(s = ''): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}
function lastmod(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '' : d.toISOString().slice(0, 10)
}

export default defineEventHandler(async (event) => {
  const apiBase = useRuntimeConfig(event).apiBase
  const origin = getRequestURL(event).origin
  const items = await $fetch<{ data?: { items?: any[] } }>(`${apiBase}/api/v1/resources`, { query: { page: 1, size: 500 } })
    .then(r => r?.data?.items ?? []).catch(() => [])

  const urls: string[] = []
  const add = (loc: string, mod?: string) =>
    urls.push(`  <url><loc>${esc(loc)}</loc>${mod ? `<lastmod>${mod}</lastmod>` : ''}</url>`)

  add(`${origin}/`)
  for (const r of items) add(`${origin}/resources/${r.id}`, lastmod(r.updatedAt || r.createdAt))

  setHeader(event, 'content-type', 'application/xml; charset=utf-8')
  return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls.join('\n')}
</urlset>
`
})
