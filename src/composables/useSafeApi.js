// Composable: small wrapper around callApi/uploadFile/getFileUrl with error state.
// Centralises API helper usage without duplicating the raw callApi implementation.
import { ref } from 'vue'
import { useApi } from './useApi'

export function useSafeApi() {
  const { callApi, uploadFile, getFileUrl, error: apiError } = useApi()
  const error = ref('')
  const loading = ref(false)

  const safeApi = async (options = {}) => {
    error.value = ''
    try {
      const result = await callApi(options)
      if (!result.success) {
        error.value = result.error || apiError.value || 'Unknown error'
      }
      return result
    } catch (e) {
      error.value = e?.message || apiError.value || 'Network error'
      return { success: false, error: error.value }
    }
  }

  return {
    safeApi,
    uploadFile,
    getFileUrl,
    error,
    apiError,
    loading
  }
}
