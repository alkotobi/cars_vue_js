<script setup>
import { ref, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  car: {
    type: Object,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])
const { callApi, uploadFile } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(null)
const isProcessing = ref(false)

// Upload progress tracking
const uploadProgress = ref({
  documents: 0,
  sell_pi: 0,
  buy_pi: 0,
})

// File refs
const documentsFile = ref(null)
const sellPiFile = ref(null)
const buyPiFile = ref(null)

// Drag states
const isDraggingDocuments = ref(false)
const isDraggingSellPi = ref(false)
const isDraggingBuyPi = ref(false)

const isValidFileType = (file) => {
  // Define accepted MIME types
  const acceptedTypes = {
    // PDF
    'application/pdf': true,
    // Images
    'image/jpeg': true,
    'image/png': true,
    'image/gif': true,
    'image/webp': true,
    'image/heic': true,
    // Microsoft Office
    'application/msword': true, // .doc
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document': true, // .docx
    'application/vnd.ms-excel': true, // .xls
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': true, // .xlsx
    'application/vnd.ms-powerpoint': true, // .ppt
    'application/vnd.openxmlformats-officedocument.presentationml.presentation': true, // .pptx
    // Apple iWork
    'application/x-iwork-pages-sffpages': true, // Pages
    'application/x-iwork-numbers-sffnumbers': true, // Numbers
    'application/x-iwork-keynote-sffkey': true, // Keynote
  }

  return acceptedTypes[file.type] || false
}

const getFileTypeErrorMessage = () => {
  return (
    'Only the following file types are allowed:\n' +
    '- PDF files\n' +
    '- Images (JPEG, PNG, GIF, WEBP, HEIC)\n' +
    '- Microsoft Office files (DOC, DOCX, XLS, XLSX, PPT, PPTX)\n' +
    '- Apple iWork files (Pages, Numbers, Keynote)'
  )
}

const getFileExtension = (file) => {
  // Get extension from file name
  const fileNameExt = file.name.split('.').pop().toLowerCase()

  // If we can't get it from filename, try to get it from mime type
  if (!fileNameExt) {
    const mimeToExt = {
      'application/pdf': 'pdf',
      'image/jpeg': 'jpg',
      'image/png': 'png',
      'image/gif': 'gif',
      'image/webp': 'webp',
      'image/heic': 'heic',
      'application/msword': 'doc',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'docx',
      'application/vnd.ms-excel': 'xls',
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'xlsx',
      'application/vnd.ms-powerpoint': 'ppt',
      'application/vnd.openxmlformats-officedocument.presentationml.presentation': 'pptx',
      'application/x-iwork-pages-sffpages': 'pages',
      'application/x-iwork-numbers-sffnumbers': 'numbers',
      'application/x-iwork-keynote-sffkey': 'key',
    }
    return mimeToExt[file.type] || 'pdf' // fallback to pdf if unknown
  }

  return fileNameExt
}

const handleFileUpload = async (file, type) => {
  if (!file) return null

  try {
    // Reset progress for this type
    uploadProgress.value[type] = 0

    // Get the file extension
    const fileExtension = getFileExtension(file)

    // Create FormData
    const formData = new FormData()
    formData.append('file', file)
    formData.append('destination_folder', `cars/${props.car.id}/${type}`)
    formData.append('custom_filename', `${type}_${Date.now()}.${fileExtension}`)

    // Create promise to handle XHR
    const uploadPromise = new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()

      // Handle progress
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          uploadProgress.value[type] = Math.round((event.loaded * 100) / event.total)
        }
      })

      // Handle response
      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText)
            if (response.success) {
              resolve(response)
            } else {
              reject(new Error(response.message || `Failed to upload ${type} file`))
            }
          } catch (e) {
            reject(new Error('Invalid server response'))
          }
        } else {
          reject(new Error(`HTTP Error: ${xhr.status}`))
        }
      }

      // Handle network errors
      xhr.onerror = () => {
        reject(new Error('Network error occurred'))
      }

      // Open and send request
      xhr.open('POST', '/api/upload.php')
      xhr.send(formData)
    })

    const result = await uploadPromise
    return result.file_path
  } catch (err) {
    console.error('Upload error:', err)
    throw new Error(`Error uploading ${type}: ${err.message}`)
  } finally {
    // Reset progress after completion or error
    uploadProgress.value[type] = 0
  }
}

