<script setup>
import { ref, computed, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import { useI18n } from 'vue-i18n'
import AddItemDialog from './AddItemDialog.vue'

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'saved'])

const { callApi, uploadFile } = useApi()
const { t } = useI18n()

const newClient = ref({
  name: '',
  address: '',
  email: '',
  mobiles: '',
  id_no: '',
  nin: '',
  is_client: true,
  is_broker: false,
  notes: '',
})

const selectedFile = ref(null)
const isSubmitting = ref(false)
const validationError = ref('')

// Computed properties for validation status
const isNewClientMobileValid = computed(() => {
  return newClient.value.mobiles && validateAlgerianMobile(newClient.value.mobiles)
})

const isNewClientNINValid = computed(() => {
  return newClient.value.nin && validateNIN(newClient.value.nin)
})

const isNewClientIDValid = computed(() => {
  return newClient.value.id_no && validateIDNumber(newClient.value.id_no)
})

// Validation functions
const validateEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const validateAlgerianMobile = (mobile) => {
  // Algerian mobile format: 0552663908 (10 digits, starting with 0, second digit is 5, 6, or 7)
  const algerianMobileRegex = /^0[567]\d{8}$/
  return algerianMobileRegex.test(mobile)
}

const validateIDNumber = (idNumber) => {
  // ID Number format: exactly 9 digits, numeric only
  const idNumberRegex = /^\d{9}$/
  return idNumberRegex.test(idNumber)
}

const validateNIN = (nin) => {
  // National Identifier Number format: exactly 18 digits, numeric only
  const ninRegex = /^\d{18}$/
  return ninRegex.test(nin)
}

// Input masking functions
const handleMobileInput = (event) => {
  let value = event.target.value
  value = value.replace(/\D/g, '')
  if (value.length > 10) {
    value = value.substring(0, 10)
  }
  if (value.length === 1 && value !== '0') {
    value = '0'
  } else if (value.length === 2) {
    const secondDigit = value[1]
    if (!['5', '6', '7'].includes(secondDigit)) {
      value = value[0]
    }
  }
  newClient.value.mobiles = value
  event.target.value = value
}

const handleNINInput = (event) => {
  let value = event.target.value
  value = value.replace(/\D/g, '')
  if (value.length > 18) {
    value = value.substring(0, 18)
  }
  newClient.value.nin = value
  event.target.value = value
}

const handleIDNumberInput = (event) => {
  let value = event.target.value
  value = value.replace(/\D/g, '')
  if (value.length > 9) {
    value = value.substring(0, 9)
  }
  newClient.value.id_no = value
  event.target.value = value
}

const handleFileChange = (event) => {
  selectedFile.value = event.target.files[0] || null
}

// Watch for show prop to reset form
watch(
  () => props.show,
  (isVisible) => {
    if (isVisible) {
      resetForm()
    }
  },
)

const resetForm = () => {
  newClient.value = {
    name: '',
    address: '',
    email: '',
    mobiles: '',
    id_no: '',
    nin: '',
    is_client: true,
    is_broker: false,
    notes: '',
  }
  selectedFile.value = null
  validationError.value = ''
}

const handleClose = () => {
  resetForm()
  emit('close')
}

