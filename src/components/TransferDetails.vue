<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '../composables/useApi'

const props = defineProps({
  transferId: {
    type: Number,
    required: true,
  },
  isVisible: {
    type: Boolean,
    default: false,
  },
  readOnly: {
    type: Boolean,
    default: false,
  },
})

console.log('TransferDetails component props:', props)

const emit = defineEmits(['close', 'refresh'])

const { callApi } = useApi()
const details = ref([])
const isLoading = ref(false)
const showAddDialog = ref(false)
const processingId = ref(null)
const error = ref(null)
const isProcessing = ref(false)

const detailForm = ref({
  amount: '',
  client_name: '',
  client_mobile: '',
  rate: '',
  description: '',
  notes: '',
})

onMounted(() => {
  console.log('TransferDetails mounted, transferId:', props.transferId)
  if (props.transferId) {
    fetchDetails()
  }
})

watch(
  () => props.isVisible,
  (newVal) => {
    console.log('isVisible changed:', newVal)
    if (newVal) {
      console.log('Calling fetchDetails due to visibility change')
      fetchDetails()
    }
  },
)

watch(
  () => props.transferId,
  (newId) => {
    console.log('transferId changed:', newId)
    if (newId) {
      console.log('Calling fetchDetails due to ID change')
      fetchDetails()
    }
  },
)

const fetchDetails = async () => {
  console.log('fetchDetails called with ID:', props.transferId)
  if (!props.transferId) {
    console.log('No transfer ID provided')
    return
  }

  isLoading.value = true
  error.value = null

  try {
    const query = `
      SELECT 
        id,
        id_transfer,
        amount,
        date,
        client_name,
        client_mobile,
        rate,
        description,
        notes
      FROM transfer_details 
      WHERE id_transfer = ?
      ORDER BY date DESC
    `
    console.log('Executing query:', query.replace(/\s+/g, ' ').trim())
    console.log('With params:', [props.transferId])

    const result = await callApi({
      query,
      params: [props.transferId],
    })

    console.log('Raw API response:', result)

    if (result.success) {
      if (!result.data || result.data.length === 0) {
        console.log('No details found for transfer ID:', props.transferId)
        error.value = 'No details found for this transfer. Click "Add Detail" to add some.'
      } else {
        console.log('Found details:', result.data)
      }
      details.value = result.data || []
    } else {
      console.error('API call failed:', result.error)
      error.value = 'Failed to load details: ' + (result.error || 'Unknown error')
    }
  } catch (err) {
    console.error('Error fetching details:', err)
    error.value = 'Failed to load details: ' + (err.message || 'Unknown error')
  } finally {
    isLoading.value = false
  }
}

const addDetail = async () => {
  if (isProcessing.value) return
  isProcessing.value = true

  try {
    const result = await callApi({
      query: `
        INSERT INTO transfer_details 
        (id_transfer, amount, date, client_name, client_mobile, rate, description, notes)
        VALUES (?, ?, NOW(), ?, ?, ?, ?, ?)
      `,
      params: [
        props.transferId,
        detailForm.value.amount,
        detailForm.value.client_name,
        detailForm.value.client_mobile,
        detailForm.value.rate,
        detailForm.value.description,
        detailForm.value.notes,
      ],
    })

    if (result.success) {
      showAddDialog.value = false
      detailForm.value = {
        amount: '',
        client_name: '',
        client_mobile: '',
        rate: '',
        description: '',
        notes: '',
      }
      await fetchDetails()
      emit('refresh')
    }
  } catch (err) {
    console.error('Error adding detail:', err)
    error.value = 'Failed to add detail'
  } finally {
    isProcessing.value = false
  }
}

const deleteDetail = async (id) => {
  if (processingId.value) return
  if (!confirm('Are you sure you want to delete this detail?')) return

  processingId.value = id
  try {
    const result = await callApi({
      query: 'DELETE FROM transfer_details WHERE id = ?',
      params: [id],
    })

    if (result.success) {
      await fetchDetails()
      emit('refresh')
    }
  } catch (err) {
    console.error('Error deleting detail:', err)
    error.value = 'Failed to delete detail'
  } finally {
    processingId.value = null
  }
}