const handleSubmit = async () => {
  if (isProcessing.value || loading.value) return
  loading.value = true
  isProcessing.value = true
  error.value = null
  success.value = null

  try {
    const updates = {}
    const uploadPromises = []

    // Prepare all uploads
    if (documentsFile.value) {
      uploadPromises.push(
        handleFileUpload(documentsFile.value, 'documents')
          .then((path) => {
            if (path) updates.path_documents = path
          })
          .catch((err) => {
            throw new Error(`Documents upload failed: ${err.message}`)
          }),
      )
    }

    if (sellPiFile.value) {
      uploadPromises.push(
        handleFileUpload(sellPiFile.value, 'sell_pi')
          .then((path) => {
            if (path) updates.sell_pi_path = path
          })
          .catch((err) => {
            throw new Error(`Sell PI upload failed: ${err.message}`)
          }),
      )
    }

    if (buyPiFile.value) {
      uploadPromises.push(
        handleFileUpload(buyPiFile.value, 'buy_pi')
          .then((path) => {
            if (path) updates.buy_pi_path = path
          })
          .catch((err) => {
            throw new Error(`Buy PI upload failed: ${err.message}`)
          }),
      )
    }

    // Wait for all uploads to complete
    await Promise.all(uploadPromises)

    // If any files were uploaded, update the database
    if (Object.keys(updates).length > 0) {
      const setClause = Object.keys(updates)
        .map((key) => `${key} = ?`)
        .join(', ')

      const result = await callApi({
        query: `UPDATE cars_stock SET ${setClause} WHERE id = ?`,
        params: [...Object.values(updates), props.car.id],
      })

      if (result.success) {
        success.value = 'Files uploaded successfully'
        emit('save', { ...props.car, ...updates })
        // Reset all file inputs
        documentsFile.value = null
        sellPiFile.value = null
        buyPiFile.value = null
        // Close the modal after a short delay to show success message
        setTimeout(() => {
          closeModal()
        }, 1000)
      } else {
        throw new Error(result.error || 'Failed to update file paths')
      }
    } else {
      success.value = 'No files selected for upload'
      // Close the modal after a short delay if no files were selected
      setTimeout(() => {
        closeModal()
      }, 1000)
    }
  } catch (err) {
    console.error('Submit error:', err)
    error.value = err.message || 'An error occurred during upload'
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const closeModal = () => {
  // Reset all states
  error.value = null
  success.value = null
  loading.value = false
  isProcessing.value = false
  documentsFile.value = null
  sellPiFile.value = null
  buyPiFile.value = null
  // Reset upload progress
  uploadProgress.value = {
    documents: 0,
    sell_pi: 0,
    buy_pi: 0,
  }
  // Emit close event
  emit('close')
}

const handleDragOver = (event, dragRef) => {
  event.preventDefault()
  if (typeof dragRef === 'object' && dragRef !== null) {
    dragRef.value = true
  }
}

const handleDragLeave = (event, dragRef) => {
  event.preventDefault()
  if (typeof dragRef === 'object' && dragRef !== null) {
    dragRef.value = false
  }
}

const handleFileChange = (event, type) => {
  const file = event.target.files?.[0]
  if (!file) return

  if (!isValidFileType(file)) {
    error.value = getFileTypeErrorMessage()
    event.target.value = ''
    return
  }

  if (file.size > 10 * 1024 * 1024) {
    // 10MB limit
    error.value = 'File size must be less than 10MB'
    event.target.value = ''
    return
  }

  error.value = null

  // Set the appropriate file ref based on type
  switch (type) {
    case 'documents':
      documentsFile.value = file
      break
    case 'sell_pi':
      sellPiFile.value = file
      break
    case 'buy_pi':
      buyPiFile.value = file
      break
  }
}

const handleDrop = (event, type, dragRef) => {
  event.preventDefault()
  if (typeof dragRef === 'object' && dragRef !== null) {
    dragRef.value = false
  }

  const file = event.dataTransfer.files?.[0]
  if (!file) return

  if (!isValidFileType(file)) {
    error.value = getFileTypeErrorMessage()
    return
  }

  if (file.size > 10 * 1024 * 1024) {
    // 10MB limit
    error.value = 'File size must be less than 10MB'
    return
  }

  error.value = null

  // Set the appropriate file ref based on type
  switch (type) {
    case 'documents':
      documentsFile.value = file
      break
    case 'sell_pi':
      sellPiFile.value = file
      break
    case 'buy_pi':
      buyPiFile.value = file
      break
  }
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content" :class="{ 'is-processing': isProcessing }">
      <div class="modal-header">
        <h3>
          <i class="fas fa-file-upload"></i>
          Upload Car Files
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="documents">
            <i class="fas fa-file-pdf"></i>
            Documents:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingDocuments,
              'is-uploading': uploadProgress.documents > 0,
            }"
            @dragover="(e) => handleDragOver(e, isDraggingDocuments)"
            @dragleave="(e) => handleDragLeave(e, isDraggingDocuments)"
            @drop="(e) => handleDrop(e, 'documents', isDraggingDocuments)"
          >
            <input
              type="file"
              id="documents"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp,.heic,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.pages,.numbers,.key,application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation,application/x-iwork-pages-sffpages,application/x-iwork-numbers-sffnumbers,application/x-iwork-keynote-sffkey"
              @change="(e) => handleFileChange(e, 'documents')"
              class="file-input"
              :disabled="isProcessing"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt"></i>
              <span>Choose or drop PDF file</span>
            </div>
            <div v-if="documentsFile" class="selected-file">
              <i class="fas fa-file-pdf"></i>
              {{ documentsFile.name }}
            </div>
            <div v-if="uploadProgress.documents > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.documents + '%' }"></div>
              <span>{{ uploadProgress.documents }}%</span>
            </div>
            <div v-if="props.car.path_documents" class="current-file">
              <i class="fas fa-check-circle"></i>
              Current: {{ props.car.path_documents.split('/').pop() }}
            </div>
          </div>
        </div>

        <div class="form-group">
          <label for="sell-pi">
            <i class="fas fa-file-invoice-dollar"></i>
            Sell PI:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingSellPi,
              'is-uploading': uploadProgress.sell_pi > 0,
            }"
            @dragover="(e) => handleDragOver(e, isDraggingSellPi)"
            @dragleave="(e) => handleDragLeave(e, isDraggingSellPi)"
            @drop="(e) => handleDrop(e, 'sell_pi', isDraggingSellPi)"
          >
            <input
              type="file"
              id="sell-pi"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp,.heic,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.pages,.numbers,.key,application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation,application/x-iwork-pages-sffpages,application/x-iwork-numbers-sffnumbers,application/x-iwork-keynote-sffkey"
              @change="(e) => handleFileChange(e, 'sell_pi')"
              class="file-input"
              :disabled="isProcessing"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt"></i>
              <span>Choose or drop PDF file</span>
            </div>
            <div v-if="sellPiFile" class="selected-file">
              <i class="fas fa-file-pdf"></i>
              {{ sellPiFile.name }}
            </div>
            <div v-if="uploadProgress.sell_pi > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.sell_pi + '%' }"></div>
              <span>{{ uploadProgress.sell_pi }}%</span>
            </div>
            <div v-if="props.car.sell_pi_path" class="current-file">
              <i class="fas fa-check-circle"></i>
              Current: {{ props.car.sell_pi_path.split('/').pop() }}
            </div>
          </div>
        </div>

        <div class="form-group">
          <label for="buy-pi">
            <i class="fas fa-file-contract"></i>
            Buy PI:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingBuyPi,
              'is-uploading': uploadProgress.buy_pi > 0,
            }"
            @dragover="(e) => handleDragOver(e, isDraggingBuyPi)"
            @dragleave="(e) => handleDragLeave(e, isDraggingBuyPi)"
            @drop="(e) => handleDrop(e, 'buy_pi', isDraggingBuyPi)"
          >
            <input
              type="file"
              id="buy-pi"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp,.heic,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.pages,.numbers,.key,application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation,application/x-iwork-pages-sffpages,application/x-iwork-numbers-sffnumbers,application/x-iwork-keynote-sffkey"
              @change="(e) => handleFileChange(e, 'buy_pi')"
              class="file-input"
              :disabled="isProcessing"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt"></i>
              <span>Choose or drop PDF file</span>
            </div>
            <div v-if="buyPiFile" class="selected-file">
              <i class="fas fa-file-pdf"></i>
              {{ buyPiFile.name }}
            </div>
            <div v-if="uploadProgress.buy_pi > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.buy_pi + '%' }"></div>
              <span>{{ uploadProgress.buy_pi }}%</span>
            </div>
            <div v-if="props.car.buy_pi_path" class="current-file">
              <i class="fas fa-check-circle"></i>
              Current: {{ props.car.buy_pi_path.split('/').pop() }}
            </div>
          </div>
        </div>

        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>
        <div v-if="success" class="success-message">
          <i class="fas fa-check-circle"></i>
          {{ success }}
        </div>
      </div>

      <div class="modal-footer">
        <button class="cancel-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
          Cancel
        </button>
        <button
          class="save-btn"
          @click="handleSubmit"
          :disabled="isProcessing || (!documentsFile && !sellPiFile && !buyPiFile)"
          :class="{ 'is-processing': isProcessing }"
        >
          <i class="fas fa-upload"></i>
          <span>{{ isProcessing ? 'Uploading...' : 'Upload Files' }}</span>
          <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.file-input-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.current-file {
  font-size: 12px;
  color: #6b7280;
}

.error-message {
  color: #ef4444;
  margin-bottom: 16px;
  font-size: 14px;
}

.success-message {
  color: #10b981;
  margin-bottom: 16px;
  font-size: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.cancel-btn,
.save-btn {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  border: none;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
}

.save-btn:hover {
  background-color: #2563eb;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.file-input-container.is-uploading {
  border-color: #3b82f6;
  background-color: rgba(59, 130, 246, 0.05);
}

.upload-progress {
  margin: 8px;
  background-color: #f3f4f6;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
  height: 24px;
}

.progress-bar {
  height: 100%;
  background-color: #3b82f6;
  transition: width 0.2s ease;
}

.upload-progress span {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-size: 12px;
  font-weight: 500;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
</style>
