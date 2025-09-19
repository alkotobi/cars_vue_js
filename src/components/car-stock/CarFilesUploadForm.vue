<script setup>
import { ref, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

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

// Check if current user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Upload progress tracking
const uploadProgress = ref({
  documents: 0,
  sell_pi: 0,
  buy_pi: 0,
  coo: 0,
  coc: 0,
})

// File refs
const documentsFile = ref(null)
const sellPiFile = ref(null)
const buyPiFile = ref(null)
const cooFile = ref(null)
const cocFile = ref(null)

// Drag states
const isDraggingDocuments = ref(false)
const isDraggingSellPi = ref(false)
const isDraggingBuyPi = ref(false)
const isDraggingCoo = ref(false)
const isDraggingCoc = ref(false)

const isUploading = ref({
  documents: false,
  sell_pi: false,
  buy_pi: false,
  coo: false,
  coc: false,
})

const resetUploadStates = () => {
  isUploading.value = {
    documents: false,
    sell_pi: false,
    buy_pi: false,
    coo: false,
    coc: false,
  }
}

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
  return t('carFilesUploadForm.onlyFollowingFileTypesAllowed')
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
  if (!file || isUploading.value[type]) return null

  try {
    isUploading.value[type] = true
    // Reset progress for this type
    uploadProgress.value[type] = 0

    // Get the file extension
    const fileExtension = getFileExtension(file)
    const filename = `${type}_${Date.now()}.${fileExtension}`

    const result = await uploadFile(file, `cars/${props.car.id}/${type}`, filename)

    if (result.success) {
      return result.relativePath
    } else {
      throw new Error(result.error || `Failed to upload ${type} file`)
    }
  } catch (err) {
    console.error('Upload error:', err)
    throw new Error(`Error uploading ${type}: ${err.message}`)
  } finally {
    // Reset progress and uploading state after completion or error
    uploadProgress.value[type] = 0
    isUploading.value[type] = false
  }
}

