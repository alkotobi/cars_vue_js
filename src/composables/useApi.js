import { ref } from 'vue'
import { readJsonResponse } from '@/utils/readJsonResponse'
import { isSessionLost, reportSessionLost, sessionLostError } from './useSessionLost'
import { getBasePath as sharedGetBasePath, resolveApiBaseUrl } from '../utils/basePath'

// Get the current hostname and protocol
const hostname = window.location.hostname
const protocol = window.location.protocol

// Detect base path from current location (e.g., '/mig_26/' or '/')
// This allows the app to work from any subdirectory or root
// Uses the same logic as getRouterBasePath() in router/index.js for consistency
const getBasePath = () => sharedGetBasePath()

const BASE_PATH = getBasePath()

// "Are we running against a local dev box?" Gates behaviour that must not apply
// in production: the artificial request delay, the relaxed Accept headers, and
// skipping cookie verification.
//
// This is derived from import.meta.env.DEV — the same signal resolveApiBaseUrl
// uses — rather than from the hostname. The previous check treated any
// 192.168.* address as local, which silently disabled cookie verification on a
// server genuinely deployed on a LAN. Note it used to be a bare `hostname ===`
// test and was briefly removed during the API-URL refactor while four call sites
// below still referenced it, which threw ReferenceError on every API call.
const isLocalhost = Boolean(import.meta.env?.DEV)

// Set API base URL. The API ships inside the app folder, so it is derived from
// the mount point: <origin><basePath>api. That makes it correct for any folder
// name, any domain and any IP without a rebuild. Only `npm run dev` differs —
// there PHP runs on :8000 — and that is detected from import.meta.env.DEV
// rather than from the hostname, so a server on a LAN address is not mistaken
// for a dev box. Override with VITE_API_BASE_URL to point somewhere else
// entirely, e.g. a tunnel:  VITE_API_BASE_URL=http://localhost:8000/api npx vite
// A .env.local override is untracked, so this stays out of commits by default.
const API_BASE_URL = resolveApiBaseUrl({
  override: import.meta.env?.VITE_API_BASE_URL,
  protocol,
  hostname,
  port: window.location.port,
  basePath: BASE_PATH,
  isDev: import.meta.env.DEV,
})

const API_URL = `${API_BASE_URL}/api.php`
const UPLOAD_URL = `${API_BASE_URL}/upload.php`
const DB_MANAGER_API_URL = `${API_BASE_URL}/db_manager_api.php`

// Promise to prevent concurrent loads (but no caching - always fetch fresh data)
let config_promise = null

// Store upload_path for getFileUrl (updated on each loadConfig call)
// This allows getFileUrl to work synchronously while using the correct base_directory
let current_upload_path = null

// Read the API token out of localStorage, tolerating a corrupt value.
//
// router/index.js and a dozen components parse this key unguarded, so a partial
// write turns a bad session into an unhandled throw during navigation. Callers
// that only need the token should use this rather than JSON.parse themselves.
export function getStoredToken() {
  try {
    const raw = localStorage.getItem('user')
    if (!raw) return null
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed.token === 'string' ? parsed.token : null
  } catch {
    return null
  }
}

// Function to load configuration from db_code.json
// db_code.json is PER SERVER, not per build: it names the database this
// deployment talks to, so it is fetched at runtime from <mount>db_code.json
// rather than compiled in. That is what lets a single dist/ serve many clients,
// each against its own database. deploy/deploy.sh writes it on the server and
// excludes it from the rsync; the copy in public/ is a local development default
// and is never deployed, since a stale one silently points a client at the
// wrong database.
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
      const response = await fetch(dbCodeUrl)
      if (!response.ok) {
        // A 404 means the per-server file was never written. Name the fix
        // instead of surfacing a bare status the operator cannot act on.
        throw new Error(
          response.status === 404
            ? `db_code.json not found at ${dbCodeUrl}. This file is per server and is not part ` +
              `of the build - run: ./deploy/deploy.sh <folder> <host> <db_code>`
            : `Failed to load db_code.json: ${response.status} ${response.statusText}`,
        )
      }

      // Check content type before parsing
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text()
        throw new Error(
          `db_code.json returned non-JSON response (Content-Type: ${contentType}). Check if the file exists at ${dbCodeUrl}`,
        )
      }

      const data = await readJsonResponse(response, dbCodeUrl)

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
            throw new Error(
              `Failed to fetch database info: ${dbResponse.status} ${dbResponse.statusText}`,
            )
          }
          const contentType = dbResponse.headers.get('content-type')
          if (!contentType || !contentType.includes('application/json')) {
            const text = await dbResponse.text()
            throw new Error(
              `API returned non-JSON response. Check if ${DB_MANAGER_API_URL} exists.`,
            )
          }
          const dbResult = await readJsonResponse(dbResponse, DB_MANAGER_API_URL)

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

