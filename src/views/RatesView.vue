<template>
  <div class="rates-view">
    <div class="header">
      <h1>Exchange Rates Management</h1>
      <button @click="returnToDashboard" class="return-btn" :disabled="loading">
        Return to Dashboard
      </button>
    </div>

    <div class="actions">
      <button @click="showAddDialog = true" class="add-btn" :disabled="loading">
        Add New Rate
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">Loading rates...</div>

    <!-- Error State -->
    <div v-if="apiError" class="error-message">
      {{ apiError }}
    </div>

    <!-- Rates Table -->
    <div v-if="!loading && !apiError" class="table-container">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Rate</th>
            <th>Created On</th>
            <th>Created By</th>
            <th>Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rate in rates" :key="rate.id">
            <td>{{ rate.id }}</td>
            <td>{{ rate.rate }}</td>
            <td>{{ formatDate(rate.created_on) }}</td>
            <td>{{ rate.username }}</td>
            <td>{{ rate.notes }}</td>
            <td>
              <button
                v-if="canEdit(rate)"
                @click="editRate(rate)"
                class="edit-btn"
                :disabled="loading"
              >
                Edit
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add/Edit Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay" @click.self="showAddDialog = false">
      <div class="dialog">
        <h2>{{ editingRate ? 'Edit Rate' : 'Add New Rate' }}</h2>
        <form @submit.prevent="saveRate">
          <div class="form-group">
            <label>Rate:</label>
            <input
              type="number"
              v-model="rateForm.rate"
              step="0.01"
              required
              :disabled="loading"
              placeholder="Enter exchange rate"
            />
          </div>
          <div class="form-group">
            <label>Notes:</label>
            <textarea
              v-model="rateForm.notes"
              rows="3"
              :disabled="loading"
              placeholder="Add any relevant notes"
            ></textarea>
          </div>
          <div class="dialog-actions">
            <button
              type="button"
              @click="showAddDialog = false"
              class="cancel-btn"
              :disabled="loading"
            >
              Cancel
            </button>
            <button type="submit" class="save-btn" :disabled="loading">
              {{ loading ? 'Saving...' : editingRate ? 'Update' : 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'
import { useRouter } from 'vue-router'

const router = useRouter()
const { callApi, error: apiError, loading } = useApi()

const rates = ref([])
const showAddDialog = ref(false)
const editingRate = ref(null)
const currentUser = ref(null)

const rateForm = ref({
  rate: '',
  notes: '',
})

onMounted(async () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    currentUser.value = JSON.parse(userStr)
    await fetchRates()
  }
})

