import { ref } from 'vue'

const API_URL = 'http://localhost:8000/api.php'

export const useApi = () => {
  const error = ref(null)
  const loading = ref(false)

  const callApi = async (data) => {
    loading.value = true
    error.value = null

    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
      })

      const result = await response.json()
      return result
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    callApi,
    error,
    loading
  }
}