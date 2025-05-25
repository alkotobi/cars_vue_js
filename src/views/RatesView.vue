<template>
  <div class="rates-view">
    <div class="header">
      <h1>Exchange Rates Management</h1>
      <button @click="returnToDashboard" class="return-btn">
        Return to Dashboard
      </button>
    </div>
    
    <div class="actions">
      <button @click="showAddDialog = true" class="add-btn">
        Add New Rate
      </button>
    </div>

    <div v-if="loading" class="loading">
      Loading...
    </div>

    <div v-if="apiError" class="error-message">
      {{ apiError }}
    </div>

    <!-- Rates Table -->
    <div v-if="!loading" class="table-container">
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
              >
                Edit
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add/Edit Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
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
            >
          </div>
          <div class="form-group">
            <label>Notes:</label>
            <textarea 
              v-model="rateForm.notes"
              rows="3"
            ></textarea>
          </div>
          <div class="dialog-actions">
            <button type="button" @click="showAddDialog = false" class="cancel-btn">
              Cancel
            </button>
            <button type="submit" class="save-btn" :disabled="loading">
              {{ editingRate ? 'Update' : 'Save' }}
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
  notes: ''
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
      `
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
          currentUser.value.role_id
        ]
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
        params: [
          rateForm.value.rate,
          currentUser.value.id,
          rateForm.value.notes
        ]
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
    notes: rate.notes
  }
  showAddDialog.value = true
}

const resetForm = () => {
  editingRate.value = null
  rateForm.value = {
    rate: '',
    notes: ''
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

.return-btn {
  background-color: #607d8b;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s, transform 0.2s;
}

.return-btn:hover {
  background-color: #455a64;
  transform: translateY(-2px);
}

.actions {
  margin-bottom: 20px;
}

.add-btn {
  background-color: #4CAF50;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.table-container {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

th {
  background-color: #f5f5f5;
  font-weight: 600;
}

.edit-btn {
  background-color: #2196F3;
  color: white;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
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
}

.dialog {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  width: 400px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.cancel-btn {
  background-color: #9e9e9e;
  color: white;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.save-btn {
  background-color: #4CAF50;
  color: white;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.loading {
  text-align: center;
  padding: 20px;
  font-size: 1.2em;
  color: #666;
}

.error-message {
  background-color: #ffebee;
  color: #c62828;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 20px;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style> 