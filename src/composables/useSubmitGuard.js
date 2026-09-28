import { reactive, computed, getCurrentScope, onScopeDispose } from 'vue'

/**
 * Per-component registry of in-flight submit handlers.
 *
 * Use when a handler writes to the database and has no busy flag of its own:
 *
 *   const { guard, isBusy } = useSubmitGuard()
 *   const savePayment = guard('save-payment', async (form) => { ... })
 *
 *   <button :disabled="isBusy('save-payment')" @click="savePayment(form)">
 *
 * The wrapped function is a no-op while the same key is already running, which
 * covers fast clicks and repeated Enter presses. Distinct keys run in parallel,
 * so `isBusy(id)` on a per-row delete disables every row while one is in flight
 * but leaves the others clickable.
 *
 * @param {object} [options]
 * @param {boolean} [options.log] warn when a submit is dropped as a duplicate
 * @returns {{
 *   guard: (key: string|number, fn: Function) => Function,
 *   isBusy: (key?: string|number) => boolean,
 *   isAnyBusy: import('vue').ComputedRef<boolean>,
 *   busyKeys: import('vue').ComputedRef<Array<string|number>>,
 *   isIdle: (key?: string|number) => boolean,
 *   reset: (key?: string|number) => void
 * }}
 */
export const useSubmitGuard = ({ log = false } = {}) => {
  const inFlight = reactive(new Set())

  const isBusy = (key) => (key === undefined ? inFlight.size > 0 : inFlight.has(key))
  const isIdle = (key) => !isBusy(key)
  const isAnyBusy = computed(() => inFlight.size > 0)
  const busyKeys = computed(() => Array.from(inFlight))

  const reset = (key) => {
    if (key === undefined) inFlight.clear()
    else inFlight.delete(key)
  }

  const guard = (key, fn) => {
    if (typeof fn !== 'function') {
      throw new TypeError(`useSubmitGuard: expected a function for key "${key}"`)
    }

    const wrapped = async (...args) => {
      if (inFlight.has(key)) {
        if (log) console.warn(`[useSubmitGuard] dropped duplicate submit for "${key}"`)
        return
      }

      inFlight.add(key)
      try {
        return await fn(...args)
      } finally {
        inFlight.delete(key)
      }
    }

    wrapped.guardKey = key
    return wrapped
  }

  if (getCurrentScope()) {
    onScopeDispose(() => inFlight.clear())
  }

  return { guard, isBusy, isIdle, isAnyBusy, busyKeys, reset }
}
