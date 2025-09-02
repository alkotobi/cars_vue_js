<script setup>
import { ref,   computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { ElMessageBox } from 'element-plus'
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

// Check if current user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  const adminStatus = currentUser.value?.role_id === 1
  console.log(
    'Admin check - currentUser:',
    currentUser.value,
    'role_id:',
    currentUser.value?.role_id,
    'isAdmin:',
    adminStatus,
  )
  return adminStatus
})

const emit = defineEmits(['close', 'save', 'documents-sent'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(null)

// Form field for export license
const exportLicenseRef = ref(props.car.export_lisence_ref || '')

// Date fields for each action
const blDate = ref(new Date().toISOString().split('T')[0])
const freightDate = ref(new Date().toISOString().split('T')[0])
const supplierDocsDate = ref(new Date().toISOString().split('T')[0])
const clientDocsDate = ref(new Date().toISOString().split('T')[0])

// Handle BL received
const handleGetBL = async () => {
  try {
    const confirmed = confirm(t('carDocumentsForm.confirmMarkBlAsReceived'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_get_bl = ? WHERE id = ?',
      params: [blDate.value, props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.blReceivedDateUpdated')
      const updatedCar = {
        ...props.car,
        date_get_bl: blDate.value,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToUpdateBlReceivedDate'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

// Handle freight payment
const handleFreightPaid = async () => {
  try {
    const confirmed = confirm(t('carDocumentsForm.confirmMarkFreightAsPaid'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_pay_freight = ? WHERE id = ?',
      params: [freightDate.value, props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.freightPaymentDateUpdated')
      const updatedCar = {
        ...props.car,
        date_pay_freight: freightDate.value,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToUpdateFreightPaymentDate'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

// Handle documents received from supplier
const handleDocsFromSupplier = async () => {
  try {
    const confirmed = confirm(t('carDocumentsForm.confirmMarkDocsAsReceivedFromSupplier'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_get_documents_from_supp = ? WHERE id = ?',
      params: [supplierDocsDate.value, props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.documentsFromSupplierDateUpdated')
      const updatedCar = {
        ...props.car,
        date_get_documents_from_supp: supplierDocsDate.value,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToUpdateDocumentsFromSupplierDate'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

// Handle documents sent to client
const handleDocsSentToClient = async () => {
  try {
    const confirmed = confirm(t('carDocumentsForm.confirmMarkDocsAsSentToClient'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_send_documents = ? WHERE id = ?',
      params: [clientDocsDate.value, props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.documentsSentToClientDateUpdated')
      const updatedCar = {
        ...props.car,
        date_send_documents: clientDocsDate.value,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
      emit('documents-sent', updatedCar)
      setTimeout(() => {
        emit('close')
      }, 1500)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToUpdateDocumentsSentDate'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

// Handle export license update
const handleExportLicenseUpdate = async () => {
  try {
    const confirmed = confirm(t('carDocumentsForm.confirmUpdateExportLicenseReference'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET export_lisence_ref = ? WHERE id = ?',
      params: [exportLicenseRef.value, props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.exportLicenseReferenceUpdated')
      const updatedCar = {
        ...props.car,
        export_lisence_ref: exportLicenseRef.value,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToUpdateExportLicenseReference'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

// Revert functions for admin
const handleRevertBL = async () => {
  console.log('handleRevertBL called, isAdmin:', isAdmin.value)
  if (!isAdmin.value) return

  try {
    const confirmed = confirm(t('carDocumentsForm.confirmRevertBlReceivedStatus'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_get_bl = NULL WHERE id = ?',
      params: [props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.blReceivedStatusReverted')
      const updatedCar = {
        ...props.car,
        date_get_bl: null,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToRevertBlReceivedStatus'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

const handleRevertFreight = async () => {
  console.log('handleRevertFreight called, isAdmin:', isAdmin.value)
  if (!isAdmin.value) return

  try {
    const confirmed = confirm(t('carDocumentsForm.confirmRevertFreightPaymentStatus'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_pay_freight = NULL WHERE id = ?',
      params: [props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.freightPaymentStatusReverted')
      const updatedCar = {
        ...props.car,
        date_pay_freight: null,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToRevertFreightPaymentStatus'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

const handleRevertSupplierDocs = async () => {
  console.log('handleRevertSupplierDocs called, isAdmin:', isAdmin.value)
  if (!isAdmin.value) return

  try {
    const confirmed = confirm(t('carDocumentsForm.confirmRevertSupplierDocumentsReceivedStatus'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_get_documents_from_supp = NULL WHERE id = ?',
      params: [props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.supplierDocumentsStatusReverted')
      const updatedCar = {
        ...props.car,
        date_get_documents_from_supp: null,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToRevertSupplierDocumentsStatus'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

const handleRevertClientDocs = async () => {
  console.log('handleRevertClientDocs called, isAdmin:', isAdmin.value)
  if (!isAdmin.value) return

  try {
    const confirmed = confirm(t('carDocumentsForm.confirmRevertClientDocumentsSentStatus'))
    if (!confirmed) return

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET date_send_documents = NULL WHERE id = ?',
      params: [props.car.id],
    })

    if (result.success) {
      success.value = t('carDocumentsForm.clientDocumentsSentStatusReverted')
      const updatedCar = {
        ...props.car,
        date_send_documents: null,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || t('carDocumentsForm.failedToRevertClientDocumentsSentStatus'))
    }
  } catch (err) {
    error.value = err.message || t('carDocumentsForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  exportLicenseRef.value = props.car.export_lisence_ref || ''
  // Reset date fields to today
  blDate.value = new Date().toISOString().split('T')[0]
  freightDate.value = new Date().toISOString().split('T')[0]
  supplierDocsDate.value = new Date().toISOString().split('T')[0]
  clientDocsDate.value = new Date().toISOString().split('T')[0]
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ t('carDocumentsForm.carDocumentsManagement') }}</h3>
        <button class="close-btn" @click="closeModal">&times;</button>
      </div>

      <div class="modal-body">
        <!-- Document Status Grid -->
        <div class="document-grid">
          <!-- BL Section -->
          <div class="document-section">
            <h4>{{ t('carDocumentsForm.billOfLading') }}</h4>
            <div class="status-info">
              <span class="status-label">{{ t('carDocumentsForm.status') }}:</span>
              <span :class="['status-value', props.car.date_get_bl ? 'received' : 'pending']">
                {{
                  props.car.date_get_bl
                    ? t('carDocumentsForm.received')
                    : t('carDocumentsForm.pending')
                }}
              </span>
            </div>
            <div v-if="props.car.date_get_bl" class="date-info">
              <span class="date-label">{{ t('carDocumentsForm.receivedDate') }}:</span>
              <span class="date-value">{{
                new Date(props.car.date_get_bl).toLocaleDateString()
              }}</span>
            </div>
            <div v-if="!props.car.date_get_bl" class="date-input-group">
              <label for="bl-date">{{ t('carDocumentsForm.date') }}:</label>
              <input
                type="date"
                id="bl-date"
                v-model="blDate"
                class="date-input"
                :disabled="loading"
              />
            </div>
            <div class="button-group">
              <button
                class="action-btn get-bl-btn"
                @click="handleGetBL"
                :disabled="loading || !!props.car.date_get_bl"
                :class="{ disabled: !!props.car.date_get_bl }"
              >
                {{ loading ? t('carDocumentsForm.processing') : t('carDocumentsForm.weGetBl') }}
              </button>
              <button
                v-if="isAdmin && props.car.date_get_bl"
                class="revert-btn"
                @click="handleRevertBL"
                :disabled="loading"
                title="Revert BL received status"
              >
                ↶
              </button>
            </div>
          </div>

          <!-- Freight Payment Section -->
          <div class="document-section">
            <h4>{{ t('carDocumentsForm.freightPayment') }}</h4>
            <div class="status-info">
              <span class="status-label">{{ t('carDocumentsForm.status') }}:</span>
              <span :class="['status-value', props.car.date_pay_freight ? 'paid' : 'pending']">
                {{
                  props.car.date_pay_freight
                    ? t('carDocumentsForm.paid')
                    : t('carDocumentsForm.pending')
                }}
              </span>
            </div>
            <div v-if="props.car.date_pay_freight" class="date-info">
              <span class="date-label">{{ t('carDocumentsForm.paymentDate') }}:</span>
              <span class="date-value">{{
                new Date(props.car.date_pay_freight).toLocaleDateString()
              }}</span>
            </div>
            <div v-if="!props.car.date_pay_freight" class="date-input-group">
              <label for="freight-date">{{ t('carDocumentsForm.date') }}:</label>
              <input
                type="date"
                id="freight-date"
                v-model="freightDate"
                class="date-input"
                :disabled="loading"
              />
            </div>
            <div class="button-group">
              <button
                class="action-btn freight-btn"
                @click="handleFreightPaid"
                :disabled="loading || !!props.car.date_pay_freight"
                :class="{ disabled: !!props.car.date_pay_freight }"
              >
                {{ loading ? t('carDocumentsForm.processing') : t('carDocumentsForm.freightPaid') }}
              </button>
              <button
                v-if="isAdmin && props.car.date_pay_freight"
                class="revert-btn"
                @click="handleRevertFreight"
                :disabled="loading"
                title="Revert freight payment status"
              >
                ↶
              </button>
            </div>
          </div>

          <!-- Supplier Documents Section -->
          <div class="document-section">
            <h4>{{ t('carDocumentsForm.supplierDocuments') }}</h4>
            <div class="status-info">
              <span class="status-label">{{ t('carDocumentsForm.status') }}:</span>
              <span
                :class="[
                  'status-value',
                  props.car.date_get_documents_from_supp ? 'received' : 'pending',
                ]"
              >
                {{
                  props.car.date_get_documents_from_supp
                    ? t('carDocumentsForm.received')
                    : t('carDocumentsForm.pending')
                }}
              </span>
            </div>
            <div v-if="props.car.date_get_documents_from_supp" class="date-info">
              <span class="date-label">{{ t('carDocumentsForm.receivedDate') }}:</span>
              <span class="date-value">{{
                new Date(props.car.date_get_documents_from_supp).toLocaleDateString()
              }}</span>
            </div>
            <div v-if="!props.car.date_get_documents_from_supp" class="date-input-group">
              <label for="supplier-docs-date">{{ t('carDocumentsForm.date') }}:</label>
              <input
                type="date"
                id="supplier-docs-date"
                v-model="supplierDocsDate"
                class="date-input"
                :disabled="loading"
              />
            </div>
            <div class="button-group">
              <button
                class="action-btn supplier-docs-btn"
                @click="handleDocsFromSupplier"
                :disabled="loading || !!props.car.date_get_documents_from_supp"
                :class="{ disabled: !!props.car.date_get_documents_from_supp }"
              >
                {{
                  loading
                    ? t('carDocumentsForm.processing')
                    : t('carDocumentsForm.weGetDocsFromSupplier')
                }}
              </button>
              <button
                v-if="isAdmin && props.car.date_get_documents_from_supp"
                class="revert-btn"
                @click="handleRevertSupplierDocs"
                :disabled="loading"
                title="Revert supplier documents received status"
              >
                ↶
              </button>
            </div>
          </div>

          <!-- Client Documents Section -->
          <div class="document-section">
            <h4>{{ t('carDocumentsForm.weSendClientDocuments') }}</h4>
            <div class="status-info">
              <span class="status-label">{{ t('carDocumentsForm.status') }}:</span>
              <span :class="['status-value', props.car.date_send_documents ? 'sent' : 'pending']">
                {{
                  props.car.date_send_documents
                    ? t('carDocumentsForm.sent')
                    : t('carDocumentsForm.pending')
                }}
              </span>
            </div>
            <div v-if="props.car.date_send_documents" class="date-info">
              <span class="date-label">{{ t('carDocumentsForm.sentDate') }}:</span>
              <span class="date-value">{{
                new Date(props.car.date_send_documents).toLocaleDateString()
              }}</span>
            </div>
            <div v-if="!props.car.date_send_documents" class="date-input-group">
              <label for="client-docs-date">{{ t('carDocumentsForm.date') }}:</label>
              <input
                type="date"
                id="client-docs-date"
                v-model="clientDocsDate"
                class="date-input"
                :disabled="loading"
              />
            </div>
            <div class="button-group">
              <button
                class="action-btn client-docs-btn"
                @click="handleDocsSentToClient"
                :disabled="loading || !!props.car.date_send_documents"
                :class="{ disabled: !!props.car.date_send_documents }"
              >
                {{
                  loading
                    ? t('carDocumentsForm.processing')
                    : t('carDocumentsForm.documentsSentToClient')
                }}
              </button>
              <button
                v-if="isAdmin && props.car.date_send_documents"
                class="revert-btn"
                @click="handleRevertClientDocs"
                :disabled="loading"
                title="Revert client documents sent status"
              >
                ↶
              </button>
            </div>
          </div>
        </div>

        <!-- Export License Section -->
        <div class="export-license-section">
          <h4>{{ t('carDocumentsForm.exportLicense') }}</h4>
          <div class="form-group">
            <label for="export-license">{{ t('carDocumentsForm.referenceNumber') }}:</label>
            <div class="input-group">
              <input
                type="text"
                id="export-license"
                v-model="exportLicenseRef"
                :placeholder="t('carDocumentsForm.enterExportLicenseReference')"
                class="input-field"
                :disabled="loading"
              />
              <button class="update-btn" @click="handleExportLicenseUpdate" :disabled="loading">
                {{ t('carDocumentsForm.update') }}
              </button>
            </div>
          </div>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
        <div v-if="success" class="success-message">
          {{ success }}
        </div>
      </div>

      <div class="modal-footer">
        <button class="close-btn-secondary" @click="closeModal" :disabled="loading">
          {{ t('carDocumentsForm.close') }}
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
  max-width: 800px;
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

.document-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.document-section {
  background-color: #f9fafb;
  padding: 16px;
  border-radius: 6px;
}

.document-section h4 {
  margin: 0 0 12px 0;
  color: #374151;
  font-size: 1.1rem;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-label,
.date-label {
  font-weight: 500;
  color: #6b7280;
}

.status-value {
  font-weight: 600;
}

.received,
.paid,
.sent {
  color: #059669;
}

.pending {
  color: #dc2626;
}

.date-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.date-value {
  color: #374151;
}

.date-input-group {
  margin-bottom: 12px;
}

.date-input-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
  color: #6b7280;
  font-size: 14px;
}

.date-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  background-color: white;
}

.date-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.date-input:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.action-btn {
  width: 100%;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.2s;
}

.get-bl-btn {
  background-color: #3b82f6;
  color: white;
}

.get-bl-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.freight-btn {
  background-color: #059669;
  color: white;
}

.freight-btn:hover:not(:disabled) {
  background-color: #047857;
}

.supplier-docs-btn {
  background-color: #8b5cf6;
  color: white;
}

.supplier-docs-btn:hover:not(:disabled) {
  background-color: #7c3aed;
}

.client-docs-btn {
  background-color: #f59e0b;
  color: white;
}

.client-docs-btn:hover:not(:disabled) {
  background-color: #d97706;
}

.action-btn.disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.button-group {
  display: flex;
  gap: 8px;
  align-items: center;
}

.button-group .action-btn {
  flex: 1;
}

.revert-btn {
  padding: 8px 12px;
  background-color: #dc2626;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
}

.revert-btn:hover:not(:disabled) {
  background-color: #b91c1c;
}

.revert-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.export-license-section {
  background-color: #f9fafb;
  padding: 16px;
  border-radius: 6px;
  margin-top: 20px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.input-group {
  display: flex;
  gap: 8px;
}

.input-field {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.update-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.update-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.error-message {
  color: #ef4444;
  margin: 16px 0;
  font-size: 14px;
}

.success-message {
  color: #10b981;
  margin: 16px 0;
  font-size: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.close-btn-secondary {
  padding: 8px 16px;
  background-color: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.close-btn-secondary:hover {
  background-color: #e5e7eb;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>
