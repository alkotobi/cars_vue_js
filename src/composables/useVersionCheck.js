import { ref, onMounted } from 'vue'
import { useApi } from './useApi'

export const useVersionCheck = () => {
  const { getDbVersion } = useApi()
  const currentAppVersion = ref(25) // Update this when you make breaking changes
  const dbVersion = ref(null)
  const isLoading = ref(true)
  const hasVersionMismatch = ref(false)

  const checkVersion = async () => {
    try {
      isLoading.value = true

      // A named action, not the SQL passthrough. DatabaseVersionCheck is mounted
      // unconditionally in App.vue, so this runs on /login before anyone has
      // logged in and has to keep working with no token - the action is public for
      // that reason and returns a single version integer.
      const version = await getDbVersion()

      if (version !== null && version !== undefined) {
        dbVersion.value = version

        // Compare versions
        console.log(
          `[Version Check] Database Version: ${dbVersion.value}, Application Version: ${currentAppVersion.value}`,
        )
        if (dbVersion.value !== currentAppVersion.value) {
          console.warn(
            `[Version Check] Version mismatch detected! DB: ${dbVersion.value}, App: ${currentAppVersion.value}`,
          )
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
