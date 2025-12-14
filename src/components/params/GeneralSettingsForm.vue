<template>
  <div class="general-settings-form">
    <div class="form-header">
      <h2>
        <i class="fas fa-sliders-h"></i>
        General Settings
      </h2>
      <p class="form-description">Upload logo and document images</p>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-if="successMessage" class="success-message">
      <i class="fas fa-check-circle"></i>
      {{ successMessage }}
    </div>

    <div v-if="loading" class="loading-state">
      <i class="fas fa-spinner fa-spin"></i>
      Loading configuration...
    </div>

    <div v-else class="upload-section">
      <div class="info-box">
        <i class="fas fa-info-circle"></i>
        <div>
          <p><strong>Upload Location:</strong> {{ uploadPath || 'Loading...' }}</p>
          <p v-if="dbCode" class="db-code-info">Database Code: {{ dbCode }}</p>
        </div>
      </div>

      <div class="file-uploads">
        <!-- Logo Upload -->
        <div class="file-upload-item">
          <label class="file-label">
            <i class="fas fa-image"></i>
            <span>Logo (logo.png)</span>
          </label>
          <div class="file-input-wrapper">
              <input
                id="logo-input"
                type="file"
                accept=".png"
                @change="handleFileSelect('logo', $event)"
                :disabled="uploading || !uploadPath"
                class="file-input"
              />
            <button
              type="button"
              @click="triggerFileInput('logo')"
              :disabled="uploading || !uploadPath"
              class="browse-btn"
            >
              <i class="fas fa-folder-open"></i>
              Browse
            </button>
          </div>
          <div v-if="selectedFiles.logo" class="selected-file">
            <i class="fas fa-file-image"></i>
            <span>{{ selectedFiles.logo.name }}</span>
            <span class="file-size">({{ formatFileSize(selectedFiles.logo.size) }})</span>
            <button
              type="button"
              @click="clearFile('logo')"
              class="clear-btn"
              :disabled="uploading"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
          <div v-if="uploadResults.logo" class="upload-result" :class="{ success: uploadResults.logo.success, error: !uploadResults.logo.success }">
            <i :class="uploadResults.logo.success ? 'fas fa-check-circle' : 'fas fa-exclamation-circle'"></i>
            <span>{{ uploadResults.logo.message }}</span>
          </div>
        </div>

        <!-- Letter Head Upload -->
        <div class="file-upload-item">
          <label class="file-label">
            <i class="fas fa-file-image"></i>
            <span>Letter Head (letter_head.png)</span>
          </label>
          <div class="file-input-wrapper">
              <input
                id="letter-head-input"
                type="file"
                accept=".png"
                @change="handleFileSelect('letter_head', $event)"
                :disabled="uploading || !uploadPath"
                class="file-input"
              />
            <button
              type="button"
              @click="triggerFileInput('letter_head')"
              :disabled="uploading || !uploadPath"
              class="browse-btn"
            >
              <i class="fas fa-folder-open"></i>
              Browse
            </button>
          </div>
          <div v-if="selectedFiles.letter_head" class="selected-file">
            <i class="fas fa-file-image"></i>
            <span>{{ selectedFiles.letter_head.name }}</span>
            <span class="file-size">({{ formatFileSize(selectedFiles.letter_head.size) }})</span>
            <button
              type="button"
              @click="clearFile('letter_head')"
              class="clear-btn"
              :disabled="uploading"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
          <div v-if="uploadResults.letter_head" class="upload-result" :class="{ success: uploadResults.letter_head.success, error: !uploadResults.letter_head.success }">
            <i :class="uploadResults.letter_head.success ? 'fas fa-check-circle' : 'fas fa-exclamation-circle'"></i>
            <span>{{ uploadResults.letter_head.message }}</span>
          </div>
        </div>

        <!-- GML2 Upload -->
        <div class="file-upload-item">
          <label class="file-label">
            <i class="fas fa-stamp"></i>
            <span>Stamp (gml2.png)</span>
          </label>
          <div class="file-input-wrapper">
              <input
                id="gml2-input"
                type="file"
                accept=".png"
                @change="handleFileSelect('gml2', $event)"
                :disabled="uploading || !uploadPath"
                class="file-input"
              />
            <button
              type="button"
              @click="triggerFileInput('gml2')"
              :disabled="uploading || !uploadPath"
              class="browse-btn"
            >
              <i class="fas fa-folder-open"></i>
              Browse
            </button>
          </div>
          <div v-if="selectedFiles.gml2" class="selected-file">
            <i class="fas fa-file-image"></i>
            <span>{{ selectedFiles.gml2.name }}</span>
            <span class="file-size">({{ formatFileSize(selectedFiles.gml2.size) }})</span>
            <button
              type="button"
              @click="clearFile('gml2')"
              class="clear-btn"
              :disabled="uploading"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
          <div v-if="uploadResults.gml2" class="upload-result" :class="{ success: uploadResults.gml2.success, error: !uploadResults.gml2.success }">
            <i :class="uploadResults.gml2.success ? 'fas fa-check-circle' : 'fas fa-exclamation-circle'"></i>
            <span>{{ uploadResults.gml2.message }}</span>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button
          @click="uploadFiles"
          :disabled="!canUpload || uploading"
          class="upload-btn"
        >
          <i v-if="uploading" class="fas fa-spinner fa-spin"></i>
          <i v-else class="fas fa-upload"></i>
          {{ uploading ? 'Uploading...' : 'Upload Selected Files' }}
        </button>
        <button
          @click="clearAll"
          :disabled="uploading || !hasSelectedFiles"
          class="clear-all-btn"
        >
          <i class="fas fa-times"></i>
          Clear All
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const uploading = ref(false)
const dbCode = ref(null)
const uploadPath = ref(null)
const selectedFiles = ref({
  logo: null,
  letter_head: null,
  gml2: null
})
const uploadResults = ref({
  logo: null,
  letter_head: null,
  gml2: null
})

