import { ref, onMounted } from 'vue'
import { useApi } from './useApi'

export const useVersionCheck = () => {
  const { callApi } = useApi()
  const currentAppVersion = ref(5) // Update this when you make breaking changes
  const dbVersion = ref(null)
  const isLoading = ref(true)
  const hasVersionMismatch = ref(false)

  const checkVersion = async () => {
    try {
      isLoading.value = true

      const result = await callApi({
        query: 'SELECT version FROM versions ORDER BY id DESC LIMIT 1',
        params: [],
        requiresAuth: true,
      })

      if (result.success && result.data.length > 0) {
        dbVersion.value = result.data[0].version

        // Compare versions
        if (dbVersion.value !== currentAppVersion.value) {
          hasVersionMismatch.value = true
          return false
        }

        hasVersionMismatch.value = false
        return true
      } else {
        console.error('Failed to fetch version from database')
        return false
      }
    } catch (error) {
      console.error('Version check error:', error)
      return false
    } finally {
      isLoading.value = false
    }
  }

  const forceRefresh = () => {
    window.location.reload()
  }

  return {
    currentAppVersion,
    dbVersion,
    isLoading,
    hasVersionMismatch,
    checkVersion,
    forceRefresh,
  }
}