const handleRevertFile = async (fileType) => {
  if (!isAdmin.value) {
    error.value = t('carFilesUploadForm.onlyAdminCanRevertFileUploads')
    return
  }

  // Determine which file path to check and clear based on file type
  let filePath = null
  let dbField = null
  let displayName = null

  switch (fileType) {
    case 'documents':
      filePath = props.car.path_documents
      dbField = 'path_documents'
      displayName = t('carFilesUploadForm.bl')
      break
    case 'sell_pi':
      filePath = props.car.sell_pi_path
      dbField = 'sell_pi_path'
      displayName = t('carFilesUploadForm.invoice')
      break
    case 'buy_pi':
      filePath = props.car.buy_pi_path
      dbField = 'buy_pi_path'
      displayName = t('carFilesUploadForm.packingList')
      break
    case 'coo':
      filePath = props.car.path_coo
      dbField = 'path_coo'
      displayName = t('carFilesUploadForm.coo')
      break
    case 'coc':
      filePath = props.car.path_coc
      dbField = 'path_coc'
      displayName = t('carFilesUploadForm.coc')
      break
    default:
      error.value = 'Invalid file type'
      return
  }

  if (!filePath) {
    error.value = `No ${displayName} file to revert`
    return
  }

  const confirmed = confirm(
    `Are you sure you want to remove the ${displayName} file? This action cannot be undone.`,
  )
  if (!confirmed) return

  try {
    loading.value = true
    error.value = null

    const result = await callApi({
      query: `UPDATE cars_stock SET ${dbField} = NULL WHERE id = ?`,
      params: [props.car.id],
    })

    if (result.success) {
      // Update the car object
      const updatedCar = {
        ...props.car,
        [dbField]: null,
      }
      Object.assign(props.car, updatedCar)

      success.value = `${displayName} file removed successfully`
      emit('save', updatedCar)

      // Close the modal after a short delay to show success message
      setTimeout(() => {
        closeModal()
      }, 1000)
    } else {
      throw new Error(result.error || `Failed to remove ${displayName} file`)
    }
  } catch (err) {
    console.error('Revert error:', err)
    error.value = err.message || 'An error occurred while removing the file'
  } finally {
    loading.value = false
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
            throw new Error(t('carFilesUploadForm.documentsUploadFailed') + `: ${err.message}`)
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
            throw new Error(t('carFilesUploadForm.sellPiUploadFailed') + `: ${err.message}`)
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
            throw new Error(t('carFilesUploadForm.buyPiUploadFailed') + `: ${err.message}`)
          }),
      )
    }

    if (cooFile.value) {
      uploadPromises.push(
        handleFileUpload(cooFile.value, 'coo')
          .then((path) => {
            if (path) updates.path_coo = path
          })
          .catch((err) => {
            throw new Error(t('carFilesUploadForm.cooUploadFailed') + `: ${err.message}`)
          }),
      )
    }

    if (cocFile.value) {
      uploadPromises.push(
        handleFileUpload(cocFile.value, 'coc')
          .then((path) => {
            if (path) updates.path_coc = path
          })
          .catch((err) => {
            throw new Error(t('carFilesUploadForm.cocUploadFailed') + `: ${err.message}`)
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
        success.value = t('carFilesUploadForm.filesUploadedSuccessfully')
        emit('save', { ...props.car, ...updates })
        // Reset all file inputs
        documentsFile.value = null
        sellPiFile.value = null
        buyPiFile.value = null
        cooFile.value = null
        cocFile.value = null
        // Close the modal after a short delay to show success message
        setTimeout(() => {
          closeModal()
        }, 1000)
      } else {
        throw new Error(result.error || t('carFilesUploadForm.failedToUpdateFilePaths'))
      }
    } else {
      success.value = t('carFilesUploadForm.noFilesSelectedForUpload')
      // Close the modal after a short delay if no files were selected
      setTimeout(() => {
        closeModal()
      }, 1000)
    }
  } catch (err) {
    console.error('Submit error:', err)
    error.value = err.message || t('carFilesUploadForm.anErrorOccurredDuringUpload')
  } finally {
    loading.value = false
    isProcessing.value = false
    resetUploadStates()
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
  cooFile.value = null
  cocFile.value = null
  // Reset upload progress
  uploadProgress.value = {
    documents: 0,
    sell_pi: 0,
    buy_pi: 0,
    coo: 0,
    coc: 0,
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
    error.value = t('carFilesUploadForm.fileSizeMustBeLessThan10mb')
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
    case 'coo':
      cooFile.value = file
      break
    case 'coc':
      cocFile.value = file
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
    error.value = t('carFilesUploadForm.fileSizeMustBeLessThan10mb')
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
    case 'coo':
      cooFile.value = file
      break
    case 'coc':
      cocFile.value = file
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
          {{ t('carFilesUploadForm.uploadCarFiles') }}
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="documents">
            <i class="fas fa-file-pdf"></i>
            {{ t('carFilesUploadForm.bl') }}:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingDocuments,
              'is-uploading': uploadProgress.documents > 0 || isUploading.documents,
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
              :disabled="isProcessing || isUploading.documents"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt fa-2x"></i>
              <span>{{ t('carFilesUploadForm.chooseOrDropFile') }}</span>
              <i v-if="isUploading.documents" class="fas fa-spinner fa-spin"></i>
            </div>
            <div v-if="documentsFile" class="selected-file">
              <i class="fas fa-file-alt"></i>
              {{ documentsFile.name }}
            </div>
            <div v-if="uploadProgress.documents > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.documents + '%' }"></div>
              <span>{{ uploadProgress.documents }}%</span>
            </div>
            <div v-if="props.car.path_documents" class="current-file">
              <i class="fas fa-check-circle text-success"></i>
              {{ t('carFilesUploadForm.current') }}: {{ props.car.path_documents.split('/').pop() }}
            </div>
          </div>
          <!-- Revert button for admin only - outside file input container -->
          <div v-if="isAdmin && props.car.path_documents" class="revert-container">
            <button
              class="revert-btn"
              @click="handleRevertFile('documents')"
              :disabled="loading"
              title="Remove documents file"
            >
              <i class="fas fa-trash"></i>
              {{ t('carFilesUploadForm.removeBlFile') }}
            </button>
          </div>
        </div>

        <div class="form-group">
          <label for="sell-pi">
            <i class="fas fa-file-invoice-dollar"></i>
            {{ t('carFilesUploadForm.invoice') }}:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingSellPi,
              'is-uploading': uploadProgress.sell_pi > 0 || isUploading.sell_pi,
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
              :disabled="isProcessing || isUploading.sell_pi"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt fa-2x"></i>
              <span>{{ t('carFilesUploadForm.chooseOrDropFile') }}</span>
              <i v-if="isUploading.sell_pi" class="fas fa-spinner fa-spin"></i>
            </div>
            <div v-if="sellPiFile" class="selected-file">
              <i class="fas fa-file-alt"></i>
              {{ sellPiFile.name }}
            </div>
            <div v-if="uploadProgress.sell_pi > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.sell_pi + '%' }"></div>
              <span>{{ uploadProgress.sell_pi }}%</span>
            </div>
            <div v-if="props.car.sell_pi_path" class="current-file">
              <i class="fas fa-check-circle text-success"></i>
              {{ t('carFilesUploadForm.current') }}: {{ props.car.sell_pi_path.split('/').pop() }}
            </div>
          </div>
          <!-- Revert button for admin only - outside file input container -->
          <div v-if="isAdmin && props.car.sell_pi_path" class="revert-container">
            <button
              class="revert-btn"
              @click="handleRevertFile('sell_pi')"
              :disabled="loading"
              title="Remove sell PI file"
            >
              <i class="fas fa-trash"></i>
              {{ t('carFilesUploadForm.removeInvoiceFile') }}
            </button>
          </div>
        </div>

        <div class="form-group">
          <label for="buy-pi">
            <i class="fas fa-file-contract"></i>
            {{ t('carFilesUploadForm.packingList') }}:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingBuyPi,
              'is-uploading': uploadProgress.buy_pi > 0 || isUploading.buy_pi,
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
              :disabled="isProcessing || isUploading.buy_pi"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt fa-2x"></i>
              <span>{{ t('carFilesUploadForm.chooseOrDropFile') }}</span>
              <i v-if="isUploading.buy_pi" class="fas fa-spinner fa-spin"></i>
            </div>
            <div v-if="buyPiFile" class="selected-file">
              <i class="fas fa-file-alt"></i>
              {{ buyPiFile.name }}
            </div>
            <div v-if="uploadProgress.buy_pi > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.buy_pi + '%' }"></div>
              <span>{{ uploadProgress.buy_pi }}%</span>
            </div>
            <div v-if="props.car.buy_pi_path" class="current-file">
              <i class="fas fa-check-circle text-success"></i>
              {{ t('carFilesUploadForm.current') }}: {{ props.car.buy_pi_path.split('/').pop() }}
            </div>
          </div>
          <!-- Revert button for admin only - outside file input container -->
          <div v-if="isAdmin && props.car.buy_pi_path" class="revert-container">
            <button
              class="revert-btn"
              @click="handleRevertFile('buy_pi')"
              :disabled="loading"
              title="Remove buy PI file"
            >
              <i class="fas fa-trash"></i>
              {{ t('carFilesUploadForm.removePackingListFile') }}
            </button>
          </div>
        </div>

        <div class="form-group">
          <label for="coo">
            <i class="fas fa-certificate"></i>
            {{ t('carFilesUploadForm.coo') }}:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingCoo,
              'is-uploading': uploadProgress.coo > 0 || isUploading.coo,
            }"
            @dragover="(e) => handleDragOver(e, isDraggingCoo)"
            @dragleave="(e) => handleDragLeave(e, isDraggingCoo)"
            @drop="(e) => handleDrop(e, 'coo', isDraggingCoo)"
          >
            <input
              type="file"
              id="coo"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp,.heic,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.pages,.numbers,.key,application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation,application/x-iwork-pages-sffpages,application/x-iwork-numbers-sffnumbers,application/x-iwork-keynote-sffkey"
              @change="(e) => handleFileChange(e, 'coo')"
              class="file-input"
              :disabled="isProcessing || isUploading.coo"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt fa-2x"></i>
              <span>{{ t('carFilesUploadForm.chooseOrDropFile') }}</span>
              <i v-if="isUploading.coo" class="fas fa-spinner fa-spin"></i>
            </div>
            <div v-if="cooFile" class="selected-file">
              <i class="fas fa-file-alt"></i>
              {{ cooFile.name }}
            </div>
            <div v-if="uploadProgress.coo > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.coo + '%' }"></div>
              <span>{{ uploadProgress.coo }}%</span>
            </div>
            <div v-if="props.car.path_coo" class="current-file">
              <i class="fas fa-check-circle text-success"></i>
              {{ t('carFilesUploadForm.current') }}: {{ props.car.path_coo.split('/').pop() }}
            </div>
          </div>
          <!-- Revert button for admin only - outside file input container -->
          <div v-if="isAdmin && props.car.path_coo" class="revert-container">
            <button
              class="revert-btn"
              @click="handleRevertFile('coo')"
              :disabled="loading"
              title="Remove COO file"
            >
              <i class="fas fa-trash"></i>
              {{ t('carFilesUploadForm.removeCooFile') }}
            </button>
          </div>
        </div>

        <div class="form-group">
          <label for="coc">
            <i class="fas fa-award"></i>
            {{ t('carFilesUploadForm.coc') }}:
          </label>
          <div
            class="file-input-container"
            :class="{
              'is-dragging': isDraggingCoc,
              'is-uploading': uploadProgress.coc > 0 || isUploading.coc,
            }"
            @dragover="(e) => handleDragOver(e, isDraggingCoc)"
            @dragleave="(e) => handleDragLeave(e, isDraggingCoc)"
            @drop="(e) => handleDrop(e, 'coc', isDraggingCoc)"
          >
            <input
              type="file"
              id="coc"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp,.heic,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.pages,.numbers,.key,application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation,application/x-iwork-pages-sffpages,application/x-iwork-numbers-sffnumbers,application/x-iwork-keynote-sffkey"
              @change="(e) => handleFileChange(e, 'coc')"
              class="file-input"
              :disabled="isProcessing || isUploading.coc"
            />
            <div class="file-input-overlay">
              <i class="fas fa-cloud-upload-alt fa-2x"></i>
              <span>{{ t('carFilesUploadForm.chooseOrDropFile') }}</span>
              <i v-if="isUploading.coc" class="fas fa-spinner fa-spin"></i>
            </div>
            <div v-if="cocFile" class="selected-file">
              <i class="fas fa-file-alt"></i>
              {{ cocFile.name }}
            </div>
            <div v-if="uploadProgress.coc > 0" class="upload-progress">
              <div class="progress-bar" :style="{ width: uploadProgress.coc + '%' }"></div>
              <span>{{ uploadProgress.coc }}%</span>
            </div>
            <div v-if="props.car.path_coc" class="current-file">
              <i class="fas fa-check-circle text-success"></i>
              {{ t('carFilesUploadForm.current') }}: {{ props.car.path_coc.split('/').pop() }}
            </div>
          </div>
          <!-- Revert button for admin only - outside file input container -->
          <div v-if="isAdmin && props.car.path_coc" class="revert-container">
            <button
              class="revert-btn"
              @click="handleRevertFile('coc')"
              :disabled="loading"
              title="Remove COC file"
            >
              <i class="fas fa-trash"></i>
              {{ t('carFilesUploadForm.removeCocFile') }}
            </button>
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
          {{ t('carFilesUploadForm.cancel') }}
        </button>
        <button
          class="save-btn"
          @click="handleSubmit"
          :disabled="
            isProcessing || (!documentsFile && !sellPiFile && !buyPiFile && !cooFile && !cocFile)
          "
          :class="{ 'is-processing': isProcessing }"
        >
          <i class="fas fa-upload"></i>
          <span>{{
            isProcessing ? t('carFilesUploadForm.uploading') : t('carFilesUploadForm.uploadFiles')
          }}</span>
          <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
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
  max-height: 90vh;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
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

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding-right: 10px;
  margin-right: -10px;
}

