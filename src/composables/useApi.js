import { ref } from 'vue'

// Get the current hostname
const hostname = window.location.hostname
const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'

// Set API base URL based on environment
const API_BASE_URL = isLocalhost ? 'http://localhost:8000/api' : 'https://www.merhab.com/api'

const API_URL = `${API_BASE_URL}/api.php`
const UPLOAD_URL = `${API_BASE_URL}/upload.php`

// const UPLOAD_URL = `${API_BASE_URL}/upload_simple.php`  // Use this if main upload fails

export const useApi = () => {
  const error = ref(null)
  const loading = ref(false)

  // Original API call function
  const callApi = async (data) => {
    loading.value = true
    error.value = null

    try {
      // Get user token if not a public route
      const user = !data.requiresAuth ? null : localStorage.getItem('user')
      const userData = user ? JSON.parse(user) : null

      const requestData = {
        ...data,
        token: userData?.token,
      }

      const response = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData),
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
      console.log('Starting file upload:', {
        fileName: file.name,
        fileSize: file.size,
        fileType: file.type,
        destinationFolder,
        customFilename,
        uploadUrl: UPLOAD_URL,
      })

      const formData = new FormData()
      formData.append('file', file)
      formData.append('destination_folder', destinationFolder)
      if (customFilename) {
        formData.append('custom_filename', customFilename)
      }

      // Add timeout to prevent hanging
      const controller = new AbortController()
      const timeoutId = setTimeout(() => {
        console.log('Upload timeout - aborting request')
        controller.abort()
      }, 30000) // 30 second timeout

      const response = await fetch(UPLOAD_URL, {
        method: 'POST',
        body: formData,
        signal: controller.signal,
      })

      clearTimeout(timeoutId)
      console.log('Upload response status:', response.status, response.statusText)

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status} - ${response.statusText}`)
      }

      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text()
        console.error('Non-JSON response:', text)
        throw new Error(`Server returned non-JSON response: ${text.substring(0, 200)}`)
      }

      const result = await response.json()
      console.log('Upload result:', result)

      if (!result.success) {
        throw new Error(result.message || 'Upload failed')
      }

      // Return both the server response and the relative path
      let relativePath
      if (customFilename) {
        relativePath = `${destinationFolder}/${customFilename}`
      } else {
        // Extract the path from the server response
        // Server returns: /api/upload.php?path=ids/filename.jpg
        const pathMatch = result.file_path.match(/path=([^&]+)/)
        if (pathMatch) {
          relativePath = decodeURIComponent(pathMatch[1])
        } else {
          // Fallback to extracting filename from the end
          relativePath = `${destinationFolder}/${result.file_path.split('/').pop()}`
        }
      }

      return {
        ...result,
        relativePath,
      }
    } catch (err) {
      console.error('Upload error details:', err)

      // Handle specific error types
      if (err.name === 'AbortError') {
        error.value = 'Upload timeout - please try again'
        throw new Error('Upload timeout - please try again')
      } else if (err.message.includes('Failed to fetch')) {
        error.value = 'Network error - please check your connection'
        throw new Error('Network error - please check your connection')
      } else {
        error.value = err.message
        throw err
      }
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

    // If the path already contains 'upload.php?path=' or 'upload_simple.php?path=', extract just the file path
    if (
      processedPath.includes('upload.php?path=') ||
      processedPath.includes('upload_simple.php?path=')
    ) {
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
    loading,
  }
}