const editDetail = ref(null)
const showEditDialog = ref(false)

const openEditDialog = (detail) => {
  editDetail.value = { ...detail }
  showEditDialog.value = true
}

const updateDetail = async () => {
  if (!editDetail.value || isProcessing.value) return
  isProcessing.value = true

  try {
    const result = await callApi({
      query: `
        UPDATE transfer_details 
        SET amount = ?,
            client_name = ?,
            client_mobile = ?,
            rate = ?,
            description = ?,
            notes = ?
        WHERE id = ?
      `,
      params: [
        editDetail.value.amount,
        editDetail.value.client_name,
        editDetail.value.client_mobile,
        editDetail.value.rate,
        editDetail.value.description,
        editDetail.value.notes,
        editDetail.value.id,
      ],
    })

    if (result.success) {
      showEditDialog.value = false
      editDetail.value = null
      await fetchDetails()
      emit('refresh')
    }
  } catch (err) {
    console.error('Error updating detail:', err)
    error.value = 'Failed to update detail'
  } finally {
    isProcessing.value = false
  }
}

const handleClose = () => {
  emit('close')
  emit('refresh')
}
</script>

<template>
  <div v-if="isVisible" class="transfer-details-modal">
    <div class="modal-content">
      <div class="modal-header">
        <h2><i class="fas fa-list-ul"></i> Transfer Details</h2>
        <button @click="handleClose" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div v-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div class="actions">
        <button v-if="!props.readOnly" @click="showAddDialog = true" class="btn add-btn">
          <i class="fas fa-plus"></i> Add Detail
        </button>
      </div>

      <div v-if="isLoading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading details...</span>
      </div>

      <div class="details-table" v-else>
        <table>
          <thead>
            <tr>
              <th>Date</th>
              <th>Amount</th>
              <th>Rate</th>
              <th>Client Name</th>
              <th>Mobile</th>
              <th>Description</th>
              <th>Notes</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="detail in details" :key="detail.id">
              <td>{{ new Date(detail.date).toLocaleString() }}</td>
              <td>{{ detail.amount }}</td>
              <td>{{ detail.rate }}</td>
              <td>{{ detail.client_name }}</td>
              <td>{{ detail.client_mobile }}</td>
              <td>{{ detail.description }}</td>
              <td>{{ detail.notes }}</td>
              <td>
                <div class="action-buttons">
                  <button
                    v-if="!props.readOnly"
                    @click="openEditDialog(detail)"
                    class="btn edit-btn"
                    :disabled="processingId === detail.id"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button
                    v-if="!props.readOnly"
                    @click="deleteDetail(detail.id)"
                    class="btn delete-btn"
                    :disabled="processingId === detail.id"
                  >
                    <i class="fas fa-trash"></i>
                    <i v-if="processingId === detail.id" class="fas fa-spinner fa-spin"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Add Dialog -->
      <div v-if="showAddDialog" class="dialog-overlay">
        <div class="dialog">
          <h3><i class="fas fa-plus-circle"></i> Add Detail</h3>
          <form @submit.prevent="addDetail">
            <div v-if="isProcessing" class="dialog-loading">
              <i class="fas fa-spinner fa-spin dialog-spinner"></i>
            </div>
            <div class="form-group">
              <label><i class="fas fa-dollar-sign"></i> Amount:</label>
              <input
                type="number"
                v-model="detailForm.amount"
                step="0.01"
                required
                class="input-field"
              />
            </div>
            <div class="form-group">
              <label><i class="fas fa-user"></i> Client Name:</label>
              <input type="text" v-model="detailForm.client_name" required class="input-field" />
            </div>
            <div class="form-group">
              <label><i class="fas fa-mobile-alt"></i> Client Mobile:</label>
              <input type="text" v-model="detailForm.client_mobile" class="input-field" />
            </div>
            <div class="form-group">
              <label><i class="fas fa-exchange-alt"></i> Rate:</label>
              <input
                type="number"
                v-model="detailForm.rate"
                step="0.01"
                required
                class="input-field"
              />
            </div>
            <div class="form-group">
              <label><i class="fas fa-info-circle"></i> Description:</label>
              <input type="text" v-model="detailForm.description" class="input-field" />
            </div>
            <div class="form-group">
              <label><i class="fas fa-sticky-note"></i> Notes:</label>
              <textarea v-model="detailForm.notes" class="input-field"></textarea>
            </div>
            <div class="dialog-actions">
              <button type="submit" class="btn save-btn" :disabled="isProcessing">
                <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
                <i v-else class="fas fa-save"></i>
                Save
              </button>
              <button
                type="button"
                @click="showAddDialog = false"
                class="btn cancel-btn"
                :disabled="isProcessing"
              >
                <i class="fas fa-times"></i> Cancel
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Edit Dialog -->
      <div v-if="showEditDialog" class="dialog-overlay">
        <div class="dialog">
          <h3><i class="fas fa-edit"></i> Edit Detail</h3>
          <form @submit.prevent="updateDetail">
            <div v-if="isProcessing" class="dialog-loading">
              <i class="fas fa-spinner fa-spin dialog-spinner"></i>
            </div>
            <div class="form-group">
              <label><i class="fas fa-dollar-sign"></i> Amount:</label>
              <input
                type="number"
                v-model="editDetail.amount"
                step="0.01"
                required
                class="input-field"
              />
            </div>
            <div class="form-group">
              <label><i class="fas fa-user"></i> Client Name:</label>
              <input type="text" v-model="editDetail.client_name" required class="input-field" />
            </div>
            <div class="form-group">
              <label><i class="fas fa-mobile-alt"></i> Client Mobile:</label>
              <input type="text" v-model="editDetail.client_mobile" class="input-field" />
            </div>
            <div class="form-group">
              <label><i class="fas fa-exchange-alt"></i> Rate:</label>
              <input
                type="number"
                v-model="editDetail.rate"
                step="0.01"
                required
                class="input-field"
              />
            </div>
            <div class="form-group">
              <label><i class="fas fa-info-circle"></i> Description:</label>
              <input type="text" v-model="editDetail.description" class="input-field" />
            </div>
            <div class="form-group">
              <label><i class="fas fa-sticky-note"></i> Notes:</label>
              <textarea v-model="editDetail.notes" class="input-field"></textarea>
            </div>
            <div class="dialog-actions">
              <button type="submit" class="btn save-btn" :disabled="isProcessing">
                <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
                <i v-else class="fas fa-save"></i>
                Save
              </button>
              <button
                type="button"
                @click="showEditDialog = false"
                class="btn cancel-btn"
                :disabled="isProcessing"
              >
                <i class="fas fa-times"></i> Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.transfer-details-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 24px;
  width: 90%;
  max-width: 1200px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h2 {
  margin: 0;
  color: #1f2937;
  font-size: 1.5em;
  display: flex;
  align-items: center;
  gap: 8px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.2em;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
}

