export const MAX_MEDIA_BYTES = 100 * 1024 * 1024 // 100 MB

export const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg']

export const VIDEO_EXTENSIONS = ['mp4', 'webm', 'avi', 'mov']

export function extensionOf(name) {
  if (!name || typeof name !== 'string') return ''
  const idx = name.lastIndexOf('.')
  if (idx < 0) return ''
  return name.slice(idx + 1).toLowerCase()
}

export function mediaTypeFor(ext) {
  if (!ext) return 'photo'
  if (IMAGE_EXTENSIONS.includes(ext)) return 'photo'
  if (VIDEO_EXTENSIONS.includes(ext)) return 'video'
  return 'photo'
}

export function uniqueSuffix() {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`
}

export function formatFileSize(bytes) {
  if (typeof bytes !== 'number' || Number.isNaN(bytes) || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const unit = units[Math.min(i, units.length - 1)]
  const size = bytes / 1024 ** Math.min(i, units.length - 1)
  const rounded = Math.abs(size - Math.round(size)) < 1e-9 ? Math.round(size) : (size < 10 ? size.toFixed(1) : Math.round(size))
  return `${rounded} ${unit}`
}

export function formatDate(dateStr) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return d.toLocaleString()
}
