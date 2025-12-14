import { ref } from 'vue'

// Get the current hostname and protocol
const hostname = window.location.hostname
const protocol = window.location.protocol
const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'

// Set API base URL based on environment
// Use localhost API for development, production API for production
const API_BASE_URL = isLocalhost ? 'http://localhost:8000/api' : 'https://www.merhab.com/api'

const API_URL = `${API_BASE_URL}/api.php`
const UPLOAD_URL = `${API_BASE_URL}/upload.php`
const DB_MANAGER_API_URL = `${API_BASE_URL}/db_manager_api.php`

// Configuration cache (loaded from db_code.json)
let config_cache = null
let config_promise = null

// Function to load configuration from db_code.json
// This file is in the public folder and will be accessible at /db_code.json when distributed
async function loadConfig() {
  // Return cached value if already loaded
  if (config_cache !== null) {
    return config_cache
  }

  // Return existing promise if already loading
  if (config_promise) {
    return config_promise
  }

  // Start loading
  config_promise = (async () => {
    try {
      const response = await fetch('/db_code.json')
      if (!response.ok) {
        throw new Error(`Failed to load db_code.json: ${response.status}`)
      }
      const data = await response.json()

      // Check if db_code exists
      if (data.db_code) {
        // Fetch database info from dbs table using db_code
        try {
          const dbResponse = await fetch(
            `${DB_MANAGER_API_URL}?action=get_database_by_code&db_code=${encodeURIComponent(data.db_code)}`,
          )
          if (!dbResponse.ok) {
            throw new Error(`Failed to fetch database info: ${dbResponse.status}`)
          }
          const dbResult = await dbResponse.json()

          if (dbResult.success && dbResult.data) {
            // Use db_name and files_dir from dbs table
            if (!dbResult.data.db_name) {
              throw new Error(
                `Database record found but db_name is missing for db_code: ${data.db_code}`,
              )
            }
            config_cache = {
              db_name: dbResult.data.db_name,
              upload_path: dbResult.data.files_dir || '', // Use files_dir from dbs table (can be null/empty)
            }
            return config_cache
          } else {
            throw new Error(`Database not found for db_code: ${data.db_code}`)
          }
        } catch (dbErr) {
          console.error('Error fetching database from dbs table:', dbErr)
          throw dbErr
        }
      } else {
        // No db_code, use db_name and uplod_path from JSON
        if (!data.db_name || !data.uplod_path) {
          throw new Error(
            'Missing required configuration: db_name and uplod_path are required when db_code is not provided',
          )
        }

        config_cache = {
          db_name: data.db_name,
          upload_path: data.uplod_path,
        }
        return config_cache
      }
    } catch (err) {
      console.error('Fatal error loading configuration:', err)
      // Fatal error - stop the website
      throw new Error(
        `FATAL ERROR: Cannot load database configuration. ${err.message}. Please check db_code.json file.`,
      )
    } finally {
      config_promise = null
    }
  })()

  return config_promise
}

// Helper function to get database name
async function loadDbName() {
  const config = await loadConfig()
  return config.db_name
}

// Helper function to get upload path
async function loadUploadPath() {
  const config = await loadConfig()
  return config.upload_path
}

// Debug logging
// console.log('API Configuration:', {
//   hostname,
//   protocol,
//   isLocalhost,
//   API_BASE_URL,
//   API_URL,
//   UPLOAD_URL,
// })

// const UPLOAD_URL = `${API_BASE_URL}/upload_simple.php`  // Use this if main upload fails

