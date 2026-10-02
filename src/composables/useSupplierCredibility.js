import { computed, ref } from 'vue'
import { useApi } from './useApi'
import { useSubmitGuard } from './useSubmitGuard'
import {
  ASSESS_ACTION,
  HISTORY_ACTION,
  LATEST_ACTION,
  assessKey,
  credibilityErrorText,
  latestChecksBySupplier,
  normalizeCheck,
  normalizeChecks,
  reportLanguage,
} from '../lib/supplierCredibility'

/**
 * Supplier credibility state: which supplier has been checked, the checks
 * themselves, and the errors worth showing.
 *
 * The three API calls behind it are admin-only and need a real token
 * (requiresAuth: true), so a signed-in non-admin gets a translated
 * "admin only" message instead of an empty table.
 *
 * A check costs a model call, which costs money, so every run goes through
 * useSubmitGuard under a per-supplier key: double-clicking one row cannot start
 * two, and two different rows can still run side by side.
 *
 * @param {(key: string) => string} t
 * @param {{value: string}} locale vue-i18n's locale ref, so the report is
 *        written in the language the reader is using
 */
export const useSupplierCredibility = (t, locale) => {
  const { callApi } = useApi()
  const { guard, isBusy } = useSubmitGuard({ log: true })

  const checksBySupplier = ref({})
  const history = ref([])
  const error = ref(null)
  // Kept alongside the message: the modal offers a retry for a busy model but
  // not for a rejected key, and only the code can tell those apart.
  const errorCode = ref(null)
  const isLoadingLatest = ref(false)
  const isLoadingHistory = ref(false)
  const activeSupplierId = ref(null)

  const hasHistory = computed(() => history.value.length > 0)

  /**
   * One call for the whole table: the latest check per supplier, for the badge.
   * A missing table or column is not worth an error banner - the feature is
   * simply not available yet on this server.
   */
  const loadLatest = async () => {
    isLoadingLatest.value = true
    try {
      const result = await callApi({ action: LATEST_ACTION, requiresAuth: true })
      if (result?.success) {
        checksBySupplier.value = latestChecksBySupplier(result.checks)
        return true
      }
      errorCode.value = result?.code ?? 'load_failed'
      error.value = credibilityErrorText(t, result, 'supplierCredibility.errors.loadFailed')
      return false
    } catch (err) {
      console.error('Failed to load supplier credibility scores', err)
      errorCode.value = 'load_failed'
      error.value = t('supplierCredibility.errors.loadFailed')
      return false
    } finally {
      isLoadingLatest.value = false
    }
  }

  /** Past checks for one supplier, newest first. */
  const loadHistory = async (supplierId) => {
    isLoadingHistory.value = true
    activeSupplierId.value = supplierId
    try {
      const result = await callApi({
        action: HISTORY_ACTION,
        supplier_id: supplierId,
        limit: 20,
        requiresAuth: true,
      })
      if (result?.success) {
        history.value = normalizeChecks(result.checks)
        return true
      }
      errorCode.value = result?.code ?? 'history_failed'
      error.value = credibilityErrorText(t, result, 'supplierCredibility.errors.historyFailed')
      return false
    } catch (err) {
      console.error('Failed to load supplier credibility history', err)
      errorCode.value = 'history_failed'
      error.value = t('supplierCredibility.errors.historyFailed')
      return false
    } finally {
      isLoadingHistory.value = false
    }
  }

  /**
   * Run a new check. Admin only; the server refuses anyone else.
   *
   * @returns {Promise<object|null>} the stored check, or null when it failed
   */
  const assess = async (supplier) => {
    error.value = null
    errorCode.value = null
    // A per-call closure so each supplier gets its own guard key.
    const run = guard(assessKey(supplier.id), async () => {
      try {
        const result = await callApi({
          action: ASSESS_ACTION,
          supplier_id: supplier.id,
          lang: reportLanguage(locale?.value),
          requiresAuth: true,
        })
        if (!result?.success) {
          errorCode.value = result?.code ?? 'assess_failed'
          error.value = credibilityErrorText(t, result, 'supplierCredibility.errors.assessFailed')
          return null
        }

        const stored = normalizeCheck(result.check)
        history.value = [stored, ...history.value.filter((row) => row.id !== stored.id)]
        checksBySupplier.value = {
          ...checksBySupplier.value,
          [stored.supplier_id]: stored,
        }
        return stored
      } catch (err) {
        console.error('Supplier credibility check failed', err)
        errorCode.value = 'assess_failed'
        error.value = t('supplierCredibility.errors.assessFailed')
        return null
      }
    })

    return run()
  }

  return {
    checksBySupplier,
    history,
    hasHistory,
    error,
    errorCode,
    isLoadingLatest,
    isLoadingHistory,
    activeSupplierId,
    isAssessing: (supplierId) => isBusy(assessKey(supplierId)),
    isAnyAssessing: computed(() => isBusy()),
    loadLatest,
    loadHistory,
    assess,
    clearError: () => {
      error.value = null
      errorCode.value = null
    },
  }
}