const handleSave = async () => {
  if (isSubmitting.value) return

  validationError.value = ''

  // Required field validation
  if (!newClient.value.mobiles) {
    validationError.value = `⚠️ ${t('mobileValidation.critical')}: ${t('mobileValidation.mobileRequired')}`
    return
  }

  if (!validateAlgerianMobile(newClient.value.mobiles)) {
    validationError.value = `⚠️ ${t('mobileValidation.critical')}: ${t('mobileValidation.invalidAlgerianMobile')}`
    return
  }

  if (!newClient.value.id_no) {
    validationError.value = `⚠️ ${t('mobileValidation.critical')}: ${t('idValidation.idRequired')}`
    return
  }

  if (!validateIDNumber(newClient.value.id_no)) {
    validationError.value = `⚠️ ${t('mobileValidation.critical')}: ${t('idValidation.invalidFormat')}`
    return
  }

  if (!newClient.value.nin) {
    validationError.value = `⚠️ ${t('mobileValidation.critical')}: ${t('ninValidation.ninRequired')}`
    return
  }

  if (!validateNIN(newClient.value.nin)) {
    validationError.value = `⚠️ ${t('mobileValidation.critical')}: ${t('ninValidation.invalidFormat')}`
    return
  }

  if (!selectedFile.value) {
    validationError.value = 'ID Document is required'
    return
  }

  // Email validation only if provided
  if (newClient.value.email && !validateEmail(newClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  isSubmitting.value = true

  try {
    // First insert the client to get the ID
    const result = await callApi({
      query: `
        INSERT INTO clients (name, address, email, mobiles, id_no, nin, is_broker, is_client, notes, id_copy_path)
        VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, NULL)
      `,
      params: [
        newClient.value.name,
        newClient.value.address,
        newClient.value.email,
        newClient.value.mobiles,
        newClient.value.id_no,
        newClient.value.nin,
        newClient.value.is_broker ? 1 : 0,
        newClient.value.notes,
      ],
    })

    if (!result.success) {
      validationError.value = result.error || 'Failed to create client'
      isSubmitting.value = false
      return
    }

    const clientId = result.lastInsertId
    let filePath = null

    // Upload file AFTER client is created
    if (selectedFile.value) {
      try {
        const filename = `${clientId}.${selectedFile.value.name.split('.').pop()}`
        const uploadResult = await uploadFile(selectedFile.value, 'ids', filename)

        if (!uploadResult.success) {
          // Upload failed - delete the client record
          await callApi({
            query: 'DELETE FROM clients WHERE id = ?',
            params: [clientId],
          })
          validationError.value =
            t('clients.failed_to_upload_id_document') ||
            'Failed to upload ID document. Client was not created. Please check your internet connection and try again.'
          isSubmitting.value = false
          return
        }

        // Update client with file path
        filePath = uploadResult.relativePath
        const updateResult = await callApi({
          query: 'UPDATE clients SET id_copy_path = ? WHERE id = ?',
          params: [filePath, clientId],
        })

        if (!updateResult.success) {
          // File uploaded but database update failed - delete client
          await callApi({
            query: 'DELETE FROM clients WHERE id = ?',
            params: [clientId],
          })
          validationError.value =
            t('clients.failed_to_save_id_document_path') || 'Failed to save ID document path. Client was not created.'
          isSubmitting.value = false
          return
        }
      } catch (err) {
        console.error('Error uploading file:', err)
        // Upload failed - delete the client record
        try {
          await callApi({
            query: 'DELETE FROM clients WHERE id = ?',
            params: [clientId],
          })
        } catch (deleteErr) {
          console.error('Error deleting client after upload failure:', deleteErr)
        }
        validationError.value =
          t('clients.failed_to_upload_id_document') ||
          'Failed to upload ID document. Client was not created. Please check your internet connection and try again.'
        isSubmitting.value = false
        return
      }
    } else {
      // No file selected, but file is required - delete the client
      await callApi({
        query: 'DELETE FROM clients WHERE id = ?',
        params: [clientId],
      })
      validationError.value = 'ID Document is required'
      isSubmitting.value = false
      return
    }

    // Emit saved event with client ID
    emit('saved', {
      id: clientId,
      name: newClient.value.name,
      address: newClient.value.address,
      email: newClient.value.email,
      mobiles: newClient.value.mobiles,
      id_no: newClient.value.id_no,
      nin: newClient.value.nin,
      is_broker: newClient.value.is_broker ? 1 : 0,
      is_client: 1,
      notes: newClient.value.notes,
      id_copy_path: filePath,
    })

    resetForm()
    emit('close')
  } catch (err) {
    validationError.value = err.message || 'Failed to add client'
    console.error('Error in add client process:', err)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AddItemDialog
    :show="show"
    :title="t('carStockForm.addClient') || 'Add New Client'"
    :loading="isSubmitting || loading"
    max-width="900px"
    min-width="600px"
    @close="handleClose"
    @save="handleSave"
  >
    <div v-if="validationError" class="validation-error">
      <i class="fas fa-exclamation-triangle"></i>
      {{ validationError }}
    </div>

    <div class="form-grid">
      <div class="form-group">
        <label for="client_name">
          <i class="fas fa-user"></i>
          Full Name *
        </label>
        <input
          id="client_name"
          v-model="newClient.name"
          type="text"
          placeholder="Enter client's full name"
          class="form-input"
          :class="{ error: validationError && !newClient.name }"
          :disabled="isSubmitting || loading"
          required
        />
      </div>

      <div class="form-group">
        <label for="client_mobile">
          <i class="fas fa-phone"></i>
          {{ t('mobileValidation.mobileLabel') }}
          <span v-if="!isNewClientMobileValid" class="critical-field"
            >⚠️ {{ t('mobileValidation.critical') }}</span
          >
          <span v-else class="success-field">✅ {{ t('mobileValidation.valid') }}</span>
        </label>
        <input
          id="client_mobile"
          v-model="newClient.mobiles"
          type="tel"
          :placeholder="t('mobileValidation.mobilePlaceholder')"
          class="form-input"
          :class="{
            'critical-input': !isNewClientMobileValid,
            'success-input': isNewClientMobileValid,
            error: validationError && !newClient.mobiles,
          }"
          :disabled="isSubmitting || loading"
          @input="handleMobileInput"
          maxlength="10"
          required
        />
        <div v-if="!isNewClientMobileValid" class="field-help">
          <i class="fas fa-exclamation-triangle"></i>
          {{ t('mobileValidation.helpText') }}
        </div>
        <div v-else class="field-success">
          <i class="fas fa-check-circle"></i>
          {{ t('mobileValidation.validFormat') }}
        </div>
      </div>

      <div class="form-group">
        <label for="client_id_no">
          <i class="fas fa-id-card"></i>
          {{ t('idValidation.idLabel') }}
          <span v-if="!isNewClientIDValid" class="critical-field"
            >⚠️ {{ t('mobileValidation.critical') }}</span
          >
          <span v-else class="success-field">✅ {{ t('idValidation.valid') }}</span>
        </label>
        <input
          id="client_id_no"
          v-model="newClient.id_no"
          type="text"
          :placeholder="t('idValidation.idPlaceholder')"
          class="form-input"
          :class="{
            'critical-input': !isNewClientIDValid,
            'success-input': isNewClientIDValid,
            error: validationError && !newClient.id_no,
          }"
          :disabled="isSubmitting || loading"
          @input="handleIDNumberInput"
          maxlength="9"
          required
        />
        <div v-if="!isNewClientIDValid" class="field-help">
          <i class="fas fa-exclamation-triangle"></i>
          {{ t('idValidation.helpText') }}
        </div>
        <div v-else class="field-success">
          <i class="fas fa-check-circle"></i>
          {{ t('idValidation.validFormat') }}
        </div>
      </div>

      <div class="form-group">
        <label for="client_nin">
          <i class="fas fa-passport"></i>
          {{ t('ninValidation.ninLabel') }}
          <span v-if="!isNewClientNINValid" class="critical-field"
            >⚠️ {{ t('mobileValidation.critical') }}</span
          >
          <span v-else class="success-field">✅ {{ t('ninValidation.valid') }}</span>
        </label>
        <input
          id="client_nin"
          v-model="newClient.nin"
          type="text"
          :placeholder="t('ninValidation.ninPlaceholder')"
          class="form-input"
          :class="{
            'critical-input': !isNewClientNINValid,
            'success-input': isNewClientNINValid,
            error: validationError && !newClient.nin,
          }"
          :disabled="isSubmitting || loading"
          @input="handleNINInput"
          maxlength="18"
          required
        />
        <div v-if="!isNewClientNINValid" class="field-help">
          <i class="fas fa-exclamation-triangle"></i>
          {{ t('ninValidation.helpText') }}
        </div>
        <div v-else class="field-success">
          <i class="fas fa-check-circle"></i>
          {{ t('ninValidation.validFormat') }}
        </div>
      </div>

      <div class="form-group">
        <label for="client_email">
          <i class="fas fa-envelope"></i>
          Email Address
        </label>
        <input
          id="client_email"
          v-model="newClient.email"
          type="email"
          placeholder="Enter email address (optional)"
          class="form-input"
          :class="{ error: validationError && newClient.email }"
          :disabled="isSubmitting || loading"
        />
      </div>

      <div class="form-group full-width">
        <label for="client_address">
          <i class="fas fa-map-marker-alt"></i>
          Address
        </label>
        <input
          id="client_address"
          v-model="newClient.address"
          type="text"
          placeholder="Enter client's address"
          class="form-input"
          :disabled="isSubmitting || loading"
        />
      </div>

      <div class="form-group">
        <label for="client_is_broker" class="checkbox-label">
          <input
            id="client_is_broker"
            type="checkbox"
            v-model="newClient.is_broker"
            :disabled="isSubmitting || loading"
            class="checkbox-input"
          />
          <span class="checkbox-custom"></span>
          <i class="fas fa-user-tie"></i>
          Is Broker
        </label>
      </div>

      <div class="form-group full-width">
        <label for="client_notes">
          <i class="fas fa-sticky-note"></i>
          Notes
        </label>
        <textarea
          id="client_notes"
          v-model="newClient.notes"
          placeholder="Add any additional notes about this client..."
          rows="3"
          class="form-textarea"
          :disabled="isSubmitting || loading"
        ></textarea>
      </div>

      <div class="form-group full-width">
        <label for="client_id_document">
          <i class="fas fa-file-upload"></i>
          ID Document *
        </label>
        <div class="file-upload-area">
          <input
            type="file"
            id="client_id_document"
            @change="handleFileChange"
            accept="image/*,.pdf"
            class="file-input"
            :class="{ error: validationError && !selectedFile }"
            :disabled="isSubmitting || loading"
            required
          />
          <div class="file-upload-content">
            <i class="fas fa-cloud-upload-alt"></i>
            <p>Click to upload or drag and drop</p>
            <span>Supports: JPG, PNG, PDF (Max 5MB)</span>
          </div>
        </div>
        <div v-if="selectedFile" class="selected-file">
          <i class="fas fa-check-circle"></i>
          <span>{{ selectedFile.name }}</span>
        </div>
      </div>
    </div>
  </AddItemDialog>
</template>

<style scoped>

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.full-width {
  grid-column: span 2;
}

.form-group label {
  font-weight: 600;
  color: #374151;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
}

.form-input,
.form-textarea {
  padding: 12px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s ease;
  background: white;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input.error,
.form-textarea.error {
  border-color: #ef4444;
  background: #fef2f2;
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
  font-family: inherit;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 12px;
  border-radius: 8px;
  transition: background-color 0.2s ease;
}

.checkbox-label:hover {
  background: #f9fafb;
}

.checkbox-input {
  width: 18px;
  height: 18px;
  accent-color: #3b82f6;
}

.checkbox-label i {
  color: #6b7280;
}

.file-upload-area {
  position: relative;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 32px 16px;
  text-align: center;
  transition: all 0.2s ease;
  background: #f9fafb;
}

.file-upload-area:hover {
  border-color: #3b82f6;
  background: #eff6ff;
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

.file-upload-content {
  pointer-events: none;
}

.file-upload-content i {
  font-size: 2rem;
  color: #6b7280;
  margin-bottom: 12px;
}

.file-upload-content p {
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
  margin: 0 0 8px 0;
}

.file-upload-content span {
  font-size: 0.9rem;
  color: #6b7280;
}

.selected-file {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%);
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  color: #166534;
  font-weight: 500;
  margin-top: 8px;
}

.selected-file i {
  color: #16a34a;
}

.validation-error {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 500;
  margin-bottom: 16px;
}

.validation-error i {
  margin-right: 8px;
}

/* Critical Field Styles */
.critical-field {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  margin-left: 8px;
  box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
  animation: pulse-critical 2s infinite;
}

@keyframes pulse-critical {
  0%,
  100% {
    box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
  }
  50% {
    box-shadow: 0 4px 8px rgba(239, 68, 68, 0.5);
  }
}

.critical-input {
  border: 2px solid #fca5a5 !important;
  background: linear-gradient(135deg, #fef2f2 0%, #fef7f7 100%) !important;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1) !important;
}

.critical-input:focus {
  border-color: #ef4444 !important;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.2) !important;
}

/* Success Field Styles */
.success-field {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  margin-left: 8px;
  box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);
  animation: pulse-success 2s infinite;
}

@keyframes pulse-success {
  0%,
  100% {
    box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);
  }
  50% {
    box-shadow: 0 4px 8px rgba(16, 185, 129, 0.5);
  }
}

.success-input {
  border: 2px solid #6ee7b7 !important;
  background: linear-gradient(135deg, #f0fdf4 0%, #f7fef8 100%) !important;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1) !important;
}

.success-input:focus {
  border-color: #10b981 !important;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.2) !important;
}

.field-success {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #d1fae5 0%, #ecfdf5 100%);
  border: 1px solid #10b981;
  border-radius: 6px;
  color: #065f46;
  font-size: 0.85rem;
  font-weight: 500;
  animation: slideInSuccess 0.3s ease-out;
}

@keyframes slideInSuccess {
  0% {
    opacity: 0;
    transform: translateY(-10px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.field-help {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #fef3c7 0%, #fef7cd 100%);
  border: 1px solid #f59e0b;
  border-radius: 6px;
  color: #92400e;
  font-size: 0.85rem;
  font-weight: 500;
  animation: slideInHelp 0.3s ease-out;
}

@keyframes slideInHelp {
  from {
    opacity: 0;
    transform: translateY(-5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.field-help i {
  color: #f59e0b;
  font-size: 0.9rem;
}

@media (max-width: 768px) {

  .form-grid {
    grid-template-columns: 1fr;
  }

  .form-group.full-width {
    grid-column: span 1;
  }
}
</style>