// Get API base URL
const getApiBaseUrl = () => {
  const hostname = window.location.hostname
  const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'
  return isLocalhost ? 'http://localhost:8000/api' : 'https://www.merhab.com/api'
}

// Load configuration
const loadConfiguration = async () => {
  loading.value = true
  error.value = null
  
  try {
    // Step 1: Get db_code from db_code.json
    const dbCodeResponse = await fetch('/db_code.json')
    if (!dbCodeResponse.ok) {
      throw new Error('Failed to load db_code.json')
    }
    const dbCodeData = await dbCodeResponse.json()
    
    if (!dbCodeData.db_code) {
      throw new Error('db_code not found in db_code.json')
    }
    
    dbCode.value = dbCodeData.db_code
    
    // Step 2: Get database record from dbs table
    const dbResponse = await fetch(
      `${getApiBaseUrl()}/db_manager_api.php?action=get_database_by_code&db_code=${encodeURIComponent(dbCodeData.db_code)}`
    )
    
    if (!dbResponse.ok) {
      throw new Error('Failed to fetch database information')
    }
    
    const dbResult = await dbResponse.json()
    
    if (!dbResult.success || !dbResult.data) {
      throw new Error('Database not found for the provided db_code')
    }
    
    if (!dbResult.data.js_dir) {
      throw new Error('js_dir is not configured for this database')
    }
    
    uploadPath.value = dbResult.data.js_dir
    
  } catch (err) {
    error.value = err.message || 'Failed to load configuration'
    console.error(err)
  } finally {
    loading.value = false
  }
}

const handleFileSelect = (fileType, event) => {
  const file = event.target.files?.[0]
  if (!file) return
  
  // Validate file type
  if (file.type !== 'image/png') {
    error.value = `${fileType} must be a PNG file`
    setTimeout(() => {
      error.value = null
    }, 5000)
    return
  }
  
  // Expected filenames for each type
  const expectedNames = {
    logo: 'logo.png',
    letter_head: 'letter_head.png',
    gml2: 'gml2.png'
  }
  
  // Create a new File object with the correct name
  // This ensures the file is renamed to the appropriate name before upload
  const renamedFile = new File([file], expectedNames[fileType], {
    type: file.type,
    lastModified: file.lastModified
  })
  
  selectedFiles.value[fileType] = renamedFile
  uploadResults.value[fileType] = null
  error.value = null
}