export { BASE_PATH }

export const useApi = () => {
  const error = ref(null)
  const loading = ref(false)

  // Original API call function
  const callApi = async (data, retryCount = 0) => {
    loading.value = true
    error.value = null

    try {
      // Send the token whenever we have one.
      //
      // This used to be `data.requiresAuth ? user : null`, so any call that forgot
      // to opt in went out anonymous - and the API trusted that, so it mattered.
      // Now the token rides along on everything: the server decides what a request
      // is allowed, and a call site no longer has to remember to ask for a token.
      // Calls from /login, /settings and a share link still send nothing, because
      // there is no user in localStorage to take one from.
      //
      // Read *before* the delay below. It used to sit after the await, so logout
      // read localStorage ~1s after LogoutButton had already cleared it and went
      // out with token: undefined - the session stayed valid server-side.
      const storedToken = getStoredToken()

      // An explicitly-passed token wins. `{...data, token: stored}` used to
      // overwrite it, which silently defeated the logout path above.
      const requestToken = data.token ?? storedToken

      // Add a small random delay to make requests look more human-like (only on production)
      if (retryCount === 0 && !isLocalhost) {
        const delay = Math.random() * 1000 + 500 // 500-1500ms delay
        await new Promise((resolve) => setTimeout(resolve, delay))
      }

      // No dbname: the server picks its database from db_code.json now. It used to
      // read one off this request, which let a caller aim the API's credentials at
      // any database on the host. Sending it was never more than decoration.
      const requestData = {
        ...data,
        token: requestToken,
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

      const result = await readJsonResponse(response, API_URL)
      // Removed: console.log('API result:', result)

      // The auth gate answers a rejected token with code 'not_authenticated' and
      // HTTP 200, so there is no status to branch on and every caller was left to
      // notice it by hand - which is why a dead session used to surface as a raw
      // string or a native alert. Hand it to the one listener that opens the
      // session-expired modal instead. The envelope is still returned untouched,
      // so any call site that wants to render its own message still can.
      //
      // No check on requestToken: the server rejects a gated action sent with *no*
      // token with the same code, so a signed-out tab reports through here too
      // rather than depending on the router guard catching the next navigation.
      //
      // This sits after the retry block above on purpose: a 403/429 is retried,
      // this is not, because re-sending a token the server has already refused
      // three times only delays the same answer.
      if (isSessionLost(result, data.action)) {
        reportSessionLost(requestToken)
      }

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
      // upload.php is authenticated (it used to serve and store files with no
      // auth at all, which let anyone write a .php webshell or read
      // config.local.php). A form body cannot carry the X-Api-Token header a
      // file:// <img> cannot set, so the token rides as a field here.
      const uploadToken = getStoredToken()
      if (uploadToken) {
        formData.append('token', uploadToken)
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

      const result = await readJsonResponse(response, UPLOAD_URL)
      // Removed: console.log('Upload result:', result)

      // upload.php authenticates through the same require_api_user() guard and so
      // answers a dead token with the same { code: 'not_authenticated' } envelope.
      // Without this, a document attached to an expired session failed as a bare
      // "Upload failed" and the user never learned why.
      if (isSessionLost(result)) {
        reportSessionLost(uploadToken ?? null)
      }

      if (!result.success) {
        throw new Error(result.message || 'Upload failed')
      }

      // Extract the actual file path from the server response
      // Server returns: '/api/upload.php?path=documents/file.pdf&base_directory=mig_files'
      // We need to extract just the path part for storage
      let relativePath = result.file_path

      // If the server returned a full URL path, extract just the relative path
      if (result.file_path && result.file_path.includes('?path=')) {
        const urlParams = new URLSearchParams(result.file_path.split('?')[1])
        const pathParam = urlParams.get('path')
        if (pathParam) {
          relativePath = pathParam
        }
      } else {
        // Fallback: construct relative path from destination folder and filename
        relativePath = `${destinationFolder}/${customFilename || result.file_path.split('/').pop()}`
      }

      // Return both the server response and the relative path
      return {
        ...result,
        relativePath: relativePath,
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
    // If path is null, undefined, or empty, return empty string
    if (!path || typeof path !== 'string' || path.trim() === '') {
      return ''
    }

    // If path is already a full URL, return it as is
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return path
    }

    // If the path already contains 'upload.php?path=' or 'upload_simple.php?path=', extract parameters
    // This handles paths stored as '/api/upload.php?path=...&base_directory=...' or 'upload.php?path=...'
    if (path.includes('upload.php?path=') || path.includes('upload_simple.php?path=')) {
      // Extract the query string part (everything after '?')
      const queryString = path.includes('?') ? path.split('?').slice(1).join('?') : ''

      // Parse the query parameters to extract path and base_directory separately
      try {
        const urlParams = new URLSearchParams(queryString)
        const filePath = urlParams.get('path')
        const baseDirectory = urlParams.get('base_directory')

        if (filePath) {
          // Build the URL with proper query parameters using the UPLOAD_URL
          const url = new URL(UPLOAD_URL)
          url.searchParams.set('path', filePath)
          if (baseDirectory) {
            url.searchParams.set('base_directory', baseDirectory)
          }
          const token = getStoredToken()
          if (token) {
            url.searchParams.set('token', token)
          }
          return url.toString()
        }
      } catch (err) {
        console.error('Error parsing file path URL:', err, 'Path:', path)
        // Fall through to handle as regular path
      }
    }

    // If path starts with '/api/', remove it (it's a relative path to the API)
    // This handles paths like '/api/documents/file.pdf' -> 'documents/file.pdf'
    let processedPath = path.startsWith('/api/')
      ? path.substring(5) // Remove '/api/'
      : path.replace(/^\/+/, '') // Remove leading slashes

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
    // upload.php now requires a token on the GET branch too. An <img src> cannot
    // send a header, so it has to travel in the query string - which means it can
    // land in an access log. That is the cost of closing the unauthenticated
    // arbitrary-file-read hole without moving every asset behind an API route.
    const token = getStoredToken()
    if (token) {
      url.searchParams.set('token', token)
    }
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

      // First, try to access the API endpoint to establish cookies
      const mainPageResponse = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json, text/plain, */*',
          'Accept-Language': 'en-US,en;q=0.9',
        },
        body: JSON.stringify({ action: 'ping' }),
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
    return result
  }

  /**
   * Load letterhead URL (path_letter_head or default). Prefer useInvoiceCompanyInfo().getCompanyLogoUrl()
   * for report headers so browser/CrossDev use the same image as xlsx invoice (company logo from banks.logo_path).
   * This is kept for backward compatibility or settings UIs that need path_letter_head.
   */
  async function loadLetterhead(user = null) {
    // Get user from localStorage if not provided
    if (!user) {
      const userStr = localStorage.getItem('user')
      user = userStr ? JSON.parse(userStr) : null
    }

    // Get current base path for fallback URLs
    const currentBasePath = getBasePath()
    const STORAGE_KEY = 'assets_version'
    let assetsVersion = localStorage.getItem(STORAGE_KEY) || Date.now().toString()
    const cacheBuster = `?v=${assetsVersion}`

    // Check if user is a different company user - handle 0/1 from database
    const isDifferentCompany =
      user &&
      (user.is_diffrent_company === 1 ||
        user.is_diffrent_company === true ||
        user.is_diffrent_company === '1')

    // Check if user is a different company user and has a custom letterhead
    // Source: user.path_letter_head (legacy) or bank.path_letter_head (via id_bank_account)
    let letterHeadPath =
      user.path_letter_head && user.path_letter_head.trim() !== '' ? user.path_letter_head : null
    if (!letterHeadPath && user.id_bank_account) {
      try {
        const bankResult = await callApi({
          query: 'SELECT path_letter_head FROM banks WHERE id = ?',
          params: [user.id_bank_account],
        })
        if (bankResult.success && bankResult.data?.[0]?.path_letter_head) {
          letterHeadPath = bankResult.data[0].path_letter_head
        }
      } catch (e) {}
    }
    if (isDifferentCompany && letterHeadPath) {
      const customLetterheadUrl = getFileUrl(letterHeadPath)
      if (customLetterheadUrl) {
        return customLetterheadUrl
      }
    }

    // If user is different company but no custom letterhead, use letter_head_default.png
    if (isDifferentCompany) {
      return `${currentBasePath}letter_head_default.png${cacheBuster}`
    }

    // For regular users, use default letter_head.png
    const assets = await getAssets()
    return assets?.letter_head || `${currentBasePath}letter_head.png${cacheBuster}`
  }

  // Function to update assets version (call this after uploading new assets)
  function updateAssetsVersion() {
    const STORAGE_KEY = 'assets_version'
    const newVersion = Date.now().toString()
    const oldVersion = localStorage.getItem(STORAGE_KEY)
    localStorage.setItem(STORAGE_KEY, newVersion)

    // Clear any cached image data
    if ('caches' in window) {
      caches.keys().then((names) => {
        names.forEach((name) => {
          if (name.includes('logo') || name.includes('assets')) {
            caches.delete(name)
          }
        })
      })
    }

    // Force clear browser image cache by creating a dummy image and setting src to empty
    // This helps clear the browser's internal image cache
    try {
      const img = new Image()
      img.src = ''
    } catch (e) {
      // Could not clear image cache
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
      // sessionLostError(), not new Error(...), in all ten of these guards below:
      // it reports the loss so the modal opens, and carries `code` so a caller
      // that translates apiErrorText()/colorErrorText() renders it in the reader's
      // language. The bare Error this replaced leaked 'User not authenticated' and
      // opened nothing.
      throw sessionLostError()
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
      throw sessionLostError()
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
      throw sessionLostError()
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
      throw sessionLostError()
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
      throw sessionLostError()
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
      throw sessionLostError()
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
      throw sessionLostError()
    }

    const result = await callApi({
      action: 'get_pending_transfers',
      user_id: user.id,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to get pending transfers')
    }

    return result.data || []
  }

  // Approve a pending transfer
  const approveTransfer = async (transferId, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw sessionLostError()
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
      throw sessionLostError()
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

  // Delete file category (admin only)
  const deleteFileCategory = async (categoryId) => {
    const user = getCurrentUser()
    if (!user || user.role_id !== 1) {
      throw new Error('Admin access required')
    }

    const result = await callApi({
      action: 'delete_file_category',
      category_id: categoryId,
      is_admin: true,
      requiresAuth: true,
    })

    if (!result.success) {
      throw new Error(result.error || 'Failed to delete category')
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

  // Colors CRUD.
  //
  // These deliberately do NOT go through the generic {query, params} passthrough
  // like the rest of this file. That path in api/api.php executes whatever SQL it
  // is handed and its only gate is a special case for payment_confirmed writes,
  // so an INSERT or DELETE on colors was reachable by anyone who could reach the
  // endpoint - the admin v-if in the component was never a check. These call the
  // token-gated handlers in api/actions/colors.php instead.
  //
  // Reads and writes need a valid token; delete_color is admin-only server-side.
  // Colours are reference data every user picks from when entering a car, which
  // is why create/update are not admin-only the way delete is.
  //
  // Each helper throws on failure, carrying the server's `code` and `meta` on the
  // error so colorErrorText() can still translate it - a duplicate colour or an
  // in-use refusal is the expected failure here, not an exception.
  const getColors = async () => {
    const result = await callApi({ action: 'get_colors', requiresAuth: true })
    if (result.success) {
      return result.colors || []
    }
    throw apiFailure(result, 'Failed to fetch colors')
  }

  const createColor = async ({ color, hexa }) => {
    const result = await callApi({
      action: 'create_color',
      color,
      hexa,
      requiresAuth: true,
    })
    if (!result.success) {
      throw apiFailure(result, 'Failed to create color')
    }
    return result
  }

  const updateColor = async (id, { color, hexa }) => {
    const result = await callApi({
      action: 'update_color',
      id,
      color,
      hexa,
      requiresAuth: true,
    })
    if (!result.success) {
      throw apiFailure(result, 'Failed to update color')
    }
    return result
  }

  const deleteColor = async (id) => {
    const result = await callApi({ action: 'delete_color', id, requiresAuth: true })
    if (!result.success) {
      throw apiFailure(result, 'Failed to delete color')
    }
    return result
  }

  // Container GPS tracking. Replaces execute_sql, which ran caller-supplied SQL
  // with no authentication - and the map popup's version of it wrote a literal
  // id_user: 1, so every save was attributed to whoever held id 1.
  //
  // getContainerTracking() with no refs covers every container; pass refs for a
  // subset. Returns rows newest-first per container.
  const getContainerTracking = async (containerRefs = null) => {
    const payload = { action: 'get_container_tracking' }
    if (Array.isArray(containerRefs)) {
      payload.container_refs = containerRefs
    }

    const result = await callApi(payload)
    if (!result.success) {
      throw apiFailure(result, 'Failed to load container tracking')
    }
    return result.tracking || []
  }

  // coords is a "lat,lng" string, or null to clear the recorded position.
  // id_user is resolved from the token server-side.
  const saveContainerTracking = async (containerRef, coords = null) => {
    const result = await callApi({
      action: 'save_container_tracking',
      container_ref: containerRef,
      tracking: coords,
    })
    if (!result.success) {
      throw apiFailure(result, 'Failed to save container tracking')
    }
    return result
  }

  // The /clients/:token share page. Public by design - the share_token is the
  // credential - so it makes one round trip instead of four.
  const getClientShareData = async (shareToken) => {
    const result = await callApi({ action: 'get_client_share_data', share_token: shareToken })
    if (!result.success) {
      throw apiFailure(result, 'Failed to load shared client data')
    }
    return {
      client: result.client || null,
      cars: result.cars || [],
      files: result.files || [],
      tracking: result.tracking || [],
    }
  }

  // The version dialog renders on /login, so this cannot require a session.
  const getDbVersion = async () => {
    const result = await callApi({ action: 'get_db_version' })
    if (!result.success) {
      throw apiFailure(result, 'Failed to read database version')
    }
    return result.version || null
  }

  // Rollback checkout (admin only)
  const rollbackCheckout = async (fileId, notes = null) => {
    const user = getCurrentUser()
    if (!user) {
      throw sessionLostError()
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
    loadLetterhead,
    loadDbName,
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
    deleteFileCategory,
    getCustomClearanceAgents,
    createCustomClearanceAgent,
    updateCustomClearanceAgent,
    deleteCustomClearanceAgent,
    // Colors
    getColors,
    createColor,
    updateColor,
    deleteColor,
    // Container tracking
    getContainerTracking,
    saveContainerTracking,
    // Public share page / version dialog
    getClientShareData,
    getDbVersion,
    error,
    loading,
  }
}

/**
 * Build an Error that still carries the server's machine-readable failure.
 *
 * apiErrorDie() answers with { success: false, code, meta }. callApi returns that
 * payload untouched, so a helper that throws has to copy `code`/`meta` onto the
 * error or the caller is left with an opaque string and no way to tell a
 * duplicate colour from a database outage. colorErrorText() reads the same two
 * properties apiErrorText() does, which is why this works for both.
 */
function apiFailure(result, fallbackMessage) {
  const err = new Error(result?.error || fallbackMessage)
  if (result?.code) {
    err.code = result.code
  }
  if (result?.meta) {
    err.meta = result.meta
  }
  return err
}

// to locale keys so raw English never reaches the UI. Unknown codes (DB-level
// errors, unexpected failures) return '' so the caller falls back to its own
// generic translated message and keeps the raw text in the console.
//
// colorErrorText() below is the same idea for the colours screen, and the reason the
// two are separate is in its own comment: shared codes mean different things in
// different places.
export function apiErrorText(t, result) {
  const code = result?.code
  const keys = {
    // The one code here that is not about the action. Every other key points into
    // buy.detailsTable because that is where these codes mean something; a dead
    // session means the same thing on every screen, so it reuses the modal's own
    // wording rather than minting a near-duplicate string. Without this the caller
    // falls back to its own key and blames the action - "error updating stock" - for
    // what is an expired credential.
    not_authenticated: 'sessionExpired.message',
    invalid_request: 'buy.detailsTable.invalidRequest',
    not_admin: 'buy.detailsTable.adminOnly',
    detail_locked: 'buy.detailsTable.detailLocked',
    qty_below_committed: 'buy.detailsTable.qtyBelowCommitted',
    invalid_qty: 'buy.detailsTable.invalidQty',
    bill_not_found: 'buy.detailsTable.billNotFound',
    detail_not_found: 'buy.detailsTable.detailNotFound',
    stock_already_updated: 'buy.detailsTable.stockAlreadyUpdated',
    no_details: 'buy.detailsTable.noDetailsToProcess',
    stock_rows_exist: 'buy.detailsTable.stockRowsExist',
    invalid_detail_qty: 'buy.detailsTable.invalidDetailQty',
  }
  const key = keys[code]
  if (!key) return ''
  switch (code) {
    case 'stock_rows_exist': {
      const details = result?.meta?.details || []
      const parts = details.map((d) =>
        t('buy.detailsTable.stockRowsExistDetail', { id: d.id, count: d.count }),
      )
      return t(key, { count: details.length, parts: parts.join(', ') })
    }
    case 'invalid_detail_qty':
      return t(key, { id: result?.meta?.detailId })
    case 'qty_below_committed':
      return t(key, {
        requested: result?.meta?.requested || 0,
        removable: result?.meta?.removable || 0,
        committed: result?.meta?.committed || 0,
        min: result?.meta?.minQty || 0,
      })
    default:
      return t(key)
  }
}

/**
 * Translate a failure from api/actions/colors.php into the reader's language.
 *
 * Separate from apiErrorText() because the two shared codes mean something
 * different in each place: not_admin here is about a colour, not a buy detail,
 * so reusing buy.detailsTable.adminOnly ("Only an admin can edit or delete buy
 * details") on the colours screen would name the wrong screen. Same contract as
 * apiErrorText(): unknown codes return '' so the caller falls back to its own
 * generic translated message and keeps the raw text in the console.
 *
 * @param {Function} t vue-i18n translate
 * @param {{code?: string, meta?: object}|Error|null} result an apiFailure()
 *        error, or the raw {success: false, code, meta} payload
 * @returns {string} '' when the code is not one we know
 */
export function colorErrorText(t, result) {
  const code = result?.code
  const keys = {
    not_authenticated: 'colorsView.errors.notAuthenticated',
    not_admin: 'colorsView.errors.notAdmin',
    color_name_required: 'colorsView.errors.nameRequired',
    color_name_too_long: 'colorsView.errors.nameTooLong',
    color_invalid_hexa: 'colorsView.errors.invalidHexa',
    color_not_found: 'colorsView.errors.notFound',
    color_save_failed: 'colorsView.errors.saveFailed',
    color_in_use: 'colorsView.errors.inUse',
    color_exists: 'colorsView.errors.duplicateColor',
    db_unavailable: 'colorsView.errors.dbUnavailable',
    db_schema_outdated: 'colorsView.errors.schemaOutdated',
  }
  const key = keys[code]
  if (!key) return ''
  switch (code) {
    // colors.color and colors.hexa are two independent UNIQUE keys, and saying
    // which one collided is the difference between a user fixing the row and
    // guessing which field to change.
    case 'color_exists':
      return result?.meta?.field === 'hexa'
        ? t('colorsView.errors.duplicateHexa')
        : t('colorsView.errors.duplicateColor')
    case 'color_in_use':
      return t('colorsView.errors.inUseDetail', { count: result?.meta?.count || 0 })
    default:
      return t(key)
  }
}

/**
 * Translate a refused login into the reader's language, told apart by the code.
 *
 * The third of the three error-text translators, and the one that used not to exist:
 * LoginView rendered every unsuccessful login as auth.invalidCredentials, so a dead
 * database and a mistyped password produced the same words. That is not a rounding
 * error - it names the wrong cause, so the reader resets a password that did not need
 * resetting and never learns the server was down.
 *
 * Distinct from apiErrorText() and colorErrorText(), which also return '' for codes
 * they do not know: those let the caller fall back to its own message, because the
 * caller knows which action was being attempted. Here the caller is a login form with
 * one button, so there is nothing to fall back to and an unknown code gets the generic
 * login failure rather than a credential claim it has not earned.
 *
 * @param {Function} t vue-i18n translate
 * @param {{code?: string}|Error|null} result the raw {success: false, code} payload
 * @returns {string}
 */
export function loginErrorText(t, result) {
  switch (result?.code) {
    case 'invalid_credentials':
      return t('auth.invalidCredentials')
    // auth_connection() in api/actions/auth.php: a database that cannot be reached.
    case 'db_unavailable':
      return t('auth.serverUnavailable')
    default:
      return t('auth.loginError')
  }
}
