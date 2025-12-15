import { ref } from 'vue'

// Get the current hostname and protocol
const hostname = window.location.hostname
const protocol = window.location.protocol
const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'

// Detect base path from current location (e.g., '/mig/' or '/')
// This allows the app to work from any subdirectory
const getBasePath = () => {
  // Use Vite's BASE_URL if available (set in vite.config.js)
  // If it's relative (./), convert to absolute based on current pathname
  let baseUrl = import.meta.env.BASE_URL || './'

  // If base is relative, convert to absolute path
  if (baseUrl === './' || baseUrl.startsWith('./')) {
    const pathname = window.location.pathname
    // If pathname is like '/mig/login', extract '/mig/'
    // If pathname is like '/login', use '/'
    const match = pathname.match(/^(\/[^/]+\/)/)
    return match ? match[1] : '/'
  }

  // If base is already absolute, use it as is
  return baseUrl
}

const BASE_PATH = getBasePath()

// Set API base URL based on environment
// Use localhost API for development, production API for production
const API_BASE_URL = isLocalhost ? 'http://localhost:8000/api' : 'https://www.merhab.com/api'

const API_URL = `${API_BASE_URL}/api.php`
const UPLOAD_URL = `${API_BASE_URL}/upload.php`
const DB_MANAGER_API_URL = `${API_BASE_URL}/db_manager_api.php`

// Promise to prevent concurrent loads (but no caching - always fetch fresh data)
let config_promise = null

// Store upload_path for getFileUrl (updated on each loadConfig call)
// This allows getFileUrl to work synchronously while using the correct base_directory
let current_upload_path = null

// Function to load configuration from db_code.json
// This file is in the public folder and will be accessible at ./db_code.json (same folder as index.html)
// No caching - always fetches fresh data
async function loadConfig() {
  // Return existing promise if already loading (to prevent concurrent requests)
  if (config_promise) {
    return config_promise
  }

  // Start loading (always fetch fresh data)
  config_promise = (async () => {
    try {
      // Get current base path (where index.html is located)
      const currentBasePath = getBasePath()
      // Load db_code.json from same folder as index.html (use base path)
      const dbCodeUrl = `${currentBasePath}db_code.json`
      console.log('[loadConfig] Fetching db_code.json from:', dbCodeUrl)
      const response = await fetch(dbCodeUrl)
      if (!response.ok) {
        const errorText = await response.text()
        console.error('Failed to load db_code.json. Status:', response.status)
        console.error('Response:', errorText.substring(0, 500))
        throw new Error(`Failed to load db_code.json: ${response.status} ${response.statusText}`)
      }

      // Check content type before parsing
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text()
        console.error('db_code.json returned non-JSON response. Content-Type:', contentType)
        console.error('Response:', text.substring(0, 500))
        throw new Error(
          `db_code.json returned non-JSON response (Content-Type: ${contentType}). Check if the file exists at ${dbCodeUrl}`,
        )
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
            // Try to get error text for debugging
            const errorText = await dbResponse.text()
            console.error('API Error Response:', errorText.substring(0, 200))
            throw new Error(
              `Failed to fetch database info: ${dbResponse.status} ${dbResponse.statusText}`,
            )
          }
          const contentType = dbResponse.headers.get('content-type')
          if (!contentType || !contentType.includes('application/json')) {
            const text = await dbResponse.text()
            console.error('API returned non-JSON response:', text.substring(0, 500))
            throw new Error(
              `API returned non-JSON response. Check if ${DB_MANAGER_API_URL} exists.`,
            )
          }
          const dbResult = await dbResponse.json()

          if (dbResult.success && dbResult.data) {
            // Use db_name and files_dir from dbs table
            if (!dbResult.data.db_name) {
              throw new Error(
                `Database record found but db_name is missing for db_code: ${data.db_code}`,
              )
            }
            // Return fresh data (no caching)
            const config = {
              db_name: dbResult.data.db_name,
              upload_path: dbResult.data.files_dir || '', // Use files_dir from dbs table (can be null/empty)
            }
            // Update current_upload_path for getFileUrl
            current_upload_path = config.upload_path
            return config
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

        // Return fresh data (no caching)
        const config = {
          db_name: data.db_name,
          upload_path: data.uplod_path,
        }
        // Update current_upload_path for getFileUrl
        current_upload_path = config.upload_path
        return config
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
  // Uses current_upload_path (from loadConfig) as base_directory
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

    // Use current_upload_path (from database) as base_directory
    // This is updated whenever loadConfig() is called
    const baseDirectory = current_upload_path || 'mig_files'

    // Remove leading slash from baseDirectory if present (Unix path format)
    const cleanBaseDirectory = baseDirectory.startsWith('/')
      ? baseDirectory.substring(1)
      : baseDirectory

    // Build URL with base_directory parameter
    const url = new URL(UPLOAD_URL)
    url.searchParams.set('path', processedPath)
    url.searchParams.set('base_directory', cleanBaseDirectory)
    return url.toString()
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
  // Adds cache-busting query parameter to prevent browser caching
  async function getAssets() {
    console.log('[getAssets] Function called')

    // Get current base path (where index.html is located)
    const currentBasePath = getBasePath()

    // Use a persistent version stored in localStorage that gets updated when files are uploaded
    // This ensures cache-busting works even after page refresh
    const STORAGE_KEY = 'assets_version'
    let assetsVersion = localStorage.getItem(STORAGE_KEY)

    // If no version exists, initialize with current timestamp
    if (!assetsVersion) {
      assetsVersion = Date.now().toString()
      localStorage.setItem(STORAGE_KEY, assetsVersion)
    }

    // Add cache-busting version to force browser to reload images
    const cacheBuster = `?v=${assetsVersion}`

    // Files are in the same folder as index.html, use base path
    // This works regardless of deployment location (root or subdirectory)
    // No caching - always return fresh data since there's no computation
    const result = {
      logo: `${currentBasePath}logo.png${cacheBuster}`,
      letter_head: `${currentBasePath}letter_head.png${cacheBuster}`,
      gml2: `${currentBasePath}gml2.png${cacheBuster}`,
    }
    console.log('[getAssets] Created result object:', result)
    console.log('[getAssets] Current window location:', window.location.href)
    console.log(
      '[getAssets] Logo path will resolve to:',
      new URL(result.logo, window.location.href).href,
    )

    console.log('[getAssets] Returning result:', result)
    return result
  }

  // Function to update assets version (call this after uploading new assets)
  function updateAssetsVersion() {
    const STORAGE_KEY = 'assets_version'
    const newVersion = Date.now().toString()
    const oldVersion = localStorage.getItem(STORAGE_KEY)
    localStorage.setItem(STORAGE_KEY, newVersion)
    console.log('[updateAssetsVersion] Updated assets version from', oldVersion, 'to:', newVersion)
    // Also clear any cached image data
    if ('caches' in window) {
      caches.keys().then((names) => {
        names.forEach((name) => {
          if (name.includes('logo') || name.includes('assets')) {
            caches.delete(name)
            console.log('[updateAssetsVersion] Deleted cache:', name)
          }
        })
      })
    }
    return newVersion
  }

  return {
    callApi,
    uploadFile,
    getFileUrl,
    handleCookieVerification,
    getAssets,
    updateAssetsVersion,
    error,
    loading,
  }
}