const fetchRates = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT r.*, u.username 
        FROM rates r 
        LEFT JOIN users u ON r.id_user = u.id 
        ORDER BY r.created_on DESC
      `,
    })

    if (result.success) {
      rates.value = result.data
    } else {
      console.error('Error fetching rates:', result.error)
    }
  } catch (error) {
    console.error('Error fetching rates:', error)
  }
}

const saveRate = async () => {
  try {
    if (editingRate.value) {
      // Update existing rate
      const result = await callApi({
        query: `
          UPDATE rates 
          SET rate = ?, notes = ?
          WHERE id = ? AND (id_user = ? OR ? = 1)
        `,
        params: [
          rateForm.value.rate,
          rateForm.value.notes,
          editingRate.value.id,
          currentUser.value.id,
          currentUser.value.role_id,
        ],
      })

      if (!result.success) {
        throw new Error(result.error || 'Failed to update rate')
      }
    } else {
      // Create new rate
      const result = await callApi({
        query: `
          INSERT INTO rates (rate, created_on, id_user, notes) 
          VALUES (?, NOW(), ?, ?)
        `,
        params: [rateForm.value.rate, currentUser.value.id, rateForm.value.notes],
      })

      if (!result.success) {
        throw new Error(result.error || 'Failed to create rate')
      }
    }

    await fetchRates()
    showAddDialog.value = false
    resetForm()
  } catch (error) {
    console.error('Error saving rate:', error)
  }
}

const editRate = (rate) => {
  editingRate.value = rate
  rateForm.value = {
    rate: rate.rate,
    notes: rate.notes,
  }
  showAddDialog.value = true
}

const resetForm = () => {
  editingRate.value = null
  rateForm.value = {
    rate: '',
    notes: '',
  }
}

const formatDate = (date) => {
  return new Date(date).toLocaleString()
}

const canEdit = (rate) => {
  if (!currentUser.value) return false
  if (currentUser.value.role_id === 1) return true
  return currentUser.value.id === rate.id_user
}

const returnToDashboard = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
.rates-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h1 {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: #1e293b;
  margin: 0;
}

.header h1:before {
  content: '\f3d1';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  color: #3b82f6;
}

.return-btn {
  background-color: #607d8b;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  position: relative;
  overflow: hidden;
}

.return-btn:before {
  content: '\f060';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
}

.return-btn:hover {
  background-color: #455a64;
  transform: translateY(-2px);
}

.return-btn:active {
  transform: translateY(0);
}

.actions {
  margin-bottom: 20px;
}

.add-btn {
  background-color: #4caf50;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

.add-btn:before {
  content: '\f067';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
}

.add-btn:hover {
  background-color: #43a047;
  transform: translateY(-2px);
}

.add-btn:active {
  transform: translateY(0);
}

.table-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin: 0;
}

th,
td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
  transition: all 0.2s ease;
}

th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #475569;
  position: sticky;
  top: 0;
  z-index: 10;
}

th:before {
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  margin-right: 0.5rem;
  color: #64748b;
}

th:nth-child(1):before {
  content: '\f292';
} /* ID */
th:nth-child(2):before {
  content: '\f3d1';
} /* Rate */
th:nth-child(3):before {
  content: '\f133';
} /* Created On */
th:nth-child(4):before {
  content: '\f007';
} /* Created By */
th:nth-child(5):before {
  content: '\f249';
} /* Notes */
th:nth-child(6):before {
  content: '\f7d9';
} /* Actions */

tr:hover td {
  background-color: #f8fafc;
}

.edit-btn {
  background-color: #2196f3;
  color: white;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.edit-btn:before {
  content: '\f044';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
}

.edit-btn:hover:not(:disabled) {
  background-color: #1e88e5;
  transform: translateY(-2px);
}

.edit-btn:active:not(:disabled) {
  transform: translateY(0);
}

.loading {
  text-align: center;
  padding: 3rem;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.loading:before {
  content: '\f110';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  font-size: 2rem;
  color: #3b82f6;
  animation: spin 1s linear infinite;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  animation: slideIn 0.3s ease;
}

.error-message:before {
  content: '\f071';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  backdrop-filter: blur(4px);
  animation: fadeIn 0.2s ease;
}

.dialog {
  background-color: white;
  padding: 1.5rem;
  border-radius: 12px;
  width: 400px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  animation: slideUp 0.3s ease;
}

.dialog h2 {
  margin: 0 0 1.5rem 0;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.dialog h2:before {
  content: '\f067';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  color: #3b82f6;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #475569;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.form-group input:focus,
.form-group textarea:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  outline: none;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.cancel-btn {
  background-color: #e2e8f0;
  color: #475569;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.cancel-btn:before {
  content: '\f00d';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
}

.cancel-btn:hover {
  background-color: #cbd5e1;
}

.save-btn {
  background-color: #4caf50;
  color: white;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.save-btn:before {
  content: '\f0c7';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
}

.save-btn:hover:not(:disabled) {
  background-color: #43a047;
  transform: translateY(-1px);
}

.save-btn:active:not(:disabled) {
  transform: translateY(0);
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes slideIn {
  from {
    transform: translateY(-10px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
</style>