export const useApi = () => {
  const error = ref(null)
  const loading = ref(false)

  // Original API call function
  const callApi = async (data, retryCount = 0) => {
    loading.value = true
    error.value = null

    try {
      // Add a small random delay to make requests look more human-like (only on production)
      if (retryCount === 0 && !isLocalhost) {
        const delay = Math.random() * 1000 + 500 // 500-1500ms delay
        await new Promise((resolve) => setTimeout(resolve, delay))
      }

      // Get user token if not a public route
      const user = !data.requiresAuth ? null : localStorage.getItem('user')
      const userData = user ? JSON.parse(user) : null

      // Load database name from db_code.json
      const db_name = await loadDbName()

      const requestData = {
        ...data,
        token: userData?.token,
        dbname: db_name,
      }

      // Removed: console.log('API call to:', API_URL, 'with data:', requestData)

      const response = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(isLocalhost
            ? {
                // Simple headers for localhost
                Accept: 'application/json',
              }
            : {
                // Compatible headers for production (works on all browsers)
                Accept: 'application/json, text/plain, */*',
                'Accept-Language': 'en-US,en;q=0.9',
              }),
        },
        body: JSON.stringify(requestData),
      })

      // Removed: console.log('API response status:', response.status, response.statusText)

      // Check if response is OK
      if (!response.ok) {
        const errorText = await response.text()
        // Removed: console.error('API error response:', errorText)

        // If it's a 403 or 429 (rate limit/bot detection), retry with exponential backoff
        if ((response.status === 403 || response.status === 429) && retryCount < 3) {
          // Removed: console.log(
          //   `Retrying API call (attempt ${retryCount + 1}/3) after ${Math.pow(2, retryCount) * 1000}ms delay`,
          // )
          await new Promise((resolve) => setTimeout(resolve, Math.pow(2, retryCount) * 1000))
          return callApi(data, retryCount + 1)
        }

        throw new Error(
          `HTTP ${response.status}: ${response.statusText} - ${errorText.substring(0, 200)}`,
        )
      }

      // Check content type
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text()
        // Removed: console.error('Non-JSON response:', text)
        throw new Error(
          `Server returned non-JSON response. Expected JSON but got: ${contentType}. Response: ${text.substring(0, 200)}`,
        )
      }

      const result = await response.json()
      // Removed: console.log('API result:', result)
      return result
    } catch (err) {
      // Removed: console.error('API call error:', err)
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
      // Removed: console.log('Starting file upload:', {
      //   fileName: file.name,
      //   fileSize: file.size,
      //   fileType: file.type,
      //   destinationFolder,
      //   customFilename,
      //   uploadUrl: UPLOAD_URL,
      // })

      // Load upload path from db_code.json
      const uploadPath = await loadUploadPath()

      const formData = new FormData()
      formData.append('file', file)
      formData.append('base_directory', uploadPath)
      formData.append('destination_folder', destinationFolder)
      if (customFilename) {
        formData.append('custom_filename', customFilename)
      }

      // Add timeout to prevent hanging
      const controller = new AbortController()
      const timeoutId = setTimeout(() => {
        // Removed: console.log('Upload timeout - aborting request')
        controller.abort()
      }, 30000) // 30 second timeout

      const response = await fetch(UPLOAD_URL, {
        method: 'POST',
        headers: {
          ...(isLocalhost
            ? {
                // Simple headers for localhost
              }
            : {
                // Compatible headers for production (works on all browsers)
                Accept: 'application/json, text/plain, */*',
                'Accept-Language': 'en-US,en;q=0.9',
              }),
        },
        body: formData,
        signal: controller.signal,
      })

      clearTimeout(timeoutId)
      // Removed: console.log('Upload response status:', response.status, response.statusText)

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status} - ${response.statusText}`)
      }

      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text()
        // Removed: console.error('Non-JSON response:', text)
        throw new Error(`Server returned non-JSON response: ${text.substring(0, 200)}`)
      }

      const result = await response.json()
      // Removed: console.log('Upload result:', result)

      if (!result.success) {
        throw new Error(result.message || 'Upload failed')
      }

      // Return both the server response and the relative path
      return {
        ...result,
        relativePath: `${destinationFolder}/${customFilename || result.file_path.split('/').pop()}`,
      }
    } catch (err) {
      // Removed: console.error('Upload error details:', err)

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

    // If the path already contains 'upload.php?path=' or 'upload_simple.php?path=', extract parameters
    if (
      processedPath.includes('upload.php?path=') ||
      processedPath.includes('upload_simple.php?path=')
    ) {
      // Extract the query string part (everything after '?')
      const queryString = processedPath.includes('?')
        ? processedPath.split('?').slice(1).join('?')
        : ''

      // Parse the query parameters to extract path and base_directory separately
      const urlParams = new URLSearchParams(queryString)
      const filePath = urlParams.get('path')
      const baseDirectory = urlParams.get('base_directory')

      if (filePath) {
        // Build the URL with proper query parameters
        const url = new URL(UPLOAD_URL)
        url.searchParams.set('path', filePath)
        if (baseDirectory) {
          url.searchParams.set('base_directory', baseDirectory)
        }
        return url.toString()
      }
    }

    return `${UPLOAD_URL}?path=${encodeURIComponent(processedPath)}`
  }

  // Function to handle cookie verification challenges
  const handleCookieVerification = async () => {
    // Only run cookie verification on production
    if (isLocalhost) {
      // Removed: console.log('Skipping cookie verification on localhost')
      return true
    }

    try {
      // Removed: console.log('Attempting to handle cookie verification...')

      // Load database name from db_code.json
      const db_name = await loadDbName()

      // First, try to access the API endpoint to establish cookies
      const mainPageResponse = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json, text/plain, */*',
          'Accept-Language': 'en-US,en;q=0.9',
        },
        body: JSON.stringify({ action: 'ping', dbname: db_name }),
      })

      // Removed: console.log('Main page response status:', mainPageResponse.status)

      // Wait a bit before making the actual API call
      await new Promise((resolve) => setTimeout(resolve, 2000))

      return true
    } catch (err) {
      // Removed: console.error('Error handling cookie verification:', err)
      return false
    }
  }

  // Get assets (logo.png, letter_head.png, gml2.png)
  // Returns JSON with file locations, copies from js_dir to files_dir if needed
  // Uses localStorage to cache results
  async function getAssets() {
    const STORAGE_KEY = 'assets_cache'
    const ASSET_FILES = ['logo.png', 'letter_head.png', 'gml2.png']

    // Check localStorage first
    try {
      const cached = localStorage.getItem(STORAGE_KEY)
      if (cached) {
        const cachedData = JSON.parse(cached)
        // Check if cache is still valid (you can add timestamp validation here if needed)
        return cachedData
      }
    } catch (err) {
      console.warn('Error reading assets cache:', err)
    }

    try {
      // Get db_code from db_code.json
      const dbCodeResponse = await fetch('/db_code.json')
      if (!dbCodeResponse.ok) {
        throw new Error('Failed to load db_code.json')
      }
      const dbCodeData = await dbCodeResponse.json()

      if (!dbCodeData.db_code) {
        throw new Error('db_code not found in db_code.json')
      }

      // Get database info from dbs table
      const dbResponse = await fetch(
        `${DB_MANAGER_API_URL}?action=get_database_by_code&db_code=${encodeURIComponent(dbCodeData.db_code)}`,
      )

      if (!dbResponse.ok) {
        throw new Error('Failed to fetch database information')
      }

      const dbResult = await dbResponse.json()

      if (!dbResult.success || !dbResult.data) {
        throw new Error('Database not found for the provided db_code')
      }

      const { files_dir, js_dir } = dbResult.data

      if (!files_dir || !js_dir) {
        throw new Error('files_dir and js_dir must be configured')
      }

      // Check if files exist in files_dir using API
      const checkResponse = await fetch(DB_MANAGER_API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          action: 'check_assets',
          files_dir: files_dir,
        }),
      })

      const checkResult = await checkResponse.json()
      const existingFiles = checkResult.success ? checkResult.data?.existing || [] : []
      const missingFiles = ASSET_FILES.filter((file) => !existingFiles.includes(file))

      // If files are missing, copy them from js_dir to files_dir
      if (missingFiles.length > 0) {
        const copyResponse = await fetch(DB_MANAGER_API_URL, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: new URLSearchParams({
            action: 'copy_assets',
            js_dir: js_dir,
            files_dir: files_dir,
          }),
        })

        const copyResult = await copyResponse.json()

        if (!copyResult.success) {
          throw new Error(`Failed to copy assets: ${copyResult.message || 'Unknown error'}`)
        }
      }

      // Build file URLs
      const filesDirPath = files_dir.startsWith('/') ? files_dir.substring(1) : files_dir
      const filesDirUrl = `${UPLOAD_URL}?base_directory=${encodeURIComponent(filesDirPath)}&path=`

      // Build result object with all file locations
      const result = {
        logo: filesDirUrl + encodeURIComponent('logo.png'),
        letter_head: filesDirUrl + encodeURIComponent('letter_head.png'),
        gml2: filesDirUrl + encodeURIComponent('gml2.png'),
        files_dir: files_dir,
      }

      // Cache in localStorage
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(result))
      } catch (err) {
        console.warn('Error caching assets:', err)
      }

      return result
    } catch (err) {
      console.error('Error getting assets:', err)
      throw err
    }
  }

  return {
    callApi,
    uploadFile,
    getFileUrl,
    handleCookieVerification,
    getAssets,
    error,
    loading,
  }
}
