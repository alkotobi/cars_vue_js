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

  console.log('[useApi] getBasePath() called')
  console.log('[useApi] import.meta.env.BASE_URL:', import.meta.env.BASE_URL)
  console.log('[useApi] window.location.pathname:', window.location.pathname)
  console.log('[useApi] window.location.href:', window.location.href)

  // If base is relative, convert to absolute path
  if (baseUrl === './' || baseUrl.startsWith('./')) {
    const pathname = window.location.pathname
    // If pathname is like '/mig/login', extract '/mig/'
    // If pathname is like '/login', use '/'
    const match = pathname.match(/^(\/[^/]+\/)/)
    const detectedBase = match ? match[1] : '/'
    console.log('[useApi] Pathname match:', match)
    console.log('[useApi] Detected base path:', detectedBase)
    return detectedBase
  }

  // If base is already absolute, use it as is
  console.log('[useApi] Using absolute base URL:', baseUrl)
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

    // Clear any cached image data
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

    // Force clear browser image cache by creating a dummy image and setting src to empty
    // This helps clear the browser's internal image cache
    try {
      const img = new Image()
      img.src = ''
      console.log('[updateAssetsVersion] Cleared image cache')
    } catch (e) {
      console.warn('[updateAssetsVersion] Could not clear image cache:', e)
    }

    return newVersion
  }

  // ============================================
  // Car Files Management Functions
  // ============================================

  // Get current user from localStorage
  const getCurrentUser = () => {
    const userStr = localStorage.getItem('user')
    return userStr ? JSON.parse(userStr) : null
  }

  // Get file categories
  const getCarFileCategories = async () => {
    const result = await callApi({
      action: 'get_car_file_categories',
      requiresAuth: false,
    })
    console.log('[useApi] getCarFileCategories result:', result)
    // API returns { success: true, data: [...] } or { success: true, results: [...] }
    if (result.success) {
      return result.data || result.results || []
    }
    return []
  }

  // Get files for a car (with permission filtering)
  const getCarFiles = async (carId) => {
    const user = getCurrentUser()
    const result = await callApi({
      action: 'get_car_files',
      car_id: carId,
      user_id: user?.id || null,
      is_admin: user?.role_id === 1,
      requiresAuth: true,
    })
    console.log('[useApi] getCarFiles result:', result)
    // API returns { success: true, data: [...] }
    if (result.success) {
      return result.data || []
    }
    return []
  }

  // Upload a file and create file record
  const uploadCarFile = async (file, carId, categoryId, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    // Upload file first
    const fileExtension = file.name.split('.').pop().toLowerCase()
    const filename = `file_${Date.now()}.${fileExtension}`
    const destinationFolder = `cars/${carId}/${categoryId}`

    const uploadResult = await uploadFile(file, destinationFolder, filename)

    if (!uploadResult.success) {
      throw new Error(uploadResult.message || 'File upload failed')
    }

    // Create file record in database
    const filePath = uploadResult.relativePath || uploadResult.file_path
    const createResult = await callApi({
      action: 'create_car_file',
      car_id: carId,
      category_id: categoryId,
      file_path: filePath,
      file_name: file.name,
      file_size: file.size,
      file_type: file.type,
      uploaded_by: user.id,
      notes: notes,
      requiresAuth: true,
    })

    if (!createResult.success) {
      throw new Error(createResult.error || 'Failed to create file record')
    }

    return {
      ...createResult,
      file_path: filePath,
      file_name: file.name,
    }
  }

  // Delete a file (soft delete)
  const deleteCarFile = async (fileId) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    const result = await callApi({
      action: 'delete_car_file',
      file_id: fileId,
      user_id: user.id,
      is_admin: user.role_id === 1,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to delete file')
    }

    return result
  }

  // Check out physical copy
  const checkoutPhysicalCopy = async (fileId, expectedReturnDate = null, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    const result = await callApi({
      action: 'checkout_physical_copy',
      file_id: fileId,
      user_id: user.id,
      expected_return_date: expectedReturnDate,
      notes: notes,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to check out file')
    }

    return result
  }

  // Check in physical copy
  const checkinPhysicalCopy = async (fileId, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    const result = await callApi({
      action: 'checkin_physical_copy',
      file_id: fileId,
      user_id: user.id,
      notes: notes,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to check in file')
    }

    return result
  }

  // Transfer physical copy
  const transferPhysicalCopy = async (
    fileId,
    toUserId,
    notes = null,
    expectedReturnDate = null,
  ) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    // Get file info to find current holder or uploader
    // If file is available, use uploader as from_user_id
    // If file is checked out, use current_holder_id
    const fileQueryResult = await callApi({
      query:
        'SELECT cf.id, cf.uploaded_by, cpt.current_holder_id, cpt.status FROM car_files cf LEFT JOIN car_file_physical_tracking cpt ON cf.id = cpt.car_file_id AND cpt.status IN ("available", "checked_out") WHERE cf.id = ?',
      params: [fileId],
      requiresAuth: true,
    })

    const fileData = fileQueryResult.data?.[0]
    // If file is available, use uploader; otherwise use current holder
    const fromUserId =
      fileData?.status === 'available'
        ? fileData?.uploaded_by || user.id
        : fileData?.current_holder_id || user.id

    const result = await callApi({
      action: 'transfer_physical_copy',
      file_id: fileId,
      from_user_id: fromUserId,
      to_user_id: toUserId,
      transferred_by: user.id,
      notes: notes,
      expected_return_date: expectedReturnDate,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to transfer file')
    }

    return result
  }

  // Get transfer history
  const getFileTransferHistory = async (fileId) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    const result = await callApi({
      action: 'get_file_transfer_history',
      file_id: fileId,
      user_id: user.id,
      is_admin: user.role_id === 1,
      requiresAuth: true,
    })
    return result.success ? result.data : []
  }

  // Get my physical copies
  const getMyPhysicalCopies = async () => {
    const user = getCurrentUser()
    if (!user) {
      return []
    }

    const result = await callApi({
      action: 'get_my_physical_copies',
      user_id: user.id,
      requiresAuth: true,
    })
    return result.success ? result.data : []
  }

  // Get pending transfers for current user
  const getPendingTransfers = async () => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    console.log(
      '[useApi] getPendingTransfers: Calling API with user_id:',
      user.id,
      'username:',
      user.username,
    )
    const result = await callApi({
      action: 'get_pending_transfers',
      user_id: user.id,
      requiresAuth: true,
    })

    console.log('[useApi] getPendingTransfers: Full API response:', JSON.stringify(result, null, 2))

    if (result.debug) {
      console.log('[useApi] getPendingTransfers: Debug info:', result.debug)
      console.log(
        '[useApi] getPendingTransfers: All pending transfers in DB:',
        result.debug.all_pending,
      )
    }

    if (!result.success) {
      throw new Error(result.error || 'Failed to get pending transfers')
    }

    console.log('[useApi] getPendingTransfers: Returning', result.data?.length || 0, 'transfers')
    if (result.data && result.data.length > 0) {
      console.log(
        '[useApi] getPendingTransfers: Transfer details:',
        result.data.map((t) => ({
          id: t.id,
          file_name: t.file_name,
          from_user_id: t.from_user_id,
          to_user_id: t.to_user_id,
          transfer_status: t.transfer_status,
        })),
      )
    }
    return result.data || []
  }

  // Approve a pending transfer
  const approveTransfer = async (transferId, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    const result = await callApi({
      action: 'approve_transfer',
      transfer_id: transferId,
      user_id: user.id,
      notes: notes,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to approve transfer')
    }

    return result
  }

  // Reject a pending transfer
  const rejectTransfer = async (transferId, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    const result = await callApi({
      action: 'reject_transfer',
      transfer_id: transferId,
      user_id: user.id,
      notes: notes,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to reject transfer')
    }

    return result
  }

  // Get users for transfer dropdown
  const getUsersForTransfer = async (fileId) => {
    const result = await callApi({
      action: 'get_users_for_transfer',
      file_id: fileId,
      requiresAuth: true,
    })
    return result.success ? result.data : []
  }

  // Create file category (admin only)
  const createFileCategory = async (categoryData) => {
    const user = getCurrentUser()
    if (!user || user.role_id !== 1) {
      throw new Error('Admin access required')
    }

    const result = await callApi({
      action: 'create_file_category',
      ...categoryData,
      is_admin: true,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to create category')
    }

    return result
  }

  // Update file category (admin only)
  const updateFileCategory = async (categoryId, categoryData) => {
    const user = getCurrentUser()
    if (!user || user.role_id !== 1) {
      throw new Error('Admin access required')
    }

    const result = await callApi({
      action: 'update_file_category',
      category_id: categoryId,
      ...categoryData,
      is_admin: true,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to update category')
    }

    return result
  }

  // Custom Clearance Agents CRUD
  const getCustomClearanceAgents = async () => {
    const result = await callApi({
      action: 'get_custom_clearance_agents',
      requiresAuth: true,
    })
    if (result.success) {
      return result.data || []
    }
    throw new Error(result.error || 'Failed to fetch custom clearance agents')
  }

  const createCustomClearanceAgent = async (agentData) => {
    const user = getCurrentUser()
    if (!user || user.role_id !== 1) throw new Error('Admin access required')

    const result = await callApi({
      action: 'create_custom_clearance_agent',
      is_admin: true,
      ...agentData,
      requiresAuth: true,
    })
    if (result.success) {
      return result.message || 'Agent created successfully'
    }
    throw new Error(result.error || 'Failed to create agent')
  }

  const updateCustomClearanceAgent = async (agentId, agentData) => {
    const user = getCurrentUser()
    if (!user || user.role_id !== 1) throw new Error('Admin access required')

    const result = await callApi({
      action: 'update_custom_clearance_agent',
      id: agentId,
      is_admin: true,
      ...agentData,
      requiresAuth: true,
    })
    if (result.success) {
      return result.message || 'Agent updated successfully'
    }
    throw new Error(result.error || 'Failed to update agent')
  }

  const deleteCustomClearanceAgent = async (agentId) => {
    const user = getCurrentUser()
    if (!user || user.role_id !== 1) throw new Error('Admin access required')

    const result = await callApi({
      action: 'delete_custom_clearance_agent',
      id: agentId,
      is_admin: true,
      requiresAuth: true,
    })
    if (result.success) {
      return result.message || 'Agent deleted successfully'
    }
    throw new Error(result.error || 'Failed to delete agent')
  }

  // Rollback checkout (admin only)
  const rollbackCheckout = async (fileId, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw new Error('User not authenticated')
    }

    const result = await callApi({
      action: 'rollback_checkout',
      file_id: fileId,
      user_id: user.id,
      notes: notes,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to rollback checkout')
    }

    return result
  }

  return {
    callApi,
    uploadFile,
    getFileUrl,
    handleCookieVerification,
    getAssets,
    updateAssetsVersion,
    // Car Files Management
    getCarFileCategories,
    getCarFiles,
    uploadCarFile,
    deleteCarFile,
    checkoutPhysicalCopy,
    checkinPhysicalCopy,
    transferPhysicalCopy,
    rollbackCheckout,
    getFileTransferHistory,
    getMyPhysicalCopies,
    getPendingTransfers,
    approveTransfer,
    rejectTransfer,
    getUsersForTransfer,
    createFileCategory,
    updateFileCategory,
    getCustomClearanceAgents,
    createCustomClearanceAgent,
    updateCustomClearanceAgent,
    deleteCustomClearanceAgent,
    error,
    loading,
  }
}
