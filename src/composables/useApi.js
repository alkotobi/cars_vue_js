import { ref } from 'vue'

const API_BASE_URL = 'http://localhost:8000'
// const API_BASE_URL = 'https://www.merhab.com/api'
const API_URL = `${API_BASE_URL}/api.php`
const UPLOAD_URL = `${API_BASE_URL}/upload.php`

export const useApi = () => {
  const error = ref(null)
  const loading = ref(false)

  // Original API call function
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

  // File upload function
  const uploadFile = async (file, destinationFolder = 'uploads', customFilename = '') => {
    loading.value = true
    error.value = null

    try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('destination_folder', destinationFolder)
      if (customFilename) {
        formData.append('custom_filename', customFilename)
      }

      const response = await fetch(UPLOAD_URL, {
        method: 'POST',
        body: formData
      })

      const result = await response.json()
      
      if (!result.success) {
        throw new Error(result.message || 'Upload failed')
      }

      // Return both the server response and the relative path
      return {
        ...result,
        relativePath: `${destinationFolder}/${customFilename}`
      }
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  // Get file URL helper
  const getFileUrl = (path) => {
    // If path is already a full URL, return it as is
    if (path.startsWith('http')) {
      return path
    }
    
    // Remove any leading slashes from the path and store in a new variable
    let processedPath = path.replace(/^\/+/, '')
    
    // If the path already contains 'upload.php?path=', extract just the file path
    if (processedPath.includes('upload.php?path=')) {
      const match = processedPath.match(/path=(.+)$/)
      if (match) {
        processedPath = decodeURIComponent(match[1])
      }
    }
    
    return `${UPLOAD_URL}?path=${encodeURIComponent(processedPath)}`
  }

  return {
    callApi,
    uploadFile,
    getFileUrl,
    error,
    loading
  }
}