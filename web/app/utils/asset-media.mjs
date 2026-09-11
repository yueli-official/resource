const BASE62_ALPHABET = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'
const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
const RENDITION_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/

/**
 * Temporary consumer adapter for the @yueli/asset-nuxt/media contract.
 * Replace this file with the package export after its next release.
 *
 * @param {string} assetId
 * @param {string} rendition
 * @param {string} [base]
 */
export function publicAssetMediaUrl(assetId, rendition, base = '') {
  if (!UUID_RE.test(assetId)) return ''
  const name = rendition.toLowerCase()
  if (!RENDITION_RE.test(name)) throw new TypeError('Invalid media rendition')
  if (base.includes('?') || base.includes('#')) throw new TypeError('Invalid media base URL')

  let value = BigInt(`0x${assetId.replaceAll('-', '')}`)
  let key = value === 0n ? '0' : ''
  while (value > 0n) {
    key = BASE62_ALPHABET[Number(value % 62n)] + key
    value /= 62n
  }
  return `${base.replace(/\/+$/, '')}/media/${key}?format=webp&preset=${name}&v=1`
}
