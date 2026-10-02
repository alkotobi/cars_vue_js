import { describe, it, expect } from 'vitest'
import {
  MAX_MEDIA_BYTES,
  extensionOf,
  mediaTypeFor,
  uniqueSuffix,
  formatFileSize,
  formatDate
} from './carNameMedia'

describe('carNameMedia helpers', () => {
  it('extensionOf handles cases', () => {
    expect(extensionOf('photo.jpg')).toBe('jpg')
    expect(extensionOf('PHOTO.JPG')).toBe('jpg')
    expect(extensionOf('video.MP4')).toBe('mp4')
    expect(extensionOf('file')).toBe('')
    expect(extensionOf('')).toBe('')
    expect(extensionOf(null)).toBe('')
  })

  it('mediaTypeFor classifies', () => {
    expect(mediaTypeFor('jpg')).toBe('photo')
    expect(mediaTypeFor('png')).toBe('photo')
    expect(mediaTypeFor('webp')).toBe('photo')
    expect(mediaTypeFor('mp4')).toBe('video')
    expect(mediaTypeFor('mov')).toBe('video')
    expect(mediaTypeFor('')).toBe('photo')
  })

  it('uniqueSuffix returns a string', () => {
    const a = uniqueSuffix()
    const b = uniqueSuffix()
    expect(typeof a).toBe('string')
    expect(a.length > 0).toBe(true)
    expect(a).not.toBe(b)
  })

  it('formatFileSize formats', () => {
    expect(formatFileSize(0)).toBe('0 B')
    expect(formatFileSize(-1)).toBe('0 B')
    expect(formatFileSize(1024)).toBe('1 KB')
    expect(formatFileSize(1048576)).toBe('1 MB')
    expect(formatFileSize(10 * 1024 * 1024)).toBe('10 MB')
  })

  it('formatDate handles invalid', () => {
    expect(formatDate('')).toBe('-')
    expect(formatDate(null)).toBe('-')
    expect(formatDate('not-a-date')).toBe('not-a-date')
  })

  it('MAX_MEDIA_BYTES is 100 MB', () => {
    expect(MAX_MEDIA_BYTES).toBe(100 * 1024 * 1024)
  })
})
