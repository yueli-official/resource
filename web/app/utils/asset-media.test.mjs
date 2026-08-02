import assert from 'node:assert/strict'
import test from 'node:test'

import { publicAssetMediaUrl } from './asset-media.mjs'

test('builds the canonical named Asset media URL', () => {
  assert.equal(
    publicAssetMediaUrl('019b0000-0000-7000-9000-000000000030', 'display'),
    '/media/31Pj0mXv7cfR5fdZIUvra?format=webp&name=display'
  )
})

test('accepts a public media origin without leaking transform parameters', () => {
  const url = publicAssetMediaUrl(
    '019b0000-0000-7000-9000-000000000030',
    'card',
    'https://img.yueli.dev/'
  )
  assert.equal(url, 'https://img.yueli.dev/media/31Pj0mXv7cfR5fdZIUvra?format=webp&name=card')
  assert.equal(url.includes('_mode='), false)
})
