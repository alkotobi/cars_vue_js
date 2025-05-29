<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi } = useApi()
const banks = ref([])
const loading = ref(false)
const error = ref(null)
const showAddDialog = ref(false)
const editingBank = ref(null)
const isProcessing = ref(false)

// Form data
const formData = ref({
  company_name: '',
  bank_name: '',
  swift_code: '',
  bank_account: '',
  bank_address: '',
  notes: '',
})

const resetForm = () => {
  formData.value = {
    company_name: '',
    bank_name: '',
    swift_code: '',
    bank_account: '',
    bank_address: '',
    notes: '',
  }
}

const fetchBanks = async () => {
  loading.value = true
  error.value = null
  try {
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY company_name ASC',
    })
    if (result.success) {
      banks.value = result.data
    } else {
      error.value = 'Failed to fetch banks'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  resetForm()
  showAddDialog.value = true
  editingBank.value = null
}

const handleEdit = (bank) => {
  formData.value = { ...bank }
  showAddDialog.value = true
  editingBank.value = bank
}

const handleDelete = async (bank) => {
  if (!confirm('Are you sure you want to delete this bank?')) return

  loading.value = true
  error.value = null
  try {
    const result = await callApi({
      query: 'DELETE FROM banks WHERE id = ?',
      params: [bank.id],
    })
    if (result.success) {
      await fetchBanks()
    } else {
      error.value = 'Failed to delete bank'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (isProcessing.value) return
  isProcessing.value = true
  error.value = null

  try {
    let result
    if (editingBank.value) {
      // Update
      result = await callApi({
        query: `
          UPDATE banks 
          SET company_name = ?, bank_name = ?, swift_code = ?, 
              bank_account = ?, bank_address = ?, notes = ?
          WHERE id = ?
        `,
        params: [
          formData.value.company_name,
          formData.value.bank_name,
          formData.value.swift_code,
          formData.value.bank_account,
          formData.value.bank_address,
          formData.value.notes,
          editingBank.value.id,
        ],
      })
    } else {
      // Insert
      result = await callApi({
        query: `
          INSERT INTO banks 
          (company_name, bank_name, swift_code, bank_account, bank_address, notes)
          VALUES (?, ?, ?, ?, ?, ?)
        `,
        params: [
          formData.value.company_name,
          formData.value.bank_name,
          formData.value.swift_code,
          formData.value.bank_account,
          formData.value.bank_address,
          formData.value.notes,
        ],
      })
    }

    if (result.success) {
      showAddDialog.value = false
      await fetchBanks()
    } else {
      error.value = 'Failed to save bank'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    isProcessing.value = false
  }
}

onMounted(() => {
  fetchBanks()
})
</script>

<template>
  <div class="banks-table">
    <div class="header">
      <h2>Bank Accounts</h2>
      <button @click="handleAdd" class="add-btn" :disabled="loading">
        <i class="fas fa-plus"></i>
        Add Bank
      </button>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      Loading banks...
    </div>

    <!-- Banks Table -->
    <div v-else class="table-container">
      <table v-if="banks.length > 0">
        <thead>
          <tr>
            <th>Company Name</th>
            <th>Bank Name</th>
            <th>Swift Code</th>
            <th>Account Number</th>
            <th>Bank Address</th>
            <th>Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="bank in banks" :key="bank.id">
            <td>{{ bank.company_name }}</td>
            <td>{{ bank.bank_name }}</td>
            <td>{{ bank.swift_code }}</td>
            <td>{{ bank.bank_account }}</td>
            <td>{{ bank.bank_address }}</td>
            <td>{{ bank.notes }}</td>
            <td class="actions">
              <button @click="handleEdit(bank)" class="edit-btn">
                <i class="fas fa-edit"></i>
              </button>
              <button @click="handleDelete(bank)" class="delete-btn">
                <i class="fas fa-trash"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No banks found. Click "Add Bank" to create one.</div>
    </div>

    <!-- Add/Edit Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{{ editingBank ? 'Edit Bank' : 'Add Bank' }}</h3>
          <button @click="showAddDialog = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="handleSubmit" class="bank-form">
          <div class="form-group">
            <label for="company_name">Company Name</label>
            <input
              id="company_name"
              v-model="formData.company_name"
              type="text"
              required
              placeholder="Enter company name"
            />
          </div>

          <div class="form-group">
            <label for="bank_name">Bank Name</label>
            <input
              id="bank_name"
              v-model="formData.bank_name"
              type="text"
              required
              placeholder="Enter bank name"
            />
          </div>

          <div class="form-group">
            <label for="swift_code">Swift Code</label>
            <input
              id="swift_code"
              v-model="formData.swift_code"
              type="text"
              required
              placeholder="Enter SWIFT code"
            />
          </div>

          <div class="form-group">
            <label for="bank_account">Account Number</label>
            <input
              id="bank_account"
              v-model="formData.bank_account"
              type="text"
              required
              placeholder="Enter account number"
            />
          </div>

          <div class="form-group">
            <label for="bank_address">Bank Address</label>
            <input
              id="bank_address"
              v-model="formData.bank_address"
              type="text"
              required
              placeholder="Enter bank address"
            />
          </div>

          <div class="form-group">
            <label for="notes">Notes</label>
            <textarea
              id="notes"
              v-model="formData.notes"
              placeholder="Enter additional notes"
              rows="3"
            ></textarea>
          </div>

          <div class="form-actions">
            <button
              type="button"
              @click="showAddDialog = false"
              class="cancel-btn"
              :disabled="isProcessing"
            >
              Cancel
            </button>
            <button type="submit" class="save-btn" :disabled="isProcessing">
              <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
              {{ isProcessing ? 'Saving...' : 'Save Bank' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.banks-table {
  width: 100%;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #1f2937;
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.add-btn:hover {
  background-color: #059669;
}

.table-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

th {
  background-color: #f9fafb;
  font-weight: 600;
  color: #4b5563;
}

.actions {
  display: flex;
  gap: 8px;
}

.edit-btn,
.delete-btn {
  padding: 6px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 50;
}

.dialog {
  background-color: white;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  font-size: 1.25rem;
  padding: 4px;
}

.close-btn:hover {
  color: #1f2937;
}

.bank-form {
  padding: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 4px;
  color: #4b5563;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  ring: 2px solid #3b82f6;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn,
.save-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.cancel-btn {
  background-color: #e5e7eb;
  color: #4b5563;
}

.cancel-btn:hover {
  background-color: #d1d5db;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
  display: flex;
  align-items: center;
  gap: 8px;
}

.save-btn:hover {
  background-color: #2563eb;
}

.save-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.error-message {
  padding: 12px;
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  border-radius: 6px;
  color: #b91c1c;
  margin-bottom: 16px;
}

.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  padding: 24px;
  justify-content: center;
}

.empty-state {
  text-align: center;
  padding: 24px;
  color: #6b7280;
}

@media (max-width: 768px) {
  .dialog {
    width: 95%;
    margin: 16px;
  }

  .form-actions {
    flex-direction: column;
  }

  .form-actions button {
    width: 100%;
  }
}
</style>