.close-btn:hover {
  color: #1f2937;
}

.actions {
  margin-bottom: 20px;
}

.details-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  margin-bottom: 20px;
}

th {
  background: #f8fafc;
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #1f2937;
  border-bottom: 2px solid #e5e7eb;
}

td {
  padding: 12px;
  border-bottom: 1px solid #f1f5f9;
  color: #4b5563;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.add-btn {
  background: #10b981;
  color: white;
}

.edit-btn {
  background: #3b82f6;
  color: white;
  padding: 6px 10px;
}

.delete-btn {
  background: #ef4444;
  color: white;
  padding: 6px 10px;
}

.save-btn {
  background: #10b981;
  color: white;
}

.cancel-btn {
  background: #6b7280;
  color: white;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1100;
}

.dialog {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dialog h3 {
  margin: 0 0 20px;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #4b5563;
  font-weight: 500;
}

.input-field {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  font-size: 0.95em;
  transition: all 0.2s ease;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

textarea.input-field {
  min-height: 100px;
  resize: vertical;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 24px;
}

.loading-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px;
  color: #6b7280;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fee2e2;
  border-radius: 6px;
  color: #dc2626;
  margin-bottom: 16px;
}

.dialog-loading {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  border-radius: 8px;
}

.dialog form {
  position: relative;
}

.dialog-spinner {
  color: #3b82f6;
  font-size: 2em;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  pointer-events: none;
}
</style>
