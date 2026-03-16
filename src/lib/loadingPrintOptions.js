/**
 * Helper functions for loading record print options.
 * Stores the user's column/section choices in localStorage and merges with defaults.
 */

const STORAGE_KEY = 'loadingPrintOptions'

/**
 * Default print options: all columns/sections enabled.
 * @returns {{carId:boolean,carName:boolean,color:boolean,vin:boolean,paymentStatus:boolean,client:boolean,loadingInfo:boolean,summary:boolean}}
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
 * Load print options from storage, merged with defaults.
 * Unknown keys and non-boolean values are ignored.
 * @param {Storage|null} [storage=typeof window !== 'undefined' ? window.localStorage : null]
 * @returns {Record<string, boolean>}
 */
export function loadPrintOptions(storage = typeof window !== 'undefined' ? window.localStorage : null) {
  const defaults = getDefaultPrintOptions()
  if (!storage) return { ...defaults }

  try {
    const raw = storage.getItem(STORAGE_KEY)
    if (!raw) return { ...defaults }

    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object') return { ...defaults }

    const merged = { ...defaults }
    Object.keys(defaults).forEach((key) => {
      if (Object.prototype.hasOwnProperty.call(parsed, key) && typeof parsed[key] === 'boolean') {
        merged[key] = parsed[key]
      }
    })
    return merged
  } catch {
    return { ...defaults }
  }
}

/**
 * Save print options to storage. Only known keys are persisted as booleans.
 * @param {Record<string, boolean>} options
 * @param {Storage|null} [storage=typeof window !== 'undefined' ? window.localStorage : null]
 */
export function savePrintOptions(options, storage = typeof window !== 'undefined' ? window.localStorage : null) {
  if (!storage) return
  try {
    const defaults = getDefaultPrintOptions()
    const toSave = {}
    Object.keys(defaults).forEach((key) => {
      toSave[key] = options && options[key] === false ? false : true
    })
    storage.setItem(STORAGE_KEY, JSON.stringify(toSave))
  } catch {
    // ignore storage errors
  }
}

