/**
 * Loading print options: defaults, localStorage load/save.
 * Used by the loading record print dialog to choose which columns/sections to include.
 * All functions are pure except load/save which use localStorage (tested via mocks).
 */

const STORAGE_KEY = 'loadingPrintOptions'

/**
 * Default print options: all columns and sections enabled.
 * Keys match the options used by generatePrintContent in loadingPrintContent.js.
 * @returns {Record<string, boolean>} Object with keys carId, carName, color, vin, paymentStatus, client, loadingInfo, summary
 */
export function getDefaultPrintOptions() {
  return {
    carId: true,
    carName: true,
    color: true,
    vin: true,
    paymentStatus: true,
    client: true,
    loadingInfo: true,
    summary: true,
  }
}

/**
 * Load print options from localStorage, merged with defaults.
 * Missing or invalid stored data returns defaults. New keys in defaults (e.g. after app update) get default true.
 * @param {Storage} [storage=window.localStorage] - Storage instance (for tests).
 * @returns {Record<string, boolean>} Merged options object; never null/undefined.
 */
export function loadPrintOptions(storage = typeof window !== 'undefined' ? window.localStorage : null) {
  const defaults = getDefaultPrintOptions()
  if (!storage) return { ...defaults }

  try {
    const raw = storage.getItem(STORAGE_KEY)
    if (raw == null || raw === '') return { ...defaults }

    const parsed = JSON.parse(raw)
    if (typeof parsed !== 'object' || parsed === null) return { ...defaults }

    const merged = { ...defaults }
    for (const key of Object.keys(defaults)) {
      if (Object.prototype.hasOwnProperty.call(parsed, key) && typeof parsed[key] === 'boolean') {
        merged[key] = parsed[key]
      }
    }
    return merged
  } catch {
    return { ...defaults }
  }
}

/**
 * Save print options to localStorage. Overwrites the entire stored object.
 * @param {Record<string, boolean>} options - Options object to persist.
 * @param {Storage} [storage=window.localStorage] - Storage instance (for tests).
 */
export function savePrintOptions(options, storage = typeof window !== 'undefined' ? window.localStorage : null) {
  if (!storage) return
  try {
    const defaults = getDefaultPrintOptions()
    const toSave = {}
    for (const key of Object.keys(defaults)) {
      toSave[key] = options[key] === true
    }
    storage.setItem(STORAGE_KEY, JSON.stringify(toSave))
  } catch {
    // Ignore quota or other storage errors
  }
}