const triggerFileInput = (fileType) => {
  const inputIds = {
    logo: 'logo-input',
    letter_head: 'letter-head-input',
    gml2: 'gml2-input'
  }
  
  const inputId = inputIds[fileType]
  if (inputId) {
    const input = document.getElementById(inputId)
    if (input) {
      input.click()
    }
  }
}

const clearFile = (fileType) => {
  selectedFiles.value[fileType] = null
  uploadResults.value[fileType] = null
  
  // Clear the input using IDs
  if (fileType === 'logo') {
    const input = document.getElementById('logo-input')
    if (input) input.value = ''
  } else if (fileType === 'letter_head') {
    const input = document.getElementById('letter-head-input')
    if (input) input.value = ''
  } else if (fileType === 'gml2') {
    const input = document.getElementById('gml2-input')
    if (input) input.value = ''
  }
}

const clearAll = () => {
  selectedFiles.value = {
    logo: null,
    letter_head: null,
    gml2: null
  }
  uploadResults.value = {
    logo: null,
    letter_head: null,
    gml2: null
  }
  error.value = null
  successMessage.value = null
  
  // Clear all inputs
  const inputs = document.querySelectorAll('.file-input')
  inputs.forEach(input => {
    input.value = ''
  })
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const canUpload = computed(() => {
  return Object.values(selectedFiles.value).some(file => file !== null) && uploadPath.value && !uploading.value
})

const hasSelectedFiles = computed(() => {
  return Object.values(selectedFiles.value).some(file => file !== null)
})

// Helper function to ensure folder exists before upload
const ensureFolderExists = async () => {
  try {
    // Call db_manager_api to ensure the folder exists
    // This creates the folder if it doesn't exist (without clearing contents)
    const response = await fetch(
      `${getApiBaseUrl()}/db_manager_api.php?action=ensure_folder`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          js_dir: uploadPath.value // Use original path with leading slash if present
        })
      }
    )
    
    const result = await response.json()
    if (!result.success) {
      throw new Error(result.message || 'Failed to ensure folder exists')
    }
    return true
  } catch (err) {
    // If folder creation fails, upload.php will also try to create it
    // But we log the error for debugging
    console.warn('Folder creation warning:', err.message)
    throw err // Re-throw so caller knows about the issue
  }
}

const uploadFiles = async () => {
  if (!canUpload.value) return
  
  uploading.value = true
  error.value = null
  successMessage.value = null
  
  // Reset upload results
  uploadResults.value = {
    logo: null,
    letter_head: null,
    gml2: null
  }
  
  const fileTypes = ['logo', 'letter_head', 'gml2']
  const fileNames = {
    logo: 'logo.png',
    letter_head: 'letter_head.png',
    gml2: 'gml2.png'
  }
  
  let successCount = 0
  let failCount = 0
  
  try {
    // Prepare js_dir path (remove leading slash if present)
    let jsDir = uploadPath.value.trim()
    if (jsDir.startsWith('/')) {
      jsDir = jsDir.substring(1)
    }
    jsDir = jsDir.trim()
    
    // Ensure the folder exists before uploading
    // This creates the folder if it doesn't exist (without clearing contents)
    try {
      await ensureFolderExists()
    } catch (folderErr) {
      // If folder creation fails, upload.php will also try to create it
      // But we show a warning and continue
      console.warn('Folder creation failed, upload.php will attempt to create it:', folderErr)
    }
    
    // Upload each selected file
    for (const fileType of fileTypes) {
      const file = selectedFiles.value[fileType]
      if (!file) continue
      
      try {
        const formData = new FormData()
        formData.append('file', file)
        formData.append('base_directory', jsDir)
        formData.append('destination_folder', '') // Upload to root of js_dir
        formData.append('custom_filename', fileNames[fileType]) // Use exact filename
        
        const uploadResponse = await fetch(`${getApiBaseUrl()}/upload.php`, {
          method: 'POST',
          body: formData,
        })
        
        const uploadResult = await uploadResponse.json()
        
        if (uploadResult.success) {
          uploadResults.value[fileType] = {
            success: true,
            message: `Uploaded successfully: ${uploadResult.file_path || fileNames[fileType]}`
          }
          successCount++
        } else {
          uploadResults.value[fileType] = {
            success: false,
            message: uploadResult.message || 'Upload failed'
          }
          failCount++
        }
      } catch (err) {
        uploadResults.value[fileType] = {
          success: false,
          message: err.message || 'An error occurred during upload'
        }
        failCount++
      }
    }
    
    // Show summary message
    if (successCount > 0 && failCount === 0) {
      successMessage.value = `All ${successCount} file(s) uploaded successfully`
      setTimeout(() => {
        successMessage.value = null
        clearAll()
      }, 3000)
    } else if (successCount > 0) {
      error.value = `${successCount} file(s) uploaded, ${failCount} failed`
    } else {
      error.value = 'All uploads failed'
    }
  } catch (err) {
    error.value = 'An error occurred while uploading files'
    console.error(err)
  } finally {
    uploading.value = false
  }
}

