import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  getDefaultPrintOptions,
  loadPrintOptions,
  savePrintOptions,
} from './loadingPrintOptions.js'

describe('loadingPrintOptions', () => {
  /** @type {Storage} */
  let storage

  beforeEach(() => {
    storage = {
      data: {},
      getItem(key) {
        return this.data[key] ?? null
      },
      setItem(key, value) {
        this.data[key] = String(value)
      },
      removeItem(key) {
        delete this.data[key]
      },
      clear() {
        this.data = {}
      },
      get length() {
        return Object.keys(this.data).length
      },
      key() {
        return null
      },
    }
  })

  describe('getDefaultPrintOptions', () => {
    it('returns an object with all known keys set to true', () => {
      const opts = getDefaultPrintOptions()
      expect(opts).toEqual({
        carId: true,
        carName: true,
        color: true,
        vin: true,
        paymentStatus: true,
        client: true,
        loadingInfo: true,
        summary: true,
      })
    })

    it('returns a new object each time (no shared reference)', () => {
      const a = getDefaultPrintOptions()
      const b = getDefaultPrintOptions()
      expect(a).not.toBe(b)
      a.carId = false
      expect(b.carId).toBe(true)
    })
  })

  describe('loadPrintOptions', () => {
    it('returns defaults when storage is null', () => {
      const opts = loadPrintOptions(null)
      expect(opts).toEqual(getDefaultPrintOptions())
    })

    it('returns defaults when key is missing', () => {
      const opts = loadPrintOptions(storage)
      expect(opts).toEqual(getDefaultPrintOptions())
    })

    it('returns merged options when valid JSON is stored', () => {
      storage.setItem('loadingPrintOptions', JSON.stringify({ carId: false, client: false }))
      const opts = loadPrintOptions(storage)
      expect(opts.carId).toBe(false)
      expect(opts.client).toBe(false)
      expect(opts.carName).toBe(true)
      expect(opts.loadingInfo).toBe(true)
    })

    it('falls back to defaults for invalid JSON', () => {
      storage.setItem('loadingPrintOptions', 'not json')
      const opts = loadPrintOptions(storage)
      expect(opts).toEqual(getDefaultPrintOptions())
    })

    it('falls back to defaults for non-object parsed value', () => {
      storage.setItem('loadingPrintOptions', 'true')
      const opts = loadPrintOptions(storage)
      expect(opts).toEqual(getDefaultPrintOptions())
    })

    it('ignores unknown keys from stored data', () => {
      storage.setItem(
        'loadingPrintOptions',
        JSON.stringify({ carId: false, unknownKey: true })
      )
      const opts = loadPrintOptions(storage)
      expect(opts).not.toHaveProperty('unknownKey')
      expect(opts.carId).toBe(false)
    })

    it('uses default for keys with non-boolean stored value', () => {
      storage.setItem('loadingPrintOptions', JSON.stringify({ carId: 'yes' }))
      const opts = loadPrintOptions(storage)
      expect(opts.carId).toBe(true)
    })
  })

  describe('savePrintOptions', () => {
    it('writes valid JSON to storage', () => {
      const options = getDefaultPrintOptions()
      options.carId = false
      savePrintOptions(options, storage)
      const raw = storage.getItem('loadingPrintOptions')
      expect(() => JSON.parse(raw)).not.toThrow()
      const parsed = JSON.parse(raw)
      expect(parsed.carId).toBe(false)
      expect(parsed.carName).toBe(true)
    })

    it('only persists known keys', () => {
      savePrintOptions(
        { ...getDefaultPrintOptions(), carId: false, extra: true },
        storage
      )
      const parsed = JSON.parse(storage.getItem('loadingPrintOptions'))
      expect(parsed).not.toHaveProperty('extra')
      expect(parsed.carId).toBe(false)
    })

    it('does not throw when storage is null', () => {
      expect(() => savePrintOptions(getDefaultPrintOptions(), null)).not.toThrow()
    })
  })

  describe('round-trip', () => {
    it('load returns what was saved (for valid options)', () => {
      const options = {
        carId: false,
        carName: true,
        color: false,
        vin: true,
        paymentStatus: false,
        client: true,
        loadingInfo: false,
        summary: true,
      }
      savePrintOptions(options, storage)
      const loaded = loadPrintOptions(storage)
      expect(loaded).toEqual(options)
    })
  })
})
