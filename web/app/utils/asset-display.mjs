const ARCHIVE_EXTS = new Set(['zip', '7z', 'rar', 'tar', 'gz'])
const DESIGN_EXTS = new Set(['psd', 'ai', 'fig', 'sketch', 'xd', 'blend'])
const VIDEO_EXTS = new Set(['mp4', 'mov', 'webm', 'avi', 'mkv'])
const IMAGE_EXTS = new Set(['jpg', 'jpeg', 'png', 'webp', 'gif', 'avif', 'svg'])

export function assetExtension(filename = '') {
  const match = String(filename).toLowerCase().match(/\.([a-z0-9]+)$/)
  return match?.[1] || ''
}

export function assetFileIcon(filename = '', mime = '') {
  const ext = assetExtension(filename)
  if (String(mime).startsWith('image/') || IMAGE_EXTS.has(ext)) return 'i-tabler-photo'
  if (ARCHIVE_EXTS.has(ext)) return 'i-tabler-file-zip'
  if (DESIGN_EXTS.has(ext)) return 'i-tabler-file-pencil'
  if (ext === 'pdf') return 'i-tabler-file-type-pdf'
  if (VIDEO_EXTS.has(ext)) return 'i-tabler-file-video'
  if (['txt', 'md', 'doc', 'docx'].includes(ext)) return 'i-tabler-file-text'
  return 'i-tabler-file'
}

export function assetFileTone(filename = '', mime = '') {
  const ext = assetExtension(filename)
  if (String(mime).startsWith('image/') || IMAGE_EXTS.has(ext)) return 'text-success bg-success/10'
  if (ARCHIVE_EXTS.has(ext)) return 'text-warning bg-warning/10'
  if (DESIGN_EXTS.has(ext)) return 'text-primary bg-primary/10'
  if (ext === 'pdf') return 'text-error bg-error/10'
  if (VIDEO_EXTS.has(ext)) return 'text-info bg-info/10'
  return 'text-muted bg-[var(--resource-surface-card)]'
}

export function formatAssetSize(size = 0) {
  const bytes = Number(size) || 0
  if (bytes < 1024) return `${Math.max(0, bytes)} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let value = bytes / 1024
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  const rounded = value >= 10 || Number.isInteger(value) ? Math.round(value).toString() : value.toFixed(1)
  return `${rounded} ${units[unitIndex]}`
}