/* Custom scrollbar styling */
.modal-body::-webkit-scrollbar {
  width: 6px;
}

.modal-body::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
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
  position: relative;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 20px;
  text-align: center;
  transition: all 0.3s ease;
  background-color: #f9fafb;
  cursor: pointer;
}

.file-input-container:hover {
  border-color: #3b82f6;
  background-color: #f0f9ff;
}

.file-input-container.is-dragging {
  border-color: #3b82f6;
  background-color: #e0f2fe;
}

.file-input-container.is-uploading {
  border-color: #6366f1;
  background-color: #eef2ff;
}

.file-input {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
}

.file-input:disabled {
  cursor: not-allowed;
}

.file-input-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #6b7280;
}

.file-input-overlay i {
  color: #3b82f6;
}

.selected-file {
  margin-top: 12px;
  padding: 8px;
  background-color: #f3f4f6;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.875rem;
}

.selected-file i {
  color: #3b82f6;
}

.upload-progress {
  margin-top: 12px;
  background-color: #f3f4f6;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
  height: 24px;
}

.progress-bar {
  height: 100%;
  background-color: #3b82f6;
  transition: width 0.3s ease;
}

.upload-progress span {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: #1f2937;
  font-size: 12px;
  font-weight: 500;
  z-index: 1;
}

.current-file {
  margin-top: 8px;
  font-size: 0.875rem;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 8px;
}

.text-success {
  color: #10b981;
}

.revert-container {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

.revert-btn {
  background-color: #dc2626;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 6px;
}

.revert-btn:hover:not(:disabled) {
  background-color: #b91c1c;
  transform: scale(1.05);
}

.revert-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  color: #ef4444;
  margin: 16px 0;
  padding: 12px;
  background-color: #fef2f2;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.success-message {
  color: #10b981;
  margin: 16px 0;
  padding: 12px;
  background-color: #f0fdf4;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  flex-shrink: 0;
  border-top: 1px solid #e5e7eb;
  padding-top: 16px;
}

.cancel-btn,
.save-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.save-btn.is-processing {
  background-color: #6366f1;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.fa-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