onMounted(() => {
  loadConfiguration()
})
</script>

<style scoped>
.general-settings-form {
  max-width: 800px;
  margin: 0 auto;
}

.form-header {
  margin-bottom: 2rem;
}

.form-header h2 {
  color: #1f2937;
  font-size: 1.75rem;
  margin: 0 0 0.5rem 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.form-header h2 i {
  color: #6366f1;
}

.form-description {
  color: #6b7280;
  font-size: 1rem;
  margin: 0;
}

.error-message,
.success-message {
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.error-message {
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  color: #b91c1c;
}

.success-message {
  background-color: #d1fae5;
  border: 1px solid #10b981;
  color: #065f46;
}

.loading-state {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
}

.loading-state i {
  font-size: 2rem;
  margin-bottom: 1rem;
  display: block;
}

.upload-section {
  background-color: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.info-box {
  background-color: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 2rem;
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.info-box i {
  color: #3b82f6;
  font-size: 1.25rem;
  margin-top: 0.25rem;
}

.info-box p {
  margin: 0.25rem 0;
  color: #1e40af;
}

.db-code-info {
  font-size: 0.9rem;
  color: #60a5fa !important;
}

.file-uploads {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.file-upload-item {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 1.5rem;
  background-color: #f9fafb;
}

.file-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 1rem;
  font-size: 1.1rem;
}

.file-label i {
  color: #6366f1;
}

.file-input-wrapper {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.file-input {
  display: none;
}

.browse-btn {
  padding: 0.75rem 1.5rem;
  background-color: #6366f1;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s ease;
}

.browse-btn:hover:not(:disabled) {
  background-color: #4f46e5;
  transform: translateY(-1px);
}

.browse-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.selected-file {
  margin-top: 1rem;
  padding: 0.75rem;
  background-color: white;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.selected-file i {
  color: #6366f1;
}

.selected-file span {
  flex: 1;
  color: #374151;
}

.file-size {
  color: #6b7280;
  font-size: 0.9rem;
}

.clear-btn {
  background: none;
  border: none;
  color: #ef4444;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.clear-btn:hover:not(:disabled) {
  background-color: #fee2e2;
}

.clear-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.upload-result {
  margin-top: 0.75rem;
  padding: 0.75rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
}

.upload-result.success {
  background-color: #d1fae5;
  color: #065f46;
  border: 1px solid #10b981;
}

.upload-result.error {
  background-color: #fee2e2;
  color: #b91c1c;
  border: 1px solid #ef4444;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.upload-btn {
  padding: 0.75rem 2rem;
  background-color: #6366f1;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  transition: all 0.2s ease;
}

.upload-btn:hover:not(:disabled) {
  background-color: #4f46e5;
  transform: translateY(-1px);
}

.upload-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.clear-all-btn {
  padding: 0.75rem 1.5rem;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s ease;
}

.clear-all-btn:hover:not(:disabled) {
  background-color: #4b5563;
}

.clear-all-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .general-settings-form {
    padding: 1rem;
  }

  .upload-section {
    padding: 1.5rem;
  }

  .form-actions {
    flex-direction: column;
  }

  .upload-btn,
  .clear-all-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>

