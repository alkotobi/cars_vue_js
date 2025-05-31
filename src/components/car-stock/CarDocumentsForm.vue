<script setup>
import { ref, defineProps, defineEmits, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { ElMessageBox } from 'element-plus'

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

const emit = defineEmits(['close', 'save', 'documents-sent'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(null)

// Form field for export license
const exportLicenseRef = ref(props.car.export_lisence_ref || '')

// Handle BL received
const handleGetBL = async () => {
  if (props.car.date_get_bl) return

  try {
    await ElMessageBox.confirm(
      'Are you sure you want to mark the BL as received?',
      'Confirm Action',
      {
        confirmButtonText: 'Yes',
        cancelButtonText: 'No',
        type: 'warning',
      },
    )

    loading.value = true
    error.value = null
    success.value = null

    const currentDate = new Date().toISOString().split('T')[0]
    const result = await callApi({
      query: 'UPDATE cars_stock SET date_get_bl = ? WHERE id = ?',
      params: [currentDate, props.car.id],
    })

    if (result.success) {
      success.value = 'BL received date updated'
      const updatedCar = {
        ...props.car,
        date_get_bl: currentDate,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || 'Failed to update BL received date')
    }
  } catch (err) {
    if (err !== 'cancel') {
      error.value = err.message || 'An error occurred'
    }
  } finally {
    loading.value = false
  }
}

// Handle freight payment
const handleFreightPaid = async () => {
  if (props.car.date_pay_freight) return

  try {
    await ElMessageBox.confirm(
      'Are you sure you want to mark the freight as paid?',
      'Confirm Action',
      {
        confirmButtonText: 'Yes',
        cancelButtonText: 'No',
        type: 'warning',
      },
    )

    loading.value = true
    error.value = null
    success.value = null

    const currentDate = new Date().toISOString().split('T')[0]
    const result = await callApi({
      query: 'UPDATE cars_stock SET date_pay_freight = ? WHERE id = ?',
      params: [currentDate, props.car.id],
    })

    if (result.success) {
      success.value = 'Freight payment date updated'
      const updatedCar = {
        ...props.car,
        date_pay_freight: currentDate,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || 'Failed to update freight payment date')
    }
  } catch (err) {
    if (err !== 'cancel') {
      error.value = err.message || 'An error occurred'
    }
  } finally {
    loading.value = false
  }
}

// Handle documents received from supplier
const handleDocsFromSupplier = async () => {
  if (props.car.date_get_documents_from_supp) return

  try {
    await ElMessageBox.confirm(
      'Are you sure you want to mark the documents as received from supplier?',
      'Confirm Action',
      {
        confirmButtonText: 'Yes',
        cancelButtonText: 'No',
        type: 'warning',
      },
    )

    loading.value = true
    error.value = null
    success.value = null

    const currentDate = new Date().toISOString().split('T')[0]
    const result = await callApi({
      query: 'UPDATE cars_stock SET date_get_documents_from_supp = ? WHERE id = ?',
      params: [currentDate, props.car.id],
    })

    if (result.success) {
      success.value = 'Documents from supplier date updated'
      const updatedCar = {
        ...props.car,
        date_get_documents_from_supp: currentDate,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || 'Failed to update documents from supplier date')
    }
  } catch (err) {
    if (err !== 'cancel') {
      error.value = err.message || 'An error occurred'
    }
  } finally {
    loading.value = false
  }
}

// Handle documents sent to client
const handleDocsSentToClient = async () => {
  if (props.car.date_send_documents) return

  try {
    await ElMessageBox.confirm(
      'Are you sure you want to mark the documents as sent to client?',
      'Confirm Action',
      {
        confirmButtonText: 'Yes',
        cancelButtonText: 'No',
        type: 'warning',
      },
    )

    loading.value = true
    error.value = null
    success.value = null

    const currentDate = new Date().toISOString().split('T')[0]
    const result = await callApi({
      query: 'UPDATE cars_stock SET date_send_documents = ? WHERE id = ?',
      params: [currentDate, props.car.id],
    })

    if (result.success) {
      success.value = 'Documents sent to client date updated'
      const updatedCar = {
        ...props.car,
        date_send_documents: currentDate,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
      emit('documents-sent', updatedCar)
      setTimeout(() => {
        emit('close')
      }, 1500)
    } else {
      throw new Error(result.error || 'Failed to update documents sent date')
    }
  } catch (err) {
    if (err !== 'cancel') {
      error.value = err.message || 'An error occurred'
    }
  } finally {
    loading.value = false
  }
}

// Handle export license update
const handleExportLicenseUpdate = async () => {
  try {
    await ElMessageBox.confirm(
      'Are you sure you want to update the export license reference?',
      'Confirm Action',
      {
        confirmButtonText: 'Yes',
        cancelButtonText: 'No',
        type: 'warning',
      },
    )

    loading.value = true
    error.value = null
    success.value = null

    const result = await callApi({
      query: 'UPDATE cars_stock SET export_lisence_ref = ? WHERE id = ?',
      params: [exportLicenseRef.value, props.car.id],
    })

    if (result.success) {
      success.value = 'Export license reference updated'
      const updatedCar = {
        ...props.car,
        export_lisence_ref: exportLicenseRef.value,
      }
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || 'Failed to update export license reference')
    }
  } catch (err) {
    if (err !== 'cancel') {
      error.value = err.message || 'An error occurred'
    }
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  exportLicenseRef.value = props.car.export_lisence_ref || ''
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Car Documents Management</h3>
        <button class="close-btn" @click="closeModal">&times;</button>
      </div>

      <div class="modal-body">
        <!-- Document Status Grid -->
        <div class="document-grid">
          <!-- BL Section -->
          <div class="document-section">
            <h4>Bill of Lading (BL)</h4>
            <div class="status-info">
              <span class="status-label">Status:</span>
              <span :class="['status-value', props.car.date_get_bl ? 'received' : 'pending']">
                {{ props.car.date_get_bl ? 'Received' : 'Pending' }}
              </span>
            </div>
            <div v-if="props.car.date_get_bl" class="date-info">
              <span class="date-label">Received Date:</span>
              <span class="date-value">{{
                new Date(props.car.date_get_bl).toLocaleDateString()
              }}</span>
            </div>
            <button
              class="action-btn get-bl-btn"
              @click="handleGetBL"
              :disabled="loading || !!props.car.date_get_bl"
              :class="{ disabled: !!props.car.date_get_bl }"
            >
              {{ loading ? 'Processing...' : 'We Get BL' }}
            </button>
          </div>

          <!-- Freight Payment Section -->
          <div class="document-section">
            <h4>Freight Payment</h4>
            <div class="status-info">
              <span class="status-label">Status:</span>
              <span :class="['status-value', props.car.date_pay_freight ? 'paid' : 'pending']">
                {{ props.car.date_pay_freight ? 'Paid' : 'Pending' }}
              </span>
            </div>
            <div v-if="props.car.date_pay_freight" class="date-info">
              <span class="date-label">Payment Date:</span>
              <span class="date-value">{{
                new Date(props.car.date_pay_freight).toLocaleDateString()
              }}</span>
            </div>
            <button
              class="action-btn freight-btn"
              @click="handleFreightPaid"
              :disabled="loading || !!props.car.date_pay_freight"
              :class="{ disabled: !!props.car.date_pay_freight }"
            >
              {{ loading ? 'Processing...' : 'Freight Paid' }}
            </button>
          </div>

          <!-- Supplier Documents Section -->
          <div class="document-section">
            <h4>Supplier Documents</h4>
            <div class="status-info">
              <span class="status-label">Status:</span>
              <span
                :class="[
                  'status-value',
                  props.car.date_get_documents_from_supp ? 'received' : 'pending',
                ]"
              >
                {{ props.car.date_get_documents_from_supp ? 'Received' : 'Pending' }}
              </span>
            </div>
            <div v-if="props.car.date_get_documents_from_supp" class="date-info">
              <span class="date-label">Received Date:</span>
              <span class="date-value">{{
                new Date(props.car.date_get_documents_from_supp).toLocaleDateString()
              }}</span>
            </div>
            <button
              class="action-btn supplier-docs-btn"
              @click="handleDocsFromSupplier"
              :disabled="loading || !!props.car.date_get_documents_from_supp"
              :class="{ disabled: !!props.car.date_get_documents_from_supp }"
            >
              {{ loading ? 'Processing...' : 'We Get Docs from Supplier' }}
            </button>
          </div>

          <!-- Client Documents Section -->
          <div class="document-section">
            <h4>We Send Client Documents</h4>
            <div class="status-info">
              <span class="status-label">Status:</span>
              <span :class="['status-value', props.car.date_send_documents ? 'sent' : 'pending']">
                {{ props.car.date_send_documents ? 'Sent' : 'Pending' }}
              </span>
            </div>
            <div v-if="props.car.date_send_documents" class="date-info">
              <span class="date-label">Sent Date:</span>
              <span class="date-value">{{
                new Date(props.car.date_send_documents).toLocaleDateString()
              }}</span>
            </div>
            <button
              class="action-btn client-docs-btn"
              @click="handleDocsSentToClient"
              :disabled="loading || !!props.car.date_send_documents"
              :class="{ disabled: !!props.car.date_send_documents }"
            >
              {{ loading ? 'Processing...' : 'Documents Sent to Client' }}
            </button>
          </div>
        </div>

        <!-- Export License Section -->
        <div class="export-license-section">
          <h4>Export License</h4>
          <div class="form-group">
            <label for="export-license">Reference Number:</label>
            <div class="input-group">
              <input
                type="text"
                id="export-license"
                v-model="exportLicenseRef"
                placeholder="Enter export license reference"
                class="input-field"
                :disabled="loading"
              />
              <button class="update-btn" @click="handleExportLicenseUpdate" :disabled="loading">
                Update
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
        <button class="close-btn-secondary" @click="closeModal" :disabled="loading">Close</button>
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
